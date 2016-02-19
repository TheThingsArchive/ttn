// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package components

import (
	"github.com/TheThingsNetwork/ttn/core"
	. "github.com/TheThingsNetwork/ttn/core/errors"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	"github.com/TheThingsNetwork/ttn/utils/stats"
	"github.com/apex/log"
	"github.com/brocaar/lorawan"
)

// Broker type materializes the logic part handled by a broker
type Broker struct {
	ctx log.Interface // Just a logger
	db  BrokerStorage // Reference to the internal broker storage
}

// NewBroker constructs a new broker from a given storage
func NewBroker(db BrokerStorage, ctx log.Interface) *Broker {
	return &Broker{
		ctx: ctx,
		db:  db,
	}
}

// HandleUp implements the core.Component interface
func (b *Broker) HandleUp(p core.Packet, an core.AckNacker, adapter core.Adapter) error {
	stats.MarkMeter("broker.uplink.in")
	b.ctx.Debug("Handling uplink packet")

	// 1. Lookup for entries for the associated device
	devAddr, err := p.DevAddr()
	if err != nil {
		stats.MarkMeter("broker.uplink.invalid")
		b.ctx.Warn("Uplink Invalid")
		an.Nack()
		return errors.New(ErrInvalidStructure, err)
	}
	ctx := b.ctx.WithField("devAddr", devAddr)
	entries, err := b.db.Lookup(devAddr)
	if err != nil {
		switch err.(errors.Failure).Nature {
		case ErrWrongBehavior:
			stats.MarkMeter("broker.uplink.device_not_registered")
			ctx.Warn("Uplink device not found")
			return an.Nack()
		default:
			b.ctx.Warn("Database lookup failed")
			an.Nack()
			return errors.New(ErrFailedOperation, err)
		}
	}
	stats.UpdateHistogram("broker.handlers_per_dev_addr", int64(len(entries)))

	// 2. Several handler might be associated to the same device, we distinguish them using MIC
	// check. Only one should verify the MIC check.
	var handler *core.Recipient
	for _, entry := range entries {
		ok, err := p.Payload.ValidateMIC(entry.NwkSKey)
		if err != nil {
			continue
		}
		if ok {
			handler = &core.Recipient{
				Id:      entry.Id,
				Address: entry.Url,
			}
			stats.MarkMeter("broker.uplink.match_handler")
			ctx.WithField("handler", handler).Debug("Associated device with handler")
			break
		}
	}
	if handler == nil {
		stats.MarkMeter("broker.uplink.no_match_handler")
		ctx.Warn("Could not find handler for device")
		return an.Nack()
	}

	// 3. If one was found, we forward the packet and wait for the response
	response, err := adapter.Send(p, *handler)
	if err != nil {
		stats.MarkMeter("broker.uplink.bad_handler_response")
		an.Nack()
		return errors.New(ErrFailedOperation, err)
	}

	stats.MarkMeter("broker.uplink.ok")
	return an.Ack(&response)
}

// HandleDown implements the core.Component interface. Not implemented yet
func (b *Broker) HandleDown(p core.Packet, an core.AckNacker, a core.Adapter) error {
	return errors.New(ErrNotSupported, "HandleDown not supported on broker")
}

// Register implements the core.Component interface
func (b *Broker) Register(r core.Registration, an core.AckNacker) error {
	stats.MarkMeter("broker.registration.in")
	b.ctx.Debug("Handling registration")

	id, okId := r.Recipient.Id.(string)
	url, okUrl := r.Recipient.Address.(string)
	nwkSKey, okNwkSKey := r.Options.(lorawan.AES128Key)

	ctx := b.ctx.WithField("devAddr", r.DevAddr)

	if !(okId && okUrl && okNwkSKey) {
		stats.MarkMeter("broker.registration.invalid")
		ctx.Warn("Invalid Registration")
		an.Nack()
		return errors.New(ErrInvalidStructure, "Invalid registration params")
	}

	entry := brokerEntry{Id: id, Url: url, NwkSKey: nwkSKey}
	if err := b.db.Store(r.DevAddr, entry); err != nil {
		stats.MarkMeter("broker.registration.failed")
		ctx.WithError(err).Error("Failed Registration")
		an.Nack()
		return errors.New(ErrFailedOperation, err)
	}

	stats.MarkMeter("broker.registration.ok")
	ctx.Debug("Successful Registration")
	return an.Ack(nil)
}

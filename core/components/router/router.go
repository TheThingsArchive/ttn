// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package router

import (
	. "github.com/TheThingsNetwork/ttn/core"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	"github.com/TheThingsNetwork/ttn/utils/stats"
	"github.com/apex/log"
)

type component struct {
	Storage
	ctx log.Interface
}

// New constructs a new router
func New(db Storage, ctx log.Interface) Component {
	return component{Storage: db, ctx: ctx}
}

// Register implements the core.Component interface
func (r component) Register(reg Registration, an AckNacker) (err error) {
	defer ensureAckNack(an, nil, &err)
	stats.MarkMeter("router.registration.in")
	r.ctx.Debug("Handling registration")

	if err := r.Store(reg); err != nil {
		return errors.New(errors.Operational, err)
	}
	return nil
}

// HandleUp implements the core.Component interface
func (r component) HandleUp(data []byte, an AckNacker, up Adapter) (err error) {
	// Make sure we don't forget the AckNacker
	var ack Packet
	defer ensureAckNack(an, &ack, &err)

	// Get some logs / analytics
	stats.MarkMeter("router.uplink.in")
	r.ctx.Debug("Handling uplink packet")

	// Extract the given packet
	itf, err := UnmarshalPacket(data)
	if err != nil {
		stats.MarkMeter("router.uplink.invalid")
		r.ctx.Warn("Uplink Invalid")
		return errors.New(errors.Structural, err)
	}

	switch itf.(type) {
	case RPacket:
		packet := itf.(RPacket)

		// Lookup for an existing broker
		// NOTE We are still assuming only one broker associated to one device address.
		// We should find a mechanism to make sure that the broker in database is really
		// associated to the device to avoid trouble during overlaping.
		// Keeping track of the last FCnt maybe ? Having an overlap on the frame counter + the
		// device address might be less likely.
		entry, err := r.Lookup(packet.DevEUI())
		if err != nil && err.(errors.Failure).Nature != errors.Behavioural {
			r.ctx.Warn("Database lookup failed")
			return errors.New(errors.Operational, err)
		}

		var recipient Recipient
		if err == nil {
			rawRecipient := entry.Recipient
			if recipient, err = up.GetRecipient(rawRecipient); err != nil {
				r.ctx.Warn("Unable to retrieve Recipient")
				return errors.New(errors.Operational, err)
			}
		}

		// TODO -> Add Gateway Metadata to packet

		bpacket, err := NewBPacket(packet.Payload(), packet.Metadata())
		if err != nil {
			r.ctx.WithError(err).Warn("Unable to create router packet")
			return errors.New(errors.Structural, err)
		}

		response, err := up.Send(bpacket, recipient)

		if err != nil {
			stats.MarkMeter("router.uplink.bad_broker_response")
			r.ctx.WithError(err).Warn("Invalid response from Broker")
			return errors.New(errors.Operational, err)
		}

		itf, err := UnmarshalPacket(response)
		if err != nil {
			stats.MarkMeter("router.uplink.bad_broker_response")
			r.ctx.WithError(err).Warn("Invalid response from Broker")
			return errors.New(errors.Structural, err)
		}

		switch itf.(type) {
		case RPacket:
			ack = itf.(RPacket)
		case nil:
		default:
			return errors.New(errors.Implementation, "Unexpected packet type")
		}

		stats.MarkMeter("router.uplink.ok")
	case SPacket:
		return errors.New(errors.Implementation, "Stats packet not yet implemented")
	case JPacket:
		return errors.New(errors.Implementation, "Join Request not yet implemented")
	default:
		return errors.New(errors.Implementation, "Unreckognized packet type")
	}

	return nil
}

// HandleDown implements the core.Component interface
func (r component) HandleDown(data []byte, an AckNacker, up Adapter) error {
	return errors.New(errors.Implementation, "Handle down not implemented on router")
}

func ensureAckNack(an AckNacker, ack *Packet, err *error) {
	if err != nil && *err != nil {
		an.Nack()
	} else {
		var p Packet
		if ack != nil {
			p = *ack
		}
		an.Ack(p)
	}
}

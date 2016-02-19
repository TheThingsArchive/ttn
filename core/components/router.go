// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package components

import (
	"fmt"

	"github.com/TheThingsNetwork/ttn/core"
	. "github.com/TheThingsNetwork/ttn/core/errors"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	"github.com/TheThingsNetwork/ttn/utils/stats"
	"github.com/apex/log"
)

type Router struct {
	db  RouterStorage // Local storage that maps end-device addresses to broker addresses
	ctx log.Interface // Just a logger
}

// NewRouter constructs a Router and setup its internal structure
func NewRouter(db RouterStorage, ctx log.Interface) *Router {
	return &Router{
		db:  db,
		ctx: ctx,
	}
}

// Register implements the core.Component interface
func (r *Router) Register(reg core.Registration, an core.AckNacker) error {
	stats.MarkMeter("router.registration.in")
	r.ctx.Debug("Handling registration")

	entry := routerEntry{Recipient: reg.Recipient}
	if err := r.db.Store(reg.DevAddr, entry); err != nil {
		stats.MarkMeter("router.registration.failed")
		an.Nack()
		return err
	}
	stats.MarkMeter("router.registration.ok")
	return an.Ack(nil)
}

// HandleDown implements the core.Component interface
func (r *Router) HandleDown(p core.Packet, an core.AckNacker, downAdapter core.Adapter) error {
	return fmt.Errorf("TODO. Not Implemented")
}

// HandleUp implements the core.Component interface
func (r *Router) HandleUp(p core.Packet, an core.AckNacker, upAdapter core.Adapter) error {
	stats.MarkMeter("router.uplink.in")
	r.ctx.Debug("Handling uplink packet")

	var err error

	// Lookup for an existing broker
	devAddr, err := p.DevAddr()
	if err != nil {
		stats.MarkMeter("broker.uplink.invalid")
		r.ctx.Warn("Invalid uplink packet")
		an.Nack()
		return err
	}

	entry, err := r.db.Lookup(devAddr)
	if err != nil && err.(errors.Failure).Nature != ErrWrongBehavior {
		r.ctx.Warn("Database lookup failed")
		an.Nack()
		return err
	}

	var response core.Packet
	if err == nil {
		response, err = upAdapter.Send(p, entry.Recipient)
	} else {
		response, err = upAdapter.Send(p)
	}

	if err != nil {
		stats.MarkMeter("router.uplink.bad_broker_response")
		r.ctx.WithError(err).Warn("Invalid response from Broker")
		an.Nack()
		return err
	}

	stats.MarkMeter("router.uplink.ok")
	return an.Ack(&response)
}

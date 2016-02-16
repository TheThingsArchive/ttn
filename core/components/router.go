// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package components

import (
	"fmt"

	"github.com/TheThingsNetwork/ttn/core"
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
	entry := routerEntry{Recipient: reg.Recipient}
	if err := r.db.Store(reg.DevAddr, entry); err != nil {
		an.Nack()
		return err
	}
	return an.Ack()
}

// HandleDown implements the core.Component interface
func (r *Router) HandleDown(p core.Packet, an core.AckNacker, downAdapter core.Adapter) error {
	return fmt.Errorf("TODO. Not Implemented")
}

// HandleUp implements the core.Component interface
func (r *Router) HandleUp(p core.Packet, an core.AckNacker, upAdapter core.Adapter) error {
	// Lookup for an existing broker
	devAddr, err := p.DevAddr()
	if err != nil {
		an.Nack()
		return err
	}

	entries, err := r.db.Lookup(devAddr)
	if err != ErrDeviceNotFound && err != ErrNotFound && err != ErrEntryExpired {
		an.Nack()
		return err
	}

	response, err := upAdapter.Send(p, entries.Recipient)
	if err != nil {
		an.Nack()
		return err
	}
	return an.Ack(response)
}

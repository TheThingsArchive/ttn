// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package components

import (
	"fmt"
	"time"

	"github.com/TheThingsNetwork/ttn/core"
	"github.com/apex/log"
)

const (
	EXPIRY_DELAY = time.Hour * 8
)

type Router struct {
	Ctx log.Interface
	db  routerStorage // Local storage that maps end-device addresses to broker addresses
}

// NewRouter constructs a Router and setup its internal structure
func NewRouter(ctx log.Interface) (*Router, error) {
	localDB, err := NewRouterStorage(EXPIRY_DELAY)

	if err != nil {
		return nil, err
	}

	return &Router{
		Ctx: ctx,
		db:  localDB,
	}, nil
}

// Register implements the core.Component interface
func (r *Router) Register(reg core.Registration, an core.AckNacker) error {
	if !r.ok() {
		an.Nack()
		return ErrNotInitialized
	}
	if err := r.db.store(reg.DevAddr, reg.Recipient); err != nil {
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
	if !r.ok() {
		an.Nack()
		return ErrNotInitialized
	}

	// Lookup for an existing broker
	devAddr, err := p.DevAddr()
	if err != nil {
		an.Nack()
		return err
	}

	brokers, err := r.db.lookup(devAddr)
	if err != ErrDeviceNotFound && err != ErrEntryExpired {
		an.Nack()
		return err
	}

	response, err := upAdapter.Send(p, brokers...)
	if err != nil {
		an.Nack()
		return err
	}
	return an.Ack(response)
}

// ok ensure the router has been initialized by NewRouter()
func (r *Router) ok() bool {
	return r != nil && r.db != nil
}

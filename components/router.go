// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package components

import (
	"fmt"
	"github.com/thethingsnetwork/core"
	"github.com/thethingsnetwork/core/utils/log"
	"time"
)

const (
	EXPIRY_DELAY = time.Hour * 8
)

type Router struct {
	loggers []log.Logger
	brokers []core.Recipient
	db      routerStorage // Local storage that maps end-device addresses to broker addresses
}

// NewRouter constructs a Router and setup its internal structure
func NewRouter(brokers []core.Recipient, loggers ...log.Logger) (*Router, error) {
	localDB, err := NewRouterStorage(EXPIRY_DELAY)

	if err != nil {
		return nil, err
	}

	if len(brokers) == 0 {
		return nil, ErrBadOptions
	}

	return &Router{
		loggers: loggers,
		brokers: brokers,
		db:      localDB,
	}, nil
}

// Register implements the core.Component interface
func (r *Router) Register(reg core.Registration) error {
	if !r.ok() {
		return ErrNotInitialized
	}
	return r.db.store(reg.DevAddr, reg.Recipient)
}

// HandleDown implements the core.Component interface
func (r *Router) HandleDown(p core.Packet, an core.AckNacker, downAdapter core.Adapter) error {
	return fmt.Errorf("TODO. Not Implemented")
}

// HandleUp implements the core.Component interface
func (r *Router) HandleUp(p core.Packet, an core.AckNacker, upAdapter core.Adapter) error {
	if !r.ok() {
		return ErrNotInitialized
	}

	// Lookup for an existing broker
	devAddr, err := p.DevAddr()
	if err != nil {
		return err
	}

	brokers, err := r.db.lookup(devAddr)
	if err != ErrDeviceNotFound && err != ErrEntryExpired {
		return err
	}

	response, err := upAdapter.Send(p, brokers...)
	if err != nil {
		return err
	}
	return an.Ack(response)
}

// ok ensure the router has been initialized by NewRouter()
func (r *Router) ok() bool {
	return r == nil && r.db != nil
}

// log broadcast the log message to all registered logger
func (r *Router) log(format string, i ...interface{}) {
	for _, logger := range r.loggers {
		logger.Log(format, i...)
	}
}

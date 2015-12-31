// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

// package rtr_brk_mock offers a router <-> broker mock adapter that can be used to test a router
// implementation.
package rtr_brk_mock

import (
	"fmt"
	"github.com/thethingsnetwork/core"
	"github.com/thethingsnetwork/core/lorawan/semtech"
	"time"
)

type Adapter struct {
	FailListen    bool                                        // If true, any call to Listen will fail with an error
	FailBroadcast bool                                        // If true, any call to Broadcast will trigger a core.ErrBroadcast
	FailForward   bool                                        // If true, any call to Forward will trigger a core.ErrForward
	Broadcasts    map[semtech.DeviceAddress][]semtech.Payload // Stores all payload send through broadcasts
	Forwards      map[semtech.DeviceAddress][]semtech.Payload // Stores all payload send through forwards
	connected     bool                                        // Indicates whether or not the Listen() method has been called
}

// New constructs a new router <-> broker adapter interface
func New() Adapter {
	return Adapter{
		FailListen:    false,
		FailBroadcast: false,
		FailForward:   false,
		connected:     false,
		Broadcasts:    make(map[semtech.DeviceAddress][]semtech.Payload),
		Forwards:      make(map[semtech.DeviceAddress][]semtech.Payload),
	}
}

// Connect implements the core.Adapter interface
func (a *Adapter) Listen(router core.Router, options interface{}) error {
	if a.FailListen {
		return fmt.Errorf("Unable to establish the connection")
	}
	a.connected = true
	return nil
}

// Broadcast implements the core.BrokerRouter interface
func (a *Adapter) Broadcast(router core.Router, payload semtech.Payload, broAddrs ...core.BrokerAddress) {
	if !a.connected {
		router.HandleError(core.ErrBroadcast(fmt.Errorf("Try to broadcast with non connected adapter")))
		return
	}

	devAddr, err := payload.UniformDevAddr()

	if a.FailBroadcast || payload.RXPK == nil || err != nil {
		router.HandleError(core.ErrBroadcast(fmt.Errorf("Unable to broadcast given payload %+v", payload)))
		return
	}

	<-time.After(time.Millisecond * 50)
	a.Broadcasts[*devAddr] = append(a.Broadcasts[*devAddr], payload)
	router.RegisterDevice(*devAddr, a.InChargeOf(payload, broAddrs...)...)
}

// InChargeOf returns a set of brokers in charge of a payload (result of simulating a broadcast
// operation).
func (a *Adapter) InChargeOf(payload semtech.Payload, broAddrs ...core.BrokerAddress) []core.BrokerAddress {
	var brokers []core.BrokerAddress
	for i, addr := range broAddrs {
		if i%2 == 1 {
			brokers = append(brokers, addr)
		}
	}
	return brokers
}

// Forward implements the core.BrokerRouter interface
func (a *Adapter) Forward(router core.Router, payload semtech.Payload, broAddrs ...core.BrokerAddress) {
	if !a.connected {
		router.HandleError(core.ErrForward(fmt.Errorf("Try to forward with non connected adapter")))
		return
	}

	devAddr, err := payload.UniformDevAddr()

	if a.FailForward || err != nil {
		router.HandleError(core.ErrForward(fmt.Errorf("Unable to forward given payload %+v", payload)))
		return
	}
	a.Forwards[*devAddr] = append(a.Forwards[*devAddr], payload)
}

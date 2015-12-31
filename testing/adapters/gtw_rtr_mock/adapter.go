// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

// package gtw_rtr_mock offers a gateway <-> router mock adapter that can be used to test a router
// implementation.
package gtw_rtr_mock

import (
	"fmt"
	"github.com/thethingsnetwork/core"
	"github.com/thethingsnetwork/core/lorawan/semtech"
)

type Adapter struct {
	FailAck    bool                                     // If true, each call to Ack will fail with a core.ErrAck
	FailListen bool                                     // If true, each call to Listen will return an error
	connected  bool                                     // Indicate wether or not the Listen method has been called
	Acks       map[core.GatewayAddress][]semtech.Packet // Stores all packet send through Ack()
}

// New constructs a new Gateway-Router-Mock adapter
func New() Adapter {
	return Adapter{
		FailAck:    false,
		FailListen: false,
		connected:  false,
		Acks:       make(map[core.GatewayAddress][]semtech.Packet),
	}
}

// Listen implements the core.Adapter interface
func (a *Adapter) Listen(router core.Router, options interface{}) error {
	if a.FailListen {
		return fmt.Errorf("Unable to establish connection")
	}
	a.connected = true
	return nil
}

// Ack implements the core.GatewayRouterAdapter interface
func (a *Adapter) Ack(router core.Router, packet semtech.Packet, gateway core.GatewayAddress) {
	if !a.connected {
		router.HandleError(core.ErrAck(fmt.Errorf("Try to send ack through non connected adapter")))
		return
	}
	if a.FailAck {
		router.HandleError(core.ErrAck(fmt.Errorf("Unable to ack the given packet")))
		return
	}
	a.Acks[gateway] = append(a.Acks[gateway], packet)
}

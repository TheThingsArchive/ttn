// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

// package gtw_rtr_mock offers a gateway <-> router mock adapter that can be used to test a router
// implementation.
package gtw_rtr_mock

import (
	"github.com/thethingsnetwork/core"
	"github.com/thethingsnetwork/core/lorawan/semtech"
)

type Adapter struct {
	FailAck    bool                                     // If true, each call to Ack will fail with a core.ErrAck
	FailListen bool                                     // If true, each call to Listen will return an error
	Acks       map[core.GatewayAddress][]semtech.Packet // Stores all packet send through Ack()
}

// New constructs a new Gateway-Router-Mock adapter
func New() *Adapter {
	return &Adapter{
		FailAck:    false,
		FailListen: false,
		Acks:       make(map[core.GatewayAddress][]semtech.Packet),
	}
}

// Listen implements the core.Adapter interface
func (a *Adapter) Listen(router core.Router, options interface{}) error {
	if a.FailListen {
		return core.ErrBadOptions
	}
	return nil
}

// Ack implements the core.GatewayRouterAdapter interface
func (a *Adapter) Ack(router core.Router, packet semtech.Packet, gateway core.GatewayAddress) error {
	if a.FailAck {
		return core.ErrInvalidPacket
	}
	a.Acks[gateway] = append(a.Acks[gateway], packet)
	return nil
}

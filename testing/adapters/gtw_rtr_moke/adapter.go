// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

// package gtw_rtr_moke offers a gateway <-> router moke adapter that can be used to test a router
// implementation.
package gtw_rtr_moke

import (
	"fmt"
	"github.com/thethingsnetwork/core"
	"github.com/thethingsnetwork/core/lorawan/semtech"
)

type Adapter struct {
	FailAck     bool
	FailConnect bool
	connected   bool
	acks        map[core.GatewayAddress][]semtech.Packet
}

// New constructs a new Gateway-Router-Moke adapter
func New() (*Adapter, error) {
	return &Adapter{
		FailAck:     false,
		FailConnect: false,
		connected:   false,
		acks:        make(map[core.GatewayAddress][]semtech.Packet),
	}, nil
}

// Listen implements the core.Adapter interface
func (a *Adapter) Listen(router core.Router, options interface{}) error {
	if a.FailConnect {
		return fmt.Errorf("Unable to establish connection")
	}
	a.connected = true
	return nil
}

// Ack implements the core.GatewayRouterAdapter interface
func (a *Adapter) Ack(router core.Router, packet semtech.Packet, gateway core.GatewayAddress) {
	if a.FailAck {
		router.HandleError(core.ErrAck(fmt.Errorf("Unable to ack the given packet")))
		return
	}
	a.acks[gateway] = append(a.acks[gateway], packet)
}

func (a *Adapter) GetAcks(gateway core.GatewayAddress) []semtech.Packet {
	return a.acks[gateway]
}

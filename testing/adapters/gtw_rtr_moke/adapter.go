// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package gtw_rtr_moke

import (
	"github.com/thethingsnetwork/core"
	"github.com/thethingsnetwork/core/lorawan/semtech"
)

type Adapter struct{}

// New constructs a new Gateway-Router-Moke adapter
func New(router core.Router, port uint) (*Adapter, error) {
	return nil, nil
}

// Ack implements the core.GatewayRouterAdapter interface
func (a *Adapter) Connect(router core.Router, port uint) error {
	return nil
}

// Ack implements the core.GatewayRouterAdapter interface
func (a *Adapter) Ack(router core.Router, packet semtech.Packet, gateway core.GatewayAddress) {
}

// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

// package rtr_brk_mock offers a router <-> broker mock adapter that can be used to test a router
// implementation.
package rtr_brk_mock

import (
	"github.com/thethingsnetwork/core"
	"github.com/thethingsnetwork/core/lorawan/semtech"
)

type Adapter struct{}

// Connect implements the core.Adapter interface
func (a *Adapter) Listen(router core.Router, options interface{}) error {
	return nil
}

// Broadcast implements the core.BrokerRouter interface
func (a *Adapter) Broadcast(router core.Router, payload semtech.Payload, broAddrs ...core.BrokerAddress) {
}

// Forward implements the core.BrokerRouter interface
func (a *Adapter) Forward(router core.Router, payload semtech.Payload, broAddrs ...core.BrokerAddress) {
}

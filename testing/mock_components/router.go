// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

// package mock_components offers a mock router that can be used to test adapter implementations.
package mock_components

import (
	"github.com/thethingsnetwork/core"
	"github.com/thethingsnetwork/core/lorawan/semtech"
)

// Router represents the mock Router
type Router struct {
	Errors   []interface{}
	Packets  map[core.GatewayAddress][]semtech.Packet
	Payloads map[core.BrokerAddress][]semtech.Payload
	Devices  map[semtech.DeviceAddress][]core.BrokerAddress
}

// New constructs a mock Router; This method is a shortcut that creates all the internal maps.
func New() Router {
	return Router{
		Packets:  make(map[core.GatewayAddress][]semtech.Packet),
		Payloads: make(map[core.BrokerAddress][]semtech.Payload),
		Devices:  make(map[semtech.DeviceAddress][]core.BrokerAddress),
	}
}

// HandleUplink implements the core.Router interface
func (r *Router) HandleUplink(packet semtech.Packet, gateway core.GatewayAddress) {
	r.Packets[gateway] = append(r.Packets[gateway], packet)
}

// HandleDownlink implements the core.Router interface
func (r *Router) HandleDownlink(payload semtech.Payload, broker core.BrokerAddress) {
	r.Payloads[broker] = append(r.Payloads[broker], payload)
}

// RegisterDevice implements the core.Router interface
func (r *Router) RegisterDevice(devAddr semtech.DeviceAddress, broAddrs ...core.BrokerAddress) {
	r.Devices[devAddr] = append(r.Devices[devAddr], broAddrs...)
}

// RegisterDevice implements the core.Router interface
func (r *Router) HandleError(err interface{}) {
	r.Errors = append(r.Errors, err)
}

// Connect implements the core.Router interface
func (r *Router) Connect(upAdapter core.GatewayRouterAdapter, downAdapter core.RouterBrokerAdapter) {
	// Chill
}

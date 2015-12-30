// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package router

import (
	"fmt"
	"github.com/thethingsnetwork/core"
	"github.com/thethingsnetwork/core/lorawan/semtech"
	"github.com/thethingsnetwork/core/utils/log"
	"time"
)

const (
	EXPIRY_DELAY = time.Hour * 8
)

type Router struct {
	Port      uint
	upAdapter core.GatewayRouterAdapter
	logger    log.Logger
}

func New(upAdapter core.GatewayRouterAdapter, downAdapter core.RouterBrokerAdapter, port uint) (*Router, error) {
	return &Router{
		Port:      port,
		logger:    log.VoidLogger{},
		upAdapter: upAdapter,
	}, nil
}

// HandleUplink implements the core.Router interface
func (r *Router) HandleUplink(packet semtech.Packet, connId core.ConnectionId) {

	/* PULL_DATA
	 *
	 * Send PULL_ACK
	 * Store the gateway in known gateway
	 */

	/* PUSH_DATA
	 *
	 * Send PUSH_ACK
	 * Stores the gateway connection id for later response
	 * Lookup for an existing broker associated to the device address
	 * Forward data to that broker
	 */

	/* Else
	 *
	 * Ignore / Raise an error
	 */

}

// HandleDownlink implements the core.Router interface
func (r *Router) HandleDownlink(packet semtech.Packet) {
	// TODO
}

// RegisterDevice implements the core.Router interface
func (r *Router) RegisterDevice(devAddr core.DeviceAddress, broAddrs ...core.BrokerAddress) {
	// TODO
}

// RegisterDevice implements the core.Router interface
func (r *Router) HandleError(err error) {
	fmt.Println(err)
}

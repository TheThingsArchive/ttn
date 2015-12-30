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

func (r *Router) log(format string, a ...interface{}) {
	r.logger.Log(format, a...)
}

// HandleUplink implements the core.Router interface
func (r *Router) HandleUplink(packet semtech.Packet, connId core.ConnectionId) {
	switch packet.Identifier {
	case semtech.PULL_DATA:
		r.log("PULL_DATA received, sending ack")
		r.upAdapter.Ack(semtech.Packet{
			Version:    semtech.VERSION,
			Identifier: semtech.PULL_ACK,
			Token:      packet.Token,
		}, connId)
	case semtech.PUSH_DATA:
		r.log("TODO PUSH_DATA")
		/* PUSH_DATA
		 *
		 * Send PUSH_ACK
		 * Stores the gateway connection id for later response
		 * Lookup for an existing broker associated to the device address
		 * Forward data to that broker
		 */
	default:
		r.log("Unexpected packet receive from uplink %+v", packet)

	}
}

// HandleDownlink implements the core.Router interface
func (r *Router) HandleDownlink(packet semtech.Packet) {
	// TODO MileStone 4
}

// RegisterDevice implements the core.Router interface
func (r *Router) RegisterDevice(devAddr core.DeviceAddress, broAddrs ...core.BrokerAddress) {
	// TODO MileStone 4
}

// RegisterDevice implements the core.Router interface
func (r *Router) HandleError(err error) {
	fmt.Println(err) // Wow, much handling, very reliable
}

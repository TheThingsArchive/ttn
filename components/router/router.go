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
	PortUDP  uint
	PortHTTP uint
	Logger   log.Logger
}

func New(portUDP, portHTTP uint) (*Router, error) {
	return &Router{
		PortUDP:  portUDP,
		PortHTTP: portHTTP,
		Logger:   log.VoidLogger{},
	}, nil
}

// HandleUplink implements the core.Router interface
func (r *Router) HandleUplink(upAdapter core.GatewayRouterAdapter, packet semtech.Packet, gateway core.GatewayAddress) {
	switch packet.Identifier {
	case semtech.PULL_DATA:
		r.log("PULL_DATA received, sending ack")
		upAdapter.Ack(r, semtech.Packet{
			Version:    semtech.VERSION,
			Identifier: semtech.PULL_ACK,
			Token:      packet.Token,
		}, gateway)
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
func (r *Router) HandleDownlink(downAdapter core.RouterBrokerAdapter, packet semtech.Packet, broker core.BrokerAddress) {
	// TODO MileStone 4
}

// RegisterDevice implements the core.Router interface
func (r *Router) RegisterDevice(devAddr core.DeviceAddress, broAddrs ...core.BrokerAddress) {
	// TODO MileStone 4
}

// RegisterDevice implements the core.Router interface
func (r *Router) HandleError(err interface{}) {
	switch err.(type) {
	case core.ErrAck:
	case core.ErrDownlink:
	case core.ErrForward:
	case core.ErrBroadcast:
	case core.ErrUplink:
	default:
		fmt.Println(err) // Wow, much handling, very reliable
	}
}

func (r *Router) log(format string, a ...interface{}) {
	if r.Logger == nil {
		return
	}
	r.Logger.Log(format, a...)
}

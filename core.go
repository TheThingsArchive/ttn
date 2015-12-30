// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package core

import (
	. "github.com/thethingsnetwork/core/lorawan/semtech"
)

type BrokerAddress string
type GatewayAddress string

type Router interface {
	// HandlerError manages all kind of error that occur during the router lifecycle
	HandleError(err interface{})

	// HandleUplink manages uplink packets coming from a gateway
	HandleUplink(upAdapter GatewayRouterAdapter, packet Packet, gateway GatewayAddress)

	// HandleDownlink manages downlink packets coming from a broker
	HandleDownlink(downAdapter RouterBrokerAdapter, packet Packet, broker BrokerAddress)

	// RegisterDevice associates a device address to a set of brokers for a given period
	RegisterDevice(devAddr DeviceAddress, broAddrs ...BrokerAddress)
}

// The error types belows are going to be more complex in order to handle custom behavior for
// each error type.
type ErrUplink error
type ErrAck error

type GatewayRouterAdapter interface {
	// Establish the adapter connection, whatever protocol is being used.
	Connect(router Router, port uint) error

	// Ack allows the router to send back a response to a gateway. The name of the method is quite a
	// bad call and will probably change soon.
	Ack(router Router, packet Packet, gateway GatewayAddress)
}

type ErrDownlink error
type ErrForward error
type ErrBroadcast error
type RouterBrokerAdapter interface {
	// Establish the adapter connection, whatever protocol is being used.
	Connect(router Router, broAddrs ...BrokerAddress) error

	// Broadcast makes the adapter discover all available brokers by sending them a the given packets.
	//
	// We assume that broadcast is also registering a device address towards the router depending
	// on the brokers responses.
	Broadcast(router Router, payload Payload)

	// Forward is an explicit forwarding of a packet which is known being handled by a set of
	// brokers. None of the contacted broker is supposed to reject the incoming payload; They all
	// ave been queried before and are known as dedicated brokers for the related end-device.
	Forward(router Router, packet Payload, broAddrs ...BrokerAddress)
}

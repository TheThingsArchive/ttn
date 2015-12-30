// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package core

import (
	. "github.com/thethingsnetwork/core/lorawan/semtech"
)

type DeviceAddress string
type BrokerAddress string
type GatewayAddress string

type Router interface {
	HandleError(err interface{})
	HandleUplink(upAdapter GatewayRouterAdapter, packet Packet, gateway GatewayAddress)
	HandleDownlink(downAdapter RouterBrokerAdapter, packet Packet, broker BrokerAddress)
	RegisterDevice(devAddr DeviceAddress, broAddrs ...BrokerAddress)
}

type ErrUplink error
type ErrAck error
type GatewayRouterAdapter interface {
	Connect(router Router, port uint) error
	Ack(router Router, packet Packet, gateway GatewayAddress)
}

type ErrDownlink error
type ErrForward error
type ErrBroadcast error
type RouterBrokerAdapter interface {
	Connect(router Router, broAddrs ...BrokerAddress) error
	Broadcast(router Router, payload Packet)
	Forward(router Router, packet Packet, broAddrs ...BrokerAddress)
}

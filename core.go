// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package core

import (
	. "github.com/thethingsnetwork/core/lorawan/semtech"
)

type DeviceAddress string
type BrokerAddress string
type ConnectionId string

type Router interface {
	HandleError(err error)
	HandleUplink(packet Packet, connId ConnectionId)
	HandleDownlink(packet Packet, connId ConnectionId)
	RegisterDevice(devAddr DeviceAddress, broAddrs ...BrokerAddress)
}

type ErrAck error
type GatewayRouterAdapter interface {
	Connect(router Router)
	Ack(packet Packet, cid ConnectionId)
}

type ErrForward error
type RouterBrokerAdapter interface {
	Connect(router Router)
	Broadcast(payload Packet)
	Forward(packet Packet, broAddrs ...BrokerAddress)
}

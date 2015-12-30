package core

import (
	. "github.com/thethingsnetwork/core/lorawan/semtech"
)

type DeviceAddress string
type BrokerAddress string
type ConnectionId uint
type GatewayId [8]byte

type Router interface {
	HandleError(err error)
	HandleUplink(packet Packet, connId ConnectionId)
	HandleDownlink(packet Packet, connId ConnectionId)
	RegisterDevice(devAddr DeviceAddress, broAddrs ...BrokerAddress)
}

type GatewayRouterAdapter interface {
	Connect(router Router)
	Ack(packet Packet, gid GatewayId)
}

type RouterBrokerAdapter interface {
	Connect(router Router)
	Broadcast(packet Packet)
	Forward(packet Packet, broAddrs ...BrokerAddress)
}

package core

import (
	. "github.com/thethingsnetwork/core/lorawan/semtech"
)

type DeviceAddress string
type BrokerAddress string
type ConnectionId uint
type GatewayId [8]byte

type Component interface {
	HandleError(err error)
}

type Router interface {
	Component
	HandleUplink(packet Packet, connId ConnectionId)
	HandleDownlink(packet Packet, connId ConnectionId)
	RegisterDevice(devAddr DeviceAddress, broAddrs ...BrokerAddress)
}

type Adapter interface {
	Connect(comp Component)
}

type GatewayRouterAdapter interface {
	Adapter
	Ack(packet Packet, gid GatewayId)
}

type RouterBrokerAdapter interface {
	Adapter
	Broadcast(packet Packet)
	Forward(packet Packet, broAddrs ...BrokerAddress)
}

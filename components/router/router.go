// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package router

import (
	"fmt"
	"github.com/thethingsnetwork/core"
	"github.com/thethingsnetwork/core/lorawan/semtech"
	"github.com/thethingsnetwork/core/utils/log"
	"net"
	"time"
)

const (
	EXPIRY_DELAY = time.Hour * 8
)

type Router struct{}

type errAck error

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

func (r *Router) HandleDownlink(packet semtech.Packet) {

}

func (r *Router) RegisterDevice(devAddr core.DeviceAddress, broAddrs ...core.BrokerAddress) {

}

func (r *Router) HandleError(err error) {
}

// --------------- Routers Adapters
type UpAdapter struct {
	router   *Router
	logger   log.Logger
	gateways map[core.GatewayId]net.UDPConn
}

func NewUpAdapter(router Router) *UpAdapter {
	adapter := UpAdapter{
		gateways: make(map[core.GatewayId]net.UDPConn),
		logger:   log.VoidLogger{},
	}
	adapter.Connect(router)
	return &adapter
}

func (u UpAdapter) log(format string, a ...interface{}) {
	u.logger.Log(format, a...)
}

func (u *UpAdapter) Ack(packet semtech.Packet, gid core.GatewayId) {
	if u.router == nil {
		u.log("Fails to Ack, not connected to a router")
		return
	}

	u.log("Acks packet %+v", packet)

	conn, ok := u.gateways[gid]

	if !ok {
		u.log("Gateway connection not found")
		u.router.HandleError(errAck(fmt.Errorf("Gateway connection not found")))
		return
	}

	raw, err := semtech.Marshal(packet)

	if err != nil {
		u.log("Unable to marshal given packet")
		u.router.HandleError(errAck(fmt.Errorf("Unable to marshal given packet %+v", err)))
		return
	}

	_, err = conn.Write(raw)

	if err != nil {
		u.log("Unable to send udp message")
		u.router.HandleError(errAck(fmt.Errorf("Unable to send udp message %+v", err)))
		return
	}
}

func (u *UpAdapter) Connect(router Router) {
	u.log("Connects to router %+v", router)
	u.router = &router
}

type DownAdapter struct {
}

func (d *DownAdapter) Broadcast(packet semtech.Packet) {

}

func (d *DownAdapter) Forward(packet semtech.Packet, broAddrs ...core.BrokerAddress) {

}

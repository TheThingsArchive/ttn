// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package router

import (
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
	router   Router
	logger   log.logger
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
		u.log("Failed to Ack, not connected to a router")
	}
}

func (u *UpAdapter) Connect(router Router) {
	u.router = router
}

type DownAdapter struct {
}

func (d *DownAdapter) Broadcast(packet semtech.Packet) {

}

func (d *DownAdapter) Forward(packet semtech.Packet, broAddrs ...core.BrokerAddress) {

}

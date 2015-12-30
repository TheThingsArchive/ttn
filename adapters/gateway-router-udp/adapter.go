// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package gtw_rtr_udp

import (
	"fmt"
	"github.com/thethingsnetwork/core"
	"github.com/thethingsnetwork/core/lorawan/semtech"
	"github.com/thethingsnetwork/core/utils/log"
	"net"
	"sync"
)

type Adapter struct {
	router   core.Router
	logger   log.Logger
	gateways map[core.ConnectionId]*net.UDPAddr
	conn     *net.UDPConn
	lock     sync.RWMutex
}

// New constructs a new Gateway-Router-UDP adapter
func New(router core.Router, port uint) (*Adapter, error) {
	adapter := Adapter{
		gateways: make(map[core.ConnectionId]*net.UDPAddr),
		lock:     sync.RWMutex{},
		logger:   log.VoidLogger{},
	}

	// Connect to the router and start listening on the given port of the current machine
	adapter.Connect(router)
	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("0.0.0.0:%d", port))
	if err != nil {
		return nil, err
	}
	udpConn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return nil, err
	}
	adapter.conn = udpConn
	go adapter.listen() // NOTE: There is no way to stop properly the adapter and thus this goroutine for now.

	// Return the adapter for further use
	return &adapter, nil
}

// log is nothing more than a shortcut / helper to access the logger
func (a Adapter) log(format string, i ...interface{}) {
	a.logger.Log(format, i...)
}

// Ack implements the core.GatewayRouterAdapter interface
func (a *Adapter) Ack(packet semtech.Packet, cid core.ConnectionId) {
	if a.router == nil {
		a.log("Fails to Ack, not connected to a router")
		return
	}

	a.log("Acks packet %+v", packet)

	a.lock.RLock()
	addr, ok := a.gateways[cid]
	a.lock.Unlock()

	if !ok {
		a.log("Gateway connection not found")
		a.router.HandleError(core.ErrAck(fmt.Errorf("Gateway connection not found")))
		return
	}

	raw, err := semtech.Marshal(packet)

	if err != nil {
		a.log("Unable to marshal given packet")
		a.router.HandleError(core.ErrAck(fmt.Errorf("Unable to marshal given packet %+v", err)))
		return
	}

	_, err = a.conn.WriteToUDP(raw, addr)

	if err != nil {
		a.log("Unable to send udp message")
		a.router.HandleError(core.ErrAck(fmt.Errorf("Unable to send udp message %+v", err)))
		return
	}
}

// Ack implements the core.GatewayRouterAdapter interface
func (a *Adapter) Connect(router core.Router) {
	a.log("Connects to router %+v", router)
	a.router = router
}

func (a *Adapter) listen() {
	for {
		buf := make([]byte, 1024)
		n, addr, err := a.conn.ReadFromUDP(buf)
		if err != nil {
			a.log("Error: %v", err) // NOTE Errors are just ignored for now
			continue
		}
		a.log("Incoming datagram %x", buf[:n])

		pkt, err := semtech.Unmarshal(buf[:n])
		if err != nil {
			a.log("Error: %v", err) // NOTE Errors are just ignored for now
			continue
		}

		// When a packet is received pass it to the router for processing
		cid := core.ConnectionId(addr.String())
		a.lock.Lock()
		a.gateways[cid] = addr
		a.lock.Unlock()
		a.router.HandleUplink(*pkt, cid)
	}
}

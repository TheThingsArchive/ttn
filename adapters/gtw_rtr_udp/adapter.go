// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package gtw_rtr_udp

import (
	"fmt"
	"github.com/thethingsnetwork/core"
	"github.com/thethingsnetwork/core/lorawan/semtech"
	"github.com/thethingsnetwork/core/utils/log"
	"net"
)

type Adapter struct {
	Logger log.Logger
	conn   *net.UDPConn
}

// New constructs a new Gateway-Router-UDP adapter
func New(router core.Router, port uint) (*Adapter, error) {
	adapter := Adapter{
		Logger: log.VoidLogger{},
	}

	// Connect to the router and start listening on the given port of the current machine
	if err := adapter.Connect(router, port); err != nil {
		return nil, err
	}

	// Return the adapter for further use
	return &adapter, nil
}

// Ack implements the core.GatewayRouterAdapter interface
func (a *Adapter) Connect(router core.Router, port uint) error {
	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("0.0.0.0:%d", port))
	if err != nil {
		return err
	}
	a.conn, err = net.ListenUDP("udp", addr)
	if err != nil {
		return err
	}
	go a.listen(router) // NOTE: There is no way to stop properly the adapter and thus this goroutine for now.
	return nil
}

// Ack implements the core.GatewayRouterAdapter interface
func (a *Adapter) Ack(router core.Router, packet semtech.Packet, gateway core.GatewayAddress) {
	if a.conn == nil {
		a.log("Connection not established. Connect the adaptor first.")
		router.HandleError(core.ErrAck(fmt.Errorf("Connection not established. Connect the adaptor first.")))
		return
	}

	a.log("Acks packet %+v", packet)

	addr, err := net.ResolveUDPAddr("udp", string(gateway))

	if err != nil {
		a.log("Unable to retrieve gateway address")
		router.HandleError(core.ErrAck(err))
		return
	}

	raw, err := semtech.Marshal(packet)

	if err != nil {
		a.log("Unable to marshal given packet")
		router.HandleError(core.ErrAck(err))
		return
	}

	_, err = a.conn.WriteToUDP(raw, addr)

	if err != nil {
		a.log("Unable to send udp message")
		router.HandleError(core.ErrAck(err))
		return
	}
}

// listen Handle incominng packets and forward them to the router
func (a *Adapter) listen(router core.Router) {
	for {
		buf := make([]byte, 1024)
		n, addr, err := a.conn.ReadFromUDP(buf)
		if err != nil {
			a.log("Error: %v", err)
			go router.HandleError(core.ErrUplink(err))
			continue
		}
		a.log("Incoming datagram %x", buf[:n])

		pkt, err := semtech.Unmarshal(buf[:n])
		if err != nil {
			a.log("Error: %v", err)
			go router.HandleError(core.ErrUplink(err))
			continue
		}

		// When a packet is received pass it to the router for processing
		router.HandleUplink(a, *pkt, core.GatewayAddress(addr.String()))
	}
}

// log is nothing more than a shortcut / helper to access the logger
func (a Adapter) log(format string, i ...interface{}) {
	if a.Logger == nil {
		return
	}
	a.Logger.Log(format, i...)
}

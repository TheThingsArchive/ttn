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
	Logger log.Logger  // A custom logger used to report errors
	conn   chan udpMsg // An internal communication channel use to send udp datagram through a valid connection
}

type udpMsg struct {
	addr *net.UDPAddr // The target gateway address targetted
	raw  []byte       // The raw byte sequence that has to be sent
	conn *net.UDPConn // Provide if you intent to change the current adapter connection
}

// NewAdapter constructs a gateway <-> router udp adapter
func NewAdapter() Adapter {
	a := Adapter{conn: make(chan udpMsg)}
	go a.monitorConnection() // Terminates that goroutine by closing the channel
	return a
}

// ok controls whether or not the adapter has been initialized via NewAdapter()
func (a *Adapter) ok() bool {
	return a != nil && a.conn != nil
}

// Listen implements the core.Adapter interface. It expects only one param "port" as a
// uint. Listen can be called several times to re-establish a lost connection.
func (a *Adapter) Listen(router core.Router, options interface{}) error {
	if !a.ok() {
		return core.ErrNotInitialized
	}

	// Parse options
	var port uint
	switch options.(type) {
	case uint:
		port = options.(uint)
	default:
		a.log("Invalid option provided: %+v", options)
		return core.ErrBadOptions
	}

	// Create the udp connection and start listening with a goroutine
	var udpConn *net.UDPConn
	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("0.0.0.0:%d", port))
	if udpConn, err = net.ListenUDP("udp", addr); err != nil {
		a.log("Unable to establish the connection: %v", err)
		return core.ErrBadGatewayAddress
	}

	a.conn <- udpMsg{conn: udpConn}
	go a.listen(router, udpConn) // Terminates when the connection is closed

	return nil
}

// Ack implements the core.GatewayRouterAdapter interface
func (a *Adapter) Ack(router core.Router, packet semtech.Packet, gateway core.GatewayAddress) error {
	if !a.ok() {
		return core.ErrNotInitialized
	}

	a.log("Acks packet %+v", packet)

	addr, err := net.ResolveUDPAddr("udp", string(gateway))

	if err != nil {
		a.log("Unable to retrieve gateway address: %+v", err)
		return core.ErrBadGatewayAddress
	}

	raw, err := semtech.Marshal(packet)

	if err != nil {
		a.log("Unable to marshal given packet: %+v", err)
		return core.ErrInvalidPacket
	}

	a.conn <- udpMsg{raw: raw, addr: addr}
	return nil
}

// listen Handle incoming packets and forward them to the router
func (a *Adapter) listen(router core.Router, conn *net.UDPConn) {
	defer conn.Close()
	a.log("Start listening on %s", conn.LocalAddr())
	for {
		buf := make([]byte, 1024)
		n, addr, err := conn.ReadFromUDP(buf)
		if err != nil { // Problem with the connection
			a.log("Error: %v", err)
			go router.HandleError(core.ErrMissingConnection)
			return
		}
		a.log("Incoming datagram %x", buf[:n])

		pkt, err := semtech.Unmarshal(buf[:n])
		if err != nil {
			a.log("Error: %v", err)
			go router.HandleError(core.ErrInvalidPacket)
			continue
		}

		// When a packet is received pass it to the router for processing
		router.HandleUplink(*pkt, core.GatewayAddress(addr.String()))
	}
}

// monitorConnection manages udpConnection of the adapter and send message through that connection
func (a *Adapter) monitorConnection() {
	var udpConn *net.UDPConn
	for msg := range a.conn {
		if msg.conn != nil { // Change the connection
			if udpConn != nil {
				a.log("Switch UDP connection")
				udpConn.Close()
			}
			udpConn = msg.conn
		}

		if udpConn != nil && msg.raw != nil { // Send the given udp message
			if _, err := udpConn.WriteToUDP(msg.raw, msg.addr); err != nil {
				a.log("Unable to send udp message: %+v", err)
			}
		}
	}
	if udpConn != nil {
		udpConn.Close() // Make sure we close the connection before leaving
	}
}

// log is nothing more than a shortcut / helper to access the logger
func (a Adapter) log(format string, i ...interface{}) {
	if a.Logger == nil {
		return
	}
	a.Logger.Log(format, i...)
}

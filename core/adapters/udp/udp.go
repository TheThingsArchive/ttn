// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package udp

import (
	"fmt"
	"net"

	"github.com/TheThingsNetwork/ttn/core"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	"github.com/apex/log"
)

// Adapter represents a udp adapter which sends and receives packet via UDP
type Adapter struct {
	ctx      log.Interface    // Just a logger
	conn     chan MsgUdp      // Channel used to manage response transmissions made by multiple goroutines
	packets  chan MsgReq      // Incoming valid packets are pushed to this channel and consume by an outsider
	handlers chan interface{} // Manage handlers, could be either a Handler or a []byte (new handler or handling action)
}

// Handler represents a datagram and packet handler used by the adapter to process packets
type Handler interface {
	// Handle handles incoming datagram from a gateway transmitter to the network
	Handle(conn chan<- MsgUdp, chresp chan<- MsgReq, msg MsgUdp)
}

// MsgUdp type materializes response messages transmitted towards existing recipients (commonly, gateways).
type MsgUdp struct {
	Data []byte       // The raw byte sequence that has to be sent
	Addr *net.UDPAddr // The target recipient address
}

// OutMsg type materializes valid uplink messages coming from a given recipient
type MsgReq struct {
	Data   []byte      // The actual message
	Chresp chan MsgRes // A dedicated response channel
}

// Message sent through the response channel of MsgReq
type MsgRes []byte // The actual message

// NewAdapter constructs and allocates a new udp adapter
func NewAdapter(port uint, ctx log.Interface) (*Adapter, error) {
	a := Adapter{
		ctx:      ctx,
		conn:     make(chan MsgUdp),
		packets:  make(chan MsgReq),
		handlers: make(chan interface{}),
	}

	// Create the udp connection and start listening with a goroutine
	var udpConn *net.UDPConn
	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("0.0.0.0:%d", port))
	a.ctx.WithField("port", port).Info("Starting Server")
	if udpConn, err = net.ListenUDP("udp", addr); err != nil {
		a.ctx.WithError(err).Error("Unable to start server")
		return nil, errors.New(errors.Structural, fmt.Sprintf("Invalid port %v", port))
	}

	go a.monitorConnection(udpConn)
	go a.monitorHandlers()
	go a.listen(udpConn)

	return &a, nil
}

// Send implements the core.Adapter interface. Not implemented for the udp adapter.
func (a *Adapter) Send(p core.Packet, r ...core.Recipient) ([]byte, error) {
	return nil, errors.New(errors.Implementation, "Send not supported on udp adapter")
}

// GetRecipient implements the core.Adapter interface
func (a *Adapter) GetRecipient(raw []byte) (core.Recipient, error) {
	return nil, errors.New(errors.Implementation, "GetRecipient not supported on udp adapter")
}

// Next implements the core.Adapter interface
func (a *Adapter) Next() ([]byte, core.AckNacker, error) {
	msg := <-a.packets
	return msg.Data, udpAckNacker{Chresp: msg.Chresp}, nil
}

// NextRegistration implements the core.Adapter interface
func (a *Adapter) NextRegistration() (core.Registration, core.AckNacker, error) {
	return udpRegistration{}, nil, errors.New(errors.Implementation, "NextRegistration not supported on udp adapter")
}

func (a *Adapter) Bind(h Handler) {
	a.handlers <- h
}

// listen Handle incoming packets and forward them.
//
// Runs in its own goroutine.
func (a *Adapter) listen(conn *net.UDPConn) {
	defer conn.Close()
	a.ctx.WithField("address", conn.LocalAddr()).Debug("Starting accept loop")
	for {
		buf := make([]byte, 5000)
		n, addr, err := conn.ReadFromUDP(buf)
		if err != nil { // Problem with the connection
			a.ctx.WithError(err).Error("Connection error")
			continue
		}

		a.ctx.Debug("Incoming datagram")
		a.handlers <- MsgUdp{Addr: addr, Data: buf[:n]}
	}
}

// monitorConnection manages udpConnection of the adapter and send message through that connection
//
// That function executes into a single goroutine and is the only one allowed to write UDP messages.
// Doing this makes sure that only 1 goroutine is interacting with the connection.
//
// Runs in its own goroutine
func (a *Adapter) monitorConnection(udpConn *net.UDPConn) {
	for msg := range a.conn {
		if msg.Data != nil { // Send the given udp message
			if _, err := udpConn.WriteToUDP(msg.Data, msg.Addr); err != nil {
				a.ctx.WithError(err).Error("Error while sending UDP message")
			}
		}
	}
	if udpConn != nil {
		udpConn.Close() // Make sure we close the connection before leaving if we dare ever leave.
	}
}

// monitorHandlers manages handler registration and execution concurrently. One can pass a new
// handler through the handlers channel to declare a new one or, send directly data through the
// channel to ask every defined handler to handle them.
//
// Runs in its own goroutine
func (a *Adapter) monitorHandlers() {
	var handlers []Handler

	for msg := range a.handlers {
		switch msg.(type) {
		case Handler:
			handlers = append(handlers, msg.(Handler))
		case MsgUdp:
			for _, h := range handlers {
				go func(h Handler, msg MsgUdp) {
					h.Handle(a.conn, a.packets, msg)
				}(h, msg.(MsgUdp))
			}
		}
	}
}

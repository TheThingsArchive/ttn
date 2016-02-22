// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package udp

import (
	"fmt"
	"net"

	core "github.com/TheThingsNetwork/ttn/refactor"
	. "github.com/TheThingsNetwork/ttn/refactor/errors"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	"github.com/apex/log"
)

// Adapter represents a udp adapter which sends and receives packet via UDP
type Adapter struct {
	Handler
	ctx  log.Interface // Just a logger
	conn chan UdpMsg   // Channel used to manage response transmissions made by multiple goroutines
	next chan OutMsg   // Incoming valid packets are pushed to this channel and consume by an outsider
	ack  chan AckMsg   // Channel used to consume ack or nack sent to the adapter
}

// Handler represents a datagram and packet handler used by the adapter to process packets
type Handler interface {
	// HandleAck handles a positive response to a transmitter
	HandleAck(p *core.Packet, resp chan<- HandlerMsg)

	// HandleNack handles a negative response to a transmitter
	HandleNack(resp chan<- HandlerMsg)

	// HandleDatagram handles incoming datagram from a gateway transmitter to the network
	HandleDatagram(data []byte, resp chan<- HandlerMsg)
}

// HandlerMsg type materializes response messages emitted by the UdpHandler
type HandlerMsg struct {
	Data []byte
	Type HandlerMsgType
}

type HandlerMsgType byte

const (
	HANDLER_RESP  HandlerMsgType = iota // A response towards the udp transmitter
	HANDLER_OUT                         // A response towards the network
	HANDLER_ERROR                       // An error during the process
)

// AckMsg type materializes ack or nack messages flowing into the Ack channel
type AckMsg struct {
	Packet *core.Packet
	Type   AckMsgType
	Addr   *net.UDPAddr
	Cherr  chan<- error
}

type AckMsgType bool

const (
	AN_ACK  AckMsgType = true
	AN_NACK AckMsgType = false
)

// UdpMsg type materializes response messages transmitted towards existing recipients (commonly, gateways).
type UdpMsg struct {
	Data []byte       // The raw byte sequence that has to be sent
	Conn *net.UDPConn // Provide if you intent to change the current adapter connection
	Addr *net.UDPAddr // The target recipient address
}

// OutMsg type materializes valid uplink messages coming from a given recipient
type OutMsg struct {
	Data []byte       // The actual message
	Addr *net.UDPAddr // The address of the source emitter
}

// NewAdapter constructs and allocates a new udp adapter
func NewAdapter(port uint, handler Handler, ctx log.Interface) (*Adapter, error) {
	a := Adapter{
		Handler: handler,
		ctx:     ctx,
		conn:    make(chan UdpMsg),
		next:    make(chan OutMsg),
	}

	// Create the udp connection and start listening with a goroutine
	var udpConn *net.UDPConn
	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("0.0.0.0:%d", port))
	a.ctx.WithField("port", port).Info("Starting Server")
	if udpConn, err = net.ListenUDP("udp", addr); err != nil {
		a.ctx.WithError(err).Error("Unable to start server")
		return nil, errors.New(ErrInvalidStructure, fmt.Sprintf("Invalid port %v", port))
	}

	go a.monitorConnection()
	go a.consumeAck()
	a.conn <- UdpMsg{Conn: udpConn}
	go a.listen(udpConn) // Terminates when the connection is closed

	return &a, nil
}

// Send implements the core.Adapter interface. Not implemented for the udp adapter.
func (a *Adapter) Send(p core.Packet, r ...core.Recipient) ([]byte, error) {
	return nil, errors.New(ErrNotSupported, "Send not supported on udp adapter")
}

// Next implements the core.Adapter interface
func (a *Adapter) Next() ([]byte, core.AckNacker, error) {
	msg := <-a.next
	return msg.Data, udpAckNacker{
		Addr: msg.Addr,
	}, nil
}

// NextRegistration implements the core.Adapter interface
func (a *Adapter) NextRegistration() (core.Registration, core.AckNacker, error) {
	return udpRegistration{}, nil, errors.New(ErrNotSupported, "NextRegistration not supported on udp adapter")
}

// listen Handle incoming packets and forward them. Runs in its own goroutine.
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
		chresp := make(chan HandlerMsg)
		go a.HandleDatagram(buf[:n], chresp)
		a.handleResp(addr, chresp)
	}
}

// monitorConnection manages udpConnection of the adapter and send message through that connection
//
// That function executes into a single goroutine and is the only one allowed to write UDP messages.
// Doing this makes sure that only 1 goroutine is interacting with the connection. It thereby allows
// the connection to be replaced at any moment (in case of failure for instance) without disturbing
// the ongoing process.
func (a *Adapter) monitorConnection() {
	var udpConn *net.UDPConn
	for msg := range a.conn {
		if msg.Conn != nil { // Change the connection
			if udpConn != nil {
				a.ctx.Debug("Define new UDP connection")
				udpConn.Close()
			}
			udpConn = msg.Conn
		}

		if udpConn != nil && msg.Data != nil { // Send the given udp message
			if _, err := udpConn.WriteToUDP(msg.Data, msg.Addr); err != nil {
				a.ctx.WithError(err).Error("Error while sending UDP message")
			}
		}
	}
	if udpConn != nil {
		udpConn.Close() // Make sure we close the connection before leaving if we dare ever leave.
	}
}

// handleResp consumes message from chresp and forward them to the adapter via the Out or Udp
// channel.
//
// The function is called each time a chan HandlerMsg is created (meaning that we need to handle an
// uplink or a response) to handle the response(s) coming from the message handler.
func (a *Adapter) handleResp(addr *net.UDPAddr, chresp <-chan HandlerMsg) error {
	for msg := range chresp {
		switch msg.Type {
		case HANDLER_RESP:
			a.conn <- UdpMsg{
				Data: msg.Data,
				Addr: addr,
			}
		case HANDLER_OUT:
			a.next <- OutMsg{
				Data: msg.Data,
				Addr: addr,
			}
		case HANDLER_ERROR:
			err := fmt.Errorf(string(msg.Data))
			a.ctx.WithError(err).Error("Unable to handle response")
			return errors.New(ErrFailedOperation, err)
		default:
			err := errors.New(ErrFailedOperation, "Internal unexpected error while handling response")
			a.ctx.Error(err.Error())
			return err
		}
	}
	return nil
}

// consumeAck consumes messages from the Ack channel and forward them to handleResp
//
// The function is launched in its own goroutine and run concurrently with other consumers of the
// adapter. It basically pipes acknowledgement to the right channels after having derouted the
// processing to the handler.
func (a *Adapter) consumeAck() {
	for msg := range a.ack {
		chresp := make(chan HandlerMsg)
		switch msg.Type {
		case AN_ACK:
			go a.HandleAck(msg.Packet, chresp)
		case AN_NACK:
			go a.HandleNack(chresp)
		}
		err := a.handleResp(msg.Addr, chresp)
		msg.Cherr <- err
		close(msg.Cherr)
	}
}

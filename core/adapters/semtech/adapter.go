// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package semtech

import (
	"fmt"
	"net"

	"github.com/TheThingsNetwork/ttn/core"
	. "github.com/TheThingsNetwork/ttn/core/errors"
	"github.com/TheThingsNetwork/ttn/semtech"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	"github.com/TheThingsNetwork/ttn/utils/stats"
	"github.com/apex/log"
)

// Adapter represents a semtech adapter which sends and receives packet via UDP in respect of the
// semtech forwarder protocol
type Adapter struct {
	ctx  log.Interface // Just a logger
	conn chan udpMsg   // Channel used to manage response transmissions made by multiple goroutines
	next chan rxpkMsg  // Incoming valid RXPK packets are pushed to this channel
}

// udpMsg type materializes response messages transmitted towards existing recipients (commonly,
// gateways).
type udpMsg struct {
	conn *net.UDPConn // Provide if you intent to change the current adapter connection
	addr *net.UDPAddr // The target recipient address
	raw  []byte       // The raw byte sequence that has to be sent
}

// rxpkMsg type materializes valid uplink messages coming from a given recipient
type rxpkMsg struct {
	rxpk      semtech.RXPK   // The actual RXPK message
	recipient core.Recipient // The address and id of the source emitter
}

// NewAdapter constructs and allocates a new semtech adapter
func NewAdapter(port uint, ctx log.Interface) (*Adapter, error) {
	a := Adapter{
		ctx:  ctx,
		conn: make(chan udpMsg),
		next: make(chan rxpkMsg),
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
	a.conn <- udpMsg{conn: udpConn}
	go a.listen(udpConn) // Terminates when the connection is closed

	return &a, nil
}

// Send implements the core.Adapter interface. Not implemented for the semtech adapter.
func (a *Adapter) Send(p core.Packet, r ...core.Recipient) (core.Packet, error) {
	return core.Packet{}, errors.New(ErrNotSupported, "Send not supported on semtech adapter")
}

// Next implements the core.Adapter interface
func (a *Adapter) Next() (core.Packet, core.AckNacker, error) {
	msg := <-a.next
	packet, err := core.ConvertRXPK(msg.rxpk)
	if err != nil {
		a.ctx.Debug("Received invalid packet")
		return core.Packet{}, nil, errors.New(ErrInvalidStructure, err)
	}
	return packet, semtechAckNacker{recipient: msg.recipient, conn: a.conn}, nil
}

// NextRegistration implements the core.Adapter interface
func (a *Adapter) NextRegistration() (core.Registration, core.AckNacker, error) {
	return core.Registration{}, nil, errors.New(ErrNotSupported, "NextRegistration not supported on semtech adapter")
}

// listen Handle incoming packets and forward them
func (a *Adapter) listen(conn *net.UDPConn) {
	defer conn.Close()
	a.ctx.WithField("address", conn.LocalAddr()).Debug("Starting accept loop")
	for {
		buf := make([]byte, 512)
		n, addr, err := conn.ReadFromUDP(buf)
		if err != nil { // Problem with the connection
			a.ctx.WithError(err).Error("Connection error")
			continue
		}
		a.ctx.Debug("Incoming datagram")

		pkt := new(semtech.Packet)
		err = pkt.UnmarshalBinary(buf[:n])
		if err != nil {
			a.ctx.WithError(err).Warn("Invalid packet")
			continue
		}

		switch pkt.Identifier {
		case semtech.PULL_DATA: // PULL_DATA -> Respond to the recipient with an ACK
			stats.MarkMeter("semtech_adapter.pull_data")
			pullAck, err := semtech.Packet{
				Version:    semtech.VERSION,
				Token:      pkt.Token,
				Identifier: semtech.PULL_ACK,
			}.MarshalBinary()
			if err != nil {
				a.ctx.WithError(err).Error("Unexpected error while marshaling PULL_ACK")
				continue
			}
			a.ctx.WithField("recipient", addr).Debug("Sending PULL_ACK")
			a.conn <- udpMsg{addr: addr, raw: pullAck}
		case semtech.PUSH_DATA: // PUSH_DATA -> Transfer all RXPK to the component
			stats.MarkMeter("semtech_adapter.push_data")
			pushAck, err := semtech.Packet{
				Version:    semtech.VERSION,
				Token:      pkt.Token,
				Identifier: semtech.PUSH_ACK,
			}.MarshalBinary()
			if err != nil {
				a.ctx.WithError(err).Error("Unexpected error while marshaling PUSH_ACK")
				continue
			}
			a.ctx.WithField("Recipient", addr).Debug("Sending PUSH_ACK")
			a.conn <- udpMsg{addr: addr, raw: pushAck}

			if pkt.Payload == nil {
				a.ctx.WithField("packet", pkt).Warn("Invalid PUSH_DATA packet")
				continue
			}
			for _, rxpk := range pkt.Payload.RXPK {
				a.next <- rxpkMsg{
					rxpk:      rxpk,
					recipient: core.Recipient{Address: addr, Id: pkt.GatewayId},
				}
			}
		default:
			a.ctx.WithField("packet", pkt).Debug("Ignoring unexpected packet")
			continue
		}
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
		if msg.conn != nil { // Change the connection
			if udpConn != nil {
				a.ctx.Debug("Define new UDP connection")
				udpConn.Close()
			}
			udpConn = msg.conn
		}

		if udpConn != nil && msg.raw != nil { // Send the given udp message
			if _, err := udpConn.WriteToUDP(msg.raw, msg.addr); err != nil {
				a.ctx.WithError(err).Error("Error while sending UDP message")
			}
		}
	}
	if udpConn != nil {
		udpConn.Close() // Make sure we close the connection before leaving if we dare ever leave.
	}
}

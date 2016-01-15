// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package semtech

import (
	"fmt"
	"net"

	"github.com/TheThingsNetwork/ttn/core"
	"github.com/TheThingsNetwork/ttn/semtech"
	"github.com/apex/log"
)

type Adapter struct {
	Ctx  log.Interface
	conn chan udpMsg
	next chan rxpkMsg
}

type udpMsg struct {
	conn *net.UDPConn // Provide if you intent to change the current adapter connection
	addr *net.UDPAddr // The target recipient address
	raw  []byte       // The raw byte sequence that has to be sent
}

type rxpkMsg struct {
	rxpk      semtech.RXPK
	recipient core.Recipient
}

var ErrInvalidPort error = fmt.Errorf("Invalid port supplied. The connection might be already taken")
var ErrNotInitialized error = fmt.Errorf("Illegal call on non-initialized adapter")
var ErrNotSupported error = fmt.Errorf("Unsupported operation")
var ErrInvalidPacket error = fmt.Errorf("Invalid packet supplied")

// New constructs and allocates a new udp_sender adapter
func NewAdapter(port uint, ctx log.Interface) (*Adapter, error) {
	a := Adapter{
		Ctx:  ctx,
		conn: make(chan udpMsg),
		next: make(chan rxpkMsg),
	}

	// Create the udp connection and start listening with a goroutine
	var udpConn *net.UDPConn
	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("0.0.0.0:%d", port))
	a.Ctx.WithField("port", port).Info("Starting Server")
	if udpConn, err = net.ListenUDP("udp", addr); err != nil {
		a.Ctx.WithError(err).Error("Unable to start server")
		return nil, ErrInvalidPort
	}

	go a.monitorConnection()
	a.conn <- udpMsg{conn: udpConn}
	go a.listen(udpConn) // Terminates when the connection is closed

	return &a, nil
}

// ok controls whether or not the adapter has been initialized via NewAdapter()
func (a *Adapter) ok() bool {
	return a != nil && a.conn != nil && a.next != nil
}

// Send implements the core.Adapter interface
func (a *Adapter) Send(p core.Packet, r ...core.Recipient) (core.Packet, error) {
	return core.Packet{}, ErrNotSupported
}

// Next implements the core.Adapter interface
func (a *Adapter) Next() (core.Packet, core.AckNacker, error) {
	if !a.ok() {
		return core.Packet{}, nil, ErrNotInitialized
	}
	msg := <-a.next
	packet, err := core.ConvertRXPK(msg.rxpk)
	if err != nil {
		a.Ctx.Debug("Received invalid packet")
		return core.Packet{}, nil, ErrInvalidPacket
	}
	return packet, semtechAckNacker{recipient: msg.recipient, conn: a.conn}, nil
}

// NextRegistration implements the core.Adapter interface
func (a *Adapter) NextRegistration() (core.Registration, core.AckNacker, error) {
	return core.Registration{}, nil, ErrNotSupported
}

// listen Handle incoming packets and forward them
func (a *Adapter) listen(conn *net.UDPConn) {
	defer conn.Close()
	a.Ctx.WithField("address", conn.LocalAddr()).Debug("Starting accept loop")
	for {
		buf := make([]byte, 128)
		n, addr, err := conn.ReadFromUDP(buf)
		if err != nil { // Problem with the connection
			a.Ctx.WithError(err).Error("Connection error")
			continue
		}
		a.Ctx.WithField("datagram", buf[:n]).Debug("Incoming datagram")

		pkt := new(semtech.Packet)
		err = pkt.UnmarshalBinary(buf[:n])
		if err != nil {
			a.Ctx.WithError(err).Warn("Invalid packet")
			continue
		}

		switch pkt.Identifier {
		case semtech.PULL_DATA: // PULL_DATA -> Respond to the recipient with an ACK
			pullAck, err := semtech.Packet{
				Version:    semtech.VERSION,
				Token:      pkt.Token,
				Identifier: semtech.PULL_ACK,
			}.MarshalBinary()
			if err != nil {
				a.Ctx.WithError(err).Error("Unexpected error while marshaling PULL_ACK")
				continue
			}
			a.Ctx.WithField("recipient", addr).Debug("Sending PULL_ACK")
			a.conn <- udpMsg{addr: addr, raw: pullAck}
		case semtech.PUSH_DATA: // PUSH_DATA -> Transfer all RXPK to the component
			pushAck, err := semtech.Packet{
				Version:    semtech.VERSION,
				Token:      pkt.Token,
				Identifier: semtech.PUSH_ACK,
			}.MarshalBinary()
			if err != nil {
				a.Ctx.WithError(err).Error("Unexpected error while marshaling PUSH_ACK")
				continue
			}
			a.Ctx.WithField("Recipient", addr).Debug("Sending PUSH_ACK")
			a.conn <- udpMsg{addr: addr, raw: pushAck}

			if pkt.Payload == nil {
				a.Ctx.WithField("packet", pkt).Warn("Invalid PUSH_DATA packet")
				continue
			}
			for _, rxpk := range pkt.Payload.RXPK {
				a.next <- rxpkMsg{
					rxpk:      rxpk,
					recipient: core.Recipient{Address: addr, Id: pkt.GatewayId},
				}
			}
		default:
			a.Ctx.WithField("packet", pkt).Debug("Ignoring unexpected packet")
			continue
		}
	}
}

// monitorConnection manages udpConnection of the adapter and send message through that connection
func (a *Adapter) monitorConnection() {
	var udpConn *net.UDPConn
	for msg := range a.conn {
		if msg.conn != nil { // Change the connection
			if udpConn != nil {
				a.Ctx.Debug("Define new UDP connection")
				udpConn.Close()
			}
			udpConn = msg.conn
		}

		if udpConn != nil && msg.raw != nil { // Send the given udp message
			if _, err := udpConn.WriteToUDP(msg.raw, msg.addr); err != nil {
				a.Ctx.WithError(err).Error("Error while sending UDP message")
			}
		}
	}
	if udpConn != nil {
		udpConn.Close() // Make sure we close the connection before leaving if we dare ever leave.
	}
}

// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package udp

import (
	"net"

	core "github.com/TheThingsNetwork/ttn/refactor"
)

// udpAckNacker represents an AckNacker for a udp adapter
type udpAckNacker struct {
	Chack chan<- AckMsg
	Addr  *net.UDPAddr // The actual udp address related to that
}

// Ack implements the core.Adapter interface
func (an udpAckNacker) Ack(p *core.Packet) error {
	cherr := make(chan error)
	an.Chack <- AckMsg{Type: AN_ACK, Addr: an.Addr, Packet: p, Cherr: cherr}
	return <-cherr
}

// Ack implements the core.Adapter interface
func (an udpAckNacker) Nack() error {
	cherr := make(chan error)
	an.Chack <- AckMsg{Type: AN_NACK, Addr: an.Addr, Packet: nil}
	return <-cherr
}

// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package gateway

import (
	"github.com/thethingsnetwork/core/lorawan/semtech"
	"io"
)

type Forwarder struct {
	Id      [8]byte              // Gateway's Identifier
	alti    int                  // GPS altitude in RX meters
	ackr    uint                 // Number of upstream datagrams that were acknowledged
	dwnb    uint                 // Number of downlink datagrams received
	lati    float64              // GPS latitude, North is +
	long    float64              // GPS longitude, East is +
	rxfw    uint                 // Number of radio packets forwarded
	rxnb    uint                 // Number of radio packets received
	txnb    uint                 // Number of packets emitted
	routers []io.ReadWriteCloser // List of routers addresses
}

// NewForwarder create a forwarder instance bound to a set of routers.
func NewForwarder(id [8]byte, routers ...io.ReadWriteCloser) (*Forwarder, error) {
	return nil, nil
}

// Forward dispatch a packet to all connected routers.
func (fwd *Forwarder) Forward(packet semtech.Packet) error {
	return nil
}

// Flush spits out all downlink packet received by the forwarder since the last flush.
func (fwd *Forwarder) Flush() []semtech.Packet {
	return nil
}

// Stats computes and return the forwarder statistics since it was created
func (fwd Forwarder) Stats() semtech.Stat {
	return semtech.Stat{}
}

// Stop terminate the forwarder activity. Closing all routers connections
func (fwd *Forwarder) Stop() error {
	return nil
}

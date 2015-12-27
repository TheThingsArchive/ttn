// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package gateway

import (
	"fmt"
	"github.com/thethingsnetwork/core/lorawan/semtech"
	"github.com/thethingsnetwork/core/utils/pointer"
	"io"
	"time"
)

type Forwarder struct {
	Id       [8]byte              // Gateway's Identifier
	alti     int                  // GPS altitude in RX meters
	ackr     uint                 // Number of upstream datagrams that were acknowledged
	dwnb     uint                 // Number of downlink datagrams received
	lati     float64              // GPS latitude, North is +
	long     float64              // GPS longitude, East is +
	rxfw     uint                 // Number of radio packets forwarded
	rxnb     uint                 // Number of radio packets received
	txnb     uint                 // Number of packets emitted
	adapters []io.ReadWriteCloser // List of downlink adapters
}

// NewForwarder create a forwarder instance bound to a set of routers.
func NewForwarder(id [8]byte, adapters ...io.ReadWriteCloser) (*Forwarder, error) {
	if len(adapters) == 0 {
		return nil, fmt.Errorf("At least one adapter must be supplied")
	}
	return &Forwarder{
		Id:       id,
		alti:     120,
		lati:     53.3702,
		long:     4.8952,
		adapters: adapters,
	}, nil
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
	var ackr float64
	if fwd.txnb != 0 {
		ackr = float64(fwd.ackr) / float64(fwd.txnb)
	}

	return semtech.Stat{
		Ackr: &ackr,
		Alti: pointer.Int(fwd.alti),
		Dwnb: pointer.Uint(fwd.dwnb),
		Lati: pointer.Float64(fwd.lati),
		Long: pointer.Float64(fwd.long),
		Rxfw: pointer.Uint(fwd.rxfw),
		Rxnb: pointer.Uint(fwd.rxnb),
		Rxok: pointer.Uint(fwd.rxnb),
		Time: pointer.Time(time.Now()),
		Txnb: pointer.Uint(fwd.txnb),
	}
}

// Stop terminate the forwarder activity. Closing all routers connections
func (fwd *Forwarder) Stop() error {
	var errors []error
	for _, adapter := range fwd.adapters {
		err := adapter.Close()
		if err != nil {
			errors = append(errors, err)
		}
	}
	if len(errors) > 0 {
		return fmt.Errorf("Unable to stop the forwarder: %+v", errors)
	}
	return nil
}

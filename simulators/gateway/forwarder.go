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
	upnb     uint                 // Number of upstream datagrams sent
	ackn     uint                 // Number of upstream datagrams that were acknowledged
	dwnb     uint                 // Number of downlink datagrams received
	lati     float64              // GPS latitude, North is +
	long     float64              // GPS longitude, East is +
	rxfw     uint                 // Number of radio packets forwarded
	rxnb     uint                 // Number of radio packets received
	txnb     uint                 // Number of packets emitted
	adapters []io.ReadWriteCloser // List of downlink adapters
	packets  []semtech.Packet     // Downlink packets received
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
	if packet.Identifier != semtech.PUSH_DATA {
		return fmt.Errorf("Unable to forward with identifier %x", packet.Identifier)
	}

	raw, err := semtech.Marshal(packet)
	if err != nil {
		return err
	}

	for _, adapter := range fwd.adapters {
		n, err := adapter.Write(raw)
		if err != nil {
			return err
		}
		if n < len(raw) {
			return fmt.Errorf("Packet was too long")
		}
	}

	return nil
}

// Flush spits out all downlink packet received by the forwarder since the last flush.
func (fwd *Forwarder) Flush() []semtech.Packet {
	return nil
}

// Stats computes and return the forwarder statistics since it was created
func (fwd Forwarder) Stats() semtech.Stat {
	var ackr float64
	if fwd.upnb != 0 {
		ackr = float64(fwd.ackn) / float64(fwd.upnb)
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

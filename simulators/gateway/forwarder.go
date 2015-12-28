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
	adapters []io.ReadWriteCloser // List of downlink adapters
	packets  []semtech.Packet     // Downlink packets received
	done     chan chan error      // Done channel
	commands chan command         // Concurrent access on gateway stats
}

type commandName string
type command struct {
	name commandName
	data interface{}
}

const (
	cmd_ACK     commandName = "Acknowledged"
	cmd_EMIT    commandName = "Emitted"
	cmd_RECVUP  commandName = "Radio Packet Received"
	cmd_RECVDWN commandName = "Dowlink Datagram Received"
	cmd_FWD     commandName = "Forwarded"
	cmd_FLUSH   commandName = "Flush"
	cmd_STATS   commandName = "Stats"
)

// NewForwarder create a forwarder instance bound to a set of routers.
func NewForwarder(id [8]byte, adapters ...io.ReadWriteCloser) (*Forwarder, error) {
	if len(adapters) == 0 {
		return nil, fmt.Errorf("At least one adapter must be supplied")
	}

	fwd := &Forwarder{
		Id:       id,
		alti:     120,
		lati:     53.3702,
		long:     4.8952,
		adapters: adapters,
		done:     make(chan chan error, len(adapters)),
		commands: make(chan command),
	}

	go fwd.handleCommands()
	go fwd.listen()

	return fwd, nil
}

// listen get downlink packets from routers and store them until a flush is requested
func (fwd *Forwarder) listen() {
	dwnl := make(chan semtech.Packet, len(fwd.adapters))

	// Star listening to each adapter Read() method
	for _, adapter := range fwd.adapters {
		go asChannel(adapter, dwnl)
	}

	for {
		select {
		case fwd.commands <- command{cmd_RECVDWN, <-dwnl}:

		case errc := <-fwd.done:
			// Empty the buffer first to avoid leaking goroutines
			nb := len(dwnl)
			for i := 0; i < nb; i += 1 {
				fwd.commands <- command{cmd_RECVDWN, <-dwnl}
			}

			// Then stop
			errc <- nil
			return
		}
	}
}

// asChannel listen to incoming connection from an adapter and forward them into a dedicated
// channel. Non-valid packets are ignored.
func asChannel(adapter io.ReadWriteCloser, dwnl chan<- semtech.Packet) {
	for {
		buf := make([]byte, 1024)
		n, err := adapter.Read(buf)
		if err != nil {
			fmt.Println(err)
			return // Error on reading, we assume the connection is closed / lost
		}
		packet, err := semtech.Unmarshal(buf[:n])
		if err != nil {
			fmt.Println(err)
			continue
		}
		if packet.Identifier != semtech.PUSH_DATA {
			continue
		}
		dwnl <- *packet // Only valid PUSH_DATA packet are transmitted through the chan
	}
}

// handleCommands acts as a mediator between all goroutines that attempt to modify the forwarder
// attributes. All sensitive operations are done by commands send though an appropriate channel.
// This method consume commands from the channel until it's closed.
func (fwd *Forwarder) handleCommands() {
	for cmd := range fwd.commands {
		switch cmd.name {
		case cmd_ACK:
			fwd.ackn += 1
		case cmd_FWD:
			fwd.rxfw += 1
		case cmd_EMIT:
			fwd.upnb += 1
		case cmd_RECVUP:
			fwd.rxnb += 1
		case cmd_RECVDWN:
			fwd.dwnb += 1
			fwd.packets = append(fwd.packets, cmd.data.(semtech.Packet))
		case cmd_FLUSH:
			cmd.data.(chan []semtech.Packet) <- fwd.packets
			fwd.packets = make([]semtech.Packet, 0)
		case cmd_STATS:
			var ackr float64
			if fwd.upnb != 0 {
				ackr = float64(fwd.ackn) / float64(fwd.upnb)
			}

			cmd.data.(chan semtech.Stat) <- semtech.Stat{
				Ackr: &ackr,
				Alti: pointer.Int(fwd.alti),
				Dwnb: pointer.Uint(fwd.dwnb),
				Lati: pointer.Float64(fwd.lati),
				Long: pointer.Float64(fwd.long),
				Rxfw: pointer.Uint(fwd.rxfw),
				Rxnb: pointer.Uint(fwd.rxnb),
				Rxok: pointer.Uint(fwd.rxnb),
				Time: pointer.Time(time.Now()),
				Txnb: pointer.Uint(0),
			}
		}
	}
}

// Forward dispatch a packet to all connected routers.
func (fwd *Forwarder) Forward(packet semtech.Packet) error {
	fwd.commands <- command{cmd_RECVUP, nil}
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
		fwd.commands <- command{cmd_EMIT, nil}
	}

	fwd.commands <- command{cmd_FWD, nil}
	return nil
}

// Flush spits out all downlink packet received by the forwarder since the last flush.
func (fwd *Forwarder) Flush() []semtech.Packet {
	chpkt := make(chan []semtech.Packet)
	fwd.commands <- command{cmd_FLUSH, chpkt}
	return <-chpkt
}

// Stats computes and return the forwarder statistics since it was created
func (fwd Forwarder) Stats() semtech.Stat {
	chstats := make(chan semtech.Stat)
	fwd.commands <- command{cmd_STATS, chstats}
	return <-chstats
}

// Stop terminate the forwarder activity. Closing all routers connections
func (fwd *Forwarder) Stop() error {
	var errors []error

	// Close the uplink adapters
	for _, adapter := range fwd.adapters {
		err := adapter.Close()
		if err != nil {
			errors = append(errors, err)
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("Unable to stop the forwarder: %+v", errors)
	}

	// Stop listening to downlink packets
	errc := make(chan error)
	fwd.done <- errc
	if err := <-errc; err != nil {
		return err
	}

	// Close the commands channel
	close(fwd.commands)

	return nil
}

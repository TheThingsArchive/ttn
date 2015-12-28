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
	commands chan command         // Concurrent access on gateway stats
	Errors   chan error           // Done channel
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
		commands: make(chan command),
		Errors:   make(chan error, len(adapters)),
	}

	go fwd.handleCommands()

	// Star listening to each adapter Read() method
	for _, adapter := range fwd.adapters {
		go fwd.listenAdapter(adapter)
	}

	return fwd, nil
}

// listenAdapter listen to incoming datagrams from an adapter. Non-valid packets are ignored.
func (fwd Forwarder) listenAdapter(adapter io.ReadWriteCloser) {
	acks := make(map[[3]byte]uint) // adapterIndex | packet.Identifier | packet.Token
	for {
		buf := make([]byte, 1024)
		fmt.Printf("Forwarder listens to downlink datagrams\n")
		n, err := adapter.Read(buf)
		if err != nil {
			fmt.Println(err)
			fwd.Errors <- err
			return // Error on reading, we assume the connection is closed / lost
		}
		fmt.Printf("Forwarder unmarshals datagram %x\n", buf[:n])
		packet, err := semtech.Unmarshal(buf[:n])
		if err != nil {
			fmt.Println(err)
			continue
		}

		token := [3]byte{packet.Identifier, packet.Token[0], packet.Token[1]}
		switch packet.Identifier {
		case semtech.PUSH_ACK, semtech.PULL_ACK:
			if acks[token] > 0 {
				acks[token] -= 1
				fwd.commands <- command{cmd_ACK, nil}
			}
		case semtech.PULL_RESP:
			fwd.commands <- command{cmd_RECVDWN, packet}
		default:
			fmt.Printf("Forwarder ignores contingent packet %+v\n", packet)
		}

	}
}

// handleCommands acts as a mediator between all goroutines that attempt to modify the forwarder
// attributes. All sensitive operations are done by commands send though an appropriate channel.
// This method consume commands from the channel until it's closed.
func (fwd *Forwarder) handleCommands() {
	for cmd := range fwd.commands {
		fmt.Printf("Fowarder executes command: %v\n", cmd.name)
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
func (fwd Forwarder) Forward(packet semtech.Packet) error {
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
func (fwd Forwarder) Flush() []semtech.Packet {
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
func (fwd Forwarder) Stop() error {
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

	return nil
}

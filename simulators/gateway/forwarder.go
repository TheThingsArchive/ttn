// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package gateway

import (
	"fmt"
	"io"
	"time"

	"github.com/TheThingsNetwork/ttn/semtech"
	"github.com/TheThingsNetwork/ttn/utils/pointer"
	"github.com/apex/log"
)

type Forwarder struct {
	ctx      log.Interface
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
	acks     map[[4]byte]uint     // adapterIndex | packet.Identifier | packet.Token
	commands chan command         // Concurrent access on gateway stats
	quit     chan error           // Adapter which loses connection spit here
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

var statTimer <-chan time.Time

// NewForwarder create a forwarder instance bound to a set of routers.
func NewForwarder(id [8]byte, ctx log.Interface, adapters ...io.ReadWriteCloser) (*Forwarder, error) {
	if len(adapters) == 0 {
		return nil, fmt.Errorf("At least one adapter must be supplied")
	}

	if len(adapters) > 255 { // cf fwd.acks
		return nil, fmt.Errorf("Cannot connect more than 255 adapters")
	}

	fwd := &Forwarder{
		Id:       id,
		ctx:      ctx,
		alti:     120,
		lati:     53.3702,
		long:     4.8952,
		adapters: adapters,
		packets:  make([]semtech.Packet, 0),
		acks:     make(map[[4]byte]uint),
		commands: make(chan command),
		quit:     make(chan error, len(adapters)),
	}

	go fwd.handleCommands()

	statTimer = time.Tick(5 * time.Second)

	// Star listening to each adapter Read() method
	for i, adapter := range fwd.adapters {
		go fwd.listenAdapter(adapter, i)
	}

	return fwd, nil
}

// listenAdapter listen to incoming datagrams from an adapter. Non-valid packets are ignored.
func (fwd Forwarder) listenAdapter(adapter io.ReadWriteCloser, index int) {
	for {
		buf := make([]byte, 1024)
		n, err := adapter.Read(buf)
		fwd.ctx.WithField("nb bytes", n).Debug("bytes received by adapter")
		if err != nil {
			fwd.ctx.WithError(err).Error("Connection lost / closed")
			fwd.quit <- err
			return // Connection lost / closed
		}
		fwd.ctx.WithField("datagram", buf[:n]).Debug("unmarshalling datagram")
		packet := new(semtech.Packet)
		if err = packet.UnmarshalBinary(buf[:n]); err != nil {
			fwd.ctx.WithError(err).Warn("Unable to unmarshal datagram to packet")
			continue
		}

		switch packet.Identifier {
		case semtech.PUSH_ACK, semtech.PULL_ACK:
			fwd.commands <- command{cmd_ACK, ackToken(index, *packet)}
		case semtech.PULL_RESP:
			fwd.commands <- command{cmd_RECVDWN, packet}
		default:
			fwd.ctx.WithField("packet", packet).Warn("Ignoring contingent packet")
		}
	}
}

// handleCommands acts as a monitor between all goroutines that attempt to modify the forwarder
// attributes. All sensitive operations are done by commands sent through an appropriate channel.
// This method consumes commands from the channel until it's closed.
func (fwd *Forwarder) handleCommands() {
	for cmd := range fwd.commands {
		fwd.ctx.WithField("command", cmd.name).Debug("executing command")

		switch cmd.name {
		case cmd_ACK:
			token := cmd.data.([4]byte)
			if fwd.acks[token] > 0 {
				fwd.acks[token] -= 1
				fwd.ackn += 1
			}
		case cmd_FWD:
			fwd.rxfw += 1
		case cmd_EMIT:
			token := cmd.data.([4]byte)
			fwd.acks[token] += 1
			fwd.upnb += 1
		case cmd_RECVUP:
			fwd.rxnb += 1
		case cmd_RECVDWN:
			fwd.dwnb += 1
			fwd.packets = append(fwd.packets, *cmd.data.(*semtech.Packet))
		case cmd_FLUSH:
			cmd.data.(chan []semtech.Packet) <- fwd.packets
			fwd.packets = make([]semtech.Packet, 0)
		case cmd_STATS:
			var ackr float64
			if fwd.upnb > 0 {
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
func (fwd Forwarder) Forward(rxpks ...semtech.RXPK) error {
	fwd.commands <- command{cmd_RECVUP, nil}

	packet := semtech.Packet{
		Version:    semtech.VERSION,
		Identifier: semtech.PUSH_DATA,
		GatewayId:  fwd.Id[:],
		Token:      genToken(),
		Payload:    &semtech.Payload{RXPK: rxpks},
	}

	select {
	case <-statTimer:
		stats := fwd.Stats()
		packet.Payload.Stat = &stats
	default:
	}

	raw, err := packet.MarshalBinary()
	if err != nil {
		return err
	}

	for i, adapter := range fwd.adapters {
		n, err := adapter.Write(raw)
		if err != nil {
			return err
		}
		if n < len(raw) {
			return fmt.Errorf("Packet was too long")
		}
		fwd.commands <- command{cmd_EMIT, ackToken(i, packet)}
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

	// Wait for each adapter to terminate
	for range fwd.adapters {
		<-fwd.quit
	}

	close(fwd.commands)

	if len(errors) > 0 {
		return fmt.Errorf("Unable to stop the forwarder: %+v", errors)
	}
	return nil
}

// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package gateway

import (
	"errors"
	"fmt"
	"github.com/thethingsnetwork/core/lorawan/semtech"
	"net"
	"strings"
	"time"
)

const (
	LISTENER_BUF_SIZE = 1024
	LISTENER_TIMEOUT  = 4 * time.Second
)

type Forwarder interface {
	Forward(packet semtech.Packet) error
	Start() (<-chan semtech.Packet, <-chan error, error)
	Stop() error
}

// Start a Gateway. This will create several udp connections - one per associated router.
// Then, incoming streams of packets are merged into a single one.
// Same for errors.
func (g *Gateway) Start() (<-chan semtech.Packet, <-chan error, error) {
	// Ensure not already started
	if g.quit != nil {
		return nil, nil, errors.New("Try to start a started gateway")
	}

	// Open all UDP connections
	connections := make([]*net.UDPConn, 0)
	var err error
	for _, addr := range g.routers {
		var conn *net.UDPConn
		conn, err = net.ListenUDP("udp", addr)
		if err != nil {
			break
		}
		connections = append(connections, conn)
	}

	// On error, close all opened UDP connection and leave
	if err != nil {
		for _, conn := range connections {
			conn.Close()
		}
		return nil, nil, err
	}

	// Create communication channels and launch goroutines to handle connections
	chout := make(chan semtech.Packet)
	cherr := make(chan error)
	quit := make(chan chan error)
	cmd := make(chan command)
	go reduceCmd(g, cmd)
	for _, conn := range connections {
		go listen(conn, chout, cherr, cmd, quit)
	}

	// Keep a reference to the quit channel, and return the others
	g.quit = quit
	g.cmd = cmd
	return chout, cherr, nil
}

// listen materialize the goroutine handling incoming packet from routers
func listen(conn *net.UDPConn, chout chan<- semtech.Packet, cherr chan<- error, cmd chan<- command, quit <-chan chan error) {
	connIn, connErr := asChannel(conn)
	errBuf := make([]error, 0)
	outBuf := make([]semtech.Packet, 0)
	for {
		var safeChout chan<- semtech.Packet
		var packet semtech.Packet
		if len(outBuf) > 0 {
			safeChout = chout
			packet = outBuf[0]
		}

		var safeCherr chan<- error
		var err error
		if len(errBuf) > 0 {
			safeCherr = cherr
			err = errBuf[0]
		}

		select {
		case ack := <-quit: // quit event, the gateway is stoppped
			e := conn.Close()
			ack <- e
			return
		case buf := <-connIn: // connIn event, a packet has been received by the listener goroutine
			cmd <- cmd_RECD_PACKET
			packet, err := semtech.Unmarshal(buf)
			if err != nil {
				errBuf = append(errBuf, err)
				continue
			}
			outBuf = append(outBuf, *packet)
		case safeChout <- packet: // emit an available packet to chout
			outBuf = outBuf[1:]
		case err := <-connErr: // connErr event, an error has been triggered by the listener goroutine
			errBuf = append(errBuf, err)
		case safeCherr <- err: // emit an existing error to cherr
			errBuf = errBuf[1:]
		}
	}
}

// as channel is actually a []byte generator that listen to a given udp connection.
// It is used to prevent the listen() function from blocking on ReadFromUDP() and then,
// still be available for a quit event that could come at any time.
// This function is thereby nothing more than a mapping of incoming connection to channels of
// communication.
func asChannel(conn *net.UDPConn) (<-chan []byte, <-chan error) {
	cherr := make(chan error)
	chout := make(chan []byte)
	go func() {
		buf := make([]byte, LISTENER_BUF_SIZE)
		for {
			n, _, err := conn.ReadFromUDP(buf)
			if err != nil {
				if strings.Contains(err.Error(), "use of closed network connection") {
					return
				}
				select {
				case cherr <- err:
					continue
				case <-time.After(LISTENER_TIMEOUT):
					close(cherr)
					close(chout)
					return
				}
			}
			select {
			case chout <- buf[:n]:
			case <-time.After(LISTENER_TIMEOUT):
				close(cherr)
				close(chout)
				return
			}
		}
	}()
	return chout, cherr
}

// reduceCmd handle all updates made on the gateway statistics
func reduceCmd(gateway *Gateway, commands <-chan command) {
	for command := range commands {
		switch command {
		case cmd_ACKN_PACKET:
			gateway.ackr += 1
		case cmd_EMIT_PACKET:
			gateway.txnb += 1
		case cmd_FORW_PACKET:
			gateway.rxfw += 1
		case cmd_RECU_PACKET:
			gateway.rxnb += 1
		case cmd_RECD_PACKET:
			gateway.dwnb += 1
		}
	}
}

// Stop remove all previously created connection.
func (g *Gateway) Stop() error {
	if g.quit == nil || g.cmd == nil {
		return errors.New("Try to stop a non-started gateway")
	}

	for range g.routers {
		errc := make(chan error)
		g.quit <- errc
		err := <-errc
		if err != nil {
			fmt.Printf("%+v\n", err)
		}
	}
	close(g.quit)
	close(g.cmd)
	g.routers = make([]*net.UDPAddr, 0)
	g.quit = nil
	return nil
}

// Forward transfer a packet to all known routers.
// It fails if the gateway hasn't been started beforehand.
func (g *Gateway) Forward(packet semtech.Packet) error {
	if g.quit == nil || g.cmd == nil {
		return errors.New("Unable to forward on a non-started gateway")
	}

	g.cmd <- cmd_RECU_PACKET

	connections := make([]*net.UDPConn, 0)
	var err error
	for _, addr := range g.routers {
		var conn *net.UDPConn
		conn, err = net.DialUDP("udp", nil, addr)
		if err != nil {
			break
		}
		defer conn.Close()
		connections = append(connections, conn)
	}
	raw, err := semtech.Marshal(packet)

	if err != nil {
		return errors.New(fmt.Sprintf("Unable to forward the packet. %v\n", err))
	}

	g.cmd <- cmd_FORW_PACKET

	for _, conn := range connections {
		_, err = conn.Write(raw)
		g.cmd <- cmd_EMIT_PACKET
	}

	if err != nil {
		return errors.New(fmt.Sprintf("Something went wrong during forwarding. %v\n", err))
	}

	return nil
}

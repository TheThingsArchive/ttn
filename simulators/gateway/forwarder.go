// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package gateway

import (
	"errors"
	"fmt"
	"github.com/thethingsnetwork/core/lorawan/semtech"
	"net"
	"strings"
)

const LISTENER_BUF_SIZE = 1024

type Forwarder interface {
	Forward(packet semtech.Packet)
	Start() (<-chan semtech.Packet, <-chan error, error)
	Stop() error
	Stat() semtech.Stat
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
	quit := make(chan bool, len(g.routers))
	for _, conn := range connections {
		go listen(conn, chout, cherr, quit)
	}

	// Keep a reference to the quit channel, and return the others
	g.quit = quit
	return chout, cherr, nil
}

// listen materialize the goroutine handling incoming packet from routers
func listen(conn *net.UDPConn, chout chan<- semtech.Packet, cherr chan<- error, quit chan bool) {
	connIn, connErr := asChannel(conn)
	for {
		select {
		case <-quit:
			close(connIn)
			close(connErr)
			conn.Close() // Any chance this would return an error ? :/
			return
		case buf := <-connIn:
			packet, err := semtech.Unmarshal(buf)
			if err != nil {
				cherr <- err
				continue
			}
			chout <- *packet
		case err := <-connErr:
			cherr <- err
		}
	}
}

func asChannel(conn *net.UDPConn) (chan []byte, chan error) {
	cherr := make(chan error)
	chout := make(chan []byte)
	go func() {
		defer func() {
			// The handling could be better here
			recover() // In case we're writing a close channel.
		}()
		buf := make([]byte, LISTENER_BUF_SIZE)
		for {
			n, _, err := conn.ReadFromUDP(buf)
			if err != nil {
				if strings.Contains(err.Error(), "use of closed network connection") {
					return
				}
				cherr <- err
				continue
			}
			chout <- buf[:n]
		}
	}()
	return chout, cherr
}

// Stop remove all previously created connection.
func (g *Gateway) Stop() error {
	if g.quit == nil {
		return errors.New("Try to stop a non-started gateway")
	}

	for range g.routers {
		g.quit <- true
	}
	close(g.quit)
	g.routers = make([]*net.UDPAddr, 0)
	g.quit = nil
	return nil
}

// Forward transfer a packet to all known routers.
// It panics if the gateway hasn't been started beforehand.
func (g *Gateway) Forward(packet semtech.Packet) {

}

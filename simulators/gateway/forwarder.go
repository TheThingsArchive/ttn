// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package gateway

import (
	"errors"
	"fmt"
	"github.com/thethingsnetwork/core/lorawan/semtech"
	"net"
)

type Forwarder interface {
	Forward(packet semtech.Packet)
	Start() (<-chan semtech.Packet, <-chan error)
	Stop() error
	Stat() semtech.Stat
}

func handleError(cherr chan error, err error) {
	fmt.Printf("\n[Gateway.Start()] %+v\n", err)
	if cherr != nil {
		cherr <- err
	}
}

// Start a Gateway. This will create several udp connections - one per associated router.
// Then, incoming streams of packets are merged into a single one.
// Same for errors.
func (g *Gateway) Start() (<-chan semtech.Packet, <-chan error) {
	if g.cherr != nil || g.chout != nil {
		panic("Try to start a started gateway")
	}

	g.chout = make(chan semtech.Packet)
	g.cherr = make(chan error)

	for router := range g.routers {
		addr, err := net.ResolveUDPAddr("udp", router)
		if err != nil {
			handleError(g.cherr, err)
			continue
		}

		go func() {
			conn, err := net.ListenUDP("udp", addr)
			if err != nil {
				handleError(g.cherr, err)
				return
			}
			g.routers[router] = conn
			defer conn.Close()
			buf := make([]byte, 1024)
			for g.routers[router] != nil {
				n, _, err := g.routers[router].ReadFromUDP(buf)
				if err != nil {
					handleError(g.cherr, err)
					continue
				}
				packet, err := semtech.Unmarshal(buf[:n])
				if err != nil {
					handleError(g.cherr, err)
					continue
				}
				if g.chout != nil {
					g.chout <- *packet
				}
			}
		}()
	}

	return g.chout, g.cherr
}

// Stop remove all previously created connection
func (g *Gateway) Stop() error {
	if g.cherr == nil || g.chout == nil {
		panic("Try to stop a non-started gateway")
	}

	errs := make([]error, 0)
	for router, conn := range g.routers {
		if conn != nil {
			if err := conn.Close(); err != nil {
				errs = append(errs, err)
				continue
			}
			g.routers[router] = nil
		}
	}
	if len(errs) != 0 {
		return errors.New("Unable to stop the gateway")
	}
	close(g.cherr)
	close(g.chout)
	g.cherr = nil
	g.chout = nil
	return nil
}

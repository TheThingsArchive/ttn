// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package gateway

import (
	"github.com/thethingsnetwork/core/lorawan/semtech"
	"net"
)

type Forwarder interface {
	Forward(packet semtech.Packet)
	Start() (<-chan semtech.Packet, <-chan error)
	Stat() semtech.Stat
}

func (g *Gateway) Start() (<-chan semtech.Packet, <-chan error) {
	chout := make(chan semtech.Packet)
	cherr := make(chan error)

	for _, router := range g.routers {
		addr, err := net.ResolveUDPAddr("udp", router)
		if err != nil {
			cherr <- err
			continue
		}

		go func() {
			conn, err := net.ListenUDP("udp", addr)
			if err != nil {
				cherr <- err
				return
			}
			defer conn.Close()
			buf := make([]byte, 1024)
			for {
				n, _, err := conn.ReadFromUDP(buf)
				if err != nil {
					cherr <- err
					continue
				}
				packet, err := semtech.Unmarshal(buf[:n])
				if err != nil {
					cherr <- err
					continue
				}
				chout <- *packet
			}
		}()
	}

	g.chout = chout
	g.cherr = cherr
	return chout, cherr
}

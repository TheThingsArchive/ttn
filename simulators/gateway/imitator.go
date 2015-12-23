// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package gateway

import (
	"errors"
	"fmt"
	"github.com/thethingsnetwork/core/lorawan/semtech"
	"net"
	"sync"
	"time"
)

type Imitator interface {
	Mimic() error
}

func (g *Gateway) Mimic() error {
	if g.IsRunning() {
		return errors.New("Cannot mimic on a started gateway")
	}

	chout, cherr, err := g.Start()

	if err != nil {
		return err
	}

	mutex := &sync.Mutex{}
	communications := make(map[*semtech.Packet][]*net.UDPAddr)

	go func(g *Gateway) {
		ticker := time.Tick(time.Millisecond * 800)
		for {
			<-ticker
			if !g.IsRunning() {
				return
			}
			token := genToken()
			packet := semtech.Packet{
				Version:    semtech.VERSION,
				Identifier: semtech.PUSH_DATA,
				GatewayId:  g.Id,
				Token:      token,
			}
			mutex.Lock()
			err := g.Forward(packet)
			if err != nil {
				fmt.Println(err)
			}
			communications[&packet] = g.routers
			mutex.Unlock()
		}
	}(g)

	go func(g *Gateway) {
		ticker := time.Tick(time.Millisecond * 800)
		for {
			<-ticker
			mutex.Lock()
			if !g.IsRunning() {
				mutex.Unlock()
				return
			}
			for packet, routers := range communications {
				err := g.Forward(*packet, routers...)
				if err != nil {
					fmt.Println(err)
				}
			}
			mutex.Unlock()
		}
	}(g)

	go func(g *Gateway) {
		for {
			if !g.IsRunning() {
				return
			}
			select {
			case packet := <-chout:
				fmt.Printf("%+v\n", packet)
			case err := <-cherr:
				fmt.Println(err)
			case <-time.After(time.Second):
			}
		}
	}(g)

	return nil
}

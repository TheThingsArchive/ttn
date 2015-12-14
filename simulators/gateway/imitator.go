// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package gateway

import (
	"errors"
	"github.com/thethingsnetwork/core/lorawan/semtech"
	"time"
)

type Imitator interface {
	Mimic() error
}

func (g *Gateway) Mimic() error {
	if g.cmd == nil || g.quit == nil {
		return errors.New("Cannot mimic on a stopped gateway")
	}

	go func(g *Gateway) {
		ticker := time.Tick(time.Millisecond * 800)
		for {
			select {
			case <-ticker:
				g.Forward(semtech.Packet{
					Version:    semtech.VERSION,
					Identifier: semtech.PUSH_DATA,
					GatewayId:  g.Id,
					Token:      genToken(),
				})
			case ack := <-g.quit:
				ack <- nil
				return
			}
		}
	}(g)

	return nil
}

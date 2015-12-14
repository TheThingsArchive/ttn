// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package gateway

import (
	. "github.com/smartystreets/goconvey/convey"
	"github.com/thethingsnetwork/core/lorawan/semtech"
	"testing"
	"time"
)

func TestMimic(t *testing.T) {
	gatewayId := []byte("MyGateway")[:8]
	routerAddr := "0.0.0.0:3000"
	gateway, e := New(gatewayId, routerAddr)

	if e != nil {
		t.Errorf("Unexpected error %v", e)
		return
	}

	chout, cherr, e := gateway.Start()

	if e != nil {
		t.Errorf("Unexpected error %v", e)
		return
	}

	defer gateway.Stop()

	Convey("Given a started gateway", t, func() {
		err := gateway.Mimic()
		Convey("The imitator mode can be started", func() {
			So(err, ShouldBeNil)
		})

		Convey("After having started the imitation", func() {
			Convey("It should periodically and randomly emit PUSH_DATA packet", func() {
				nb := 0
				for nb < 3 {
					select {
					case packet := <-chout:
						So(packet.Identifier, ShouldEqual, semtech.PUSH_DATA)
						nb += 1
					case err := <-cherr:
						t.Errorf("Unexpected error %v", err)
						return
					case <-time.After(time.Second * 5):
						t.Errorf("Timeout")
						return
					}
				}
			})

			Convey("It should periodically emit PULL_DATA after a PUSH_ACK has been received", func() {
				var packet semtech.Packet
				select {
				case packet = <-chout:
				case <-time.After(5 * time.Second):
					t.Errorf("Timeout")
					return
				}
				if packet.Identifier != semtech.PUSH_DATA {
					t.Errorf("Unexpected packet identifier")
					return
				}

				gateway.Forward(semtech.Packet{
					Version:    semtech.VERSION,
					Token:      packet.Token,
					Identifier: semtech.PUSH_ACK,
				})

				maxTries := 3
				for maxTries > 0 {
					select {
					case packet := <-chout:
						if packet.Identifier != semtech.PULL_DATA {
							maxTries -= 1
							continue
						}
						So(packet.Identifier, ShouldEqual, semtech.PULL_DATA)
					case err := <-cherr:
						t.Errorf("Unexpected error %v", err)
						return
					case <-time.After(time.Second * 5):
						t.Errorf("Timeout")
						return
					}
				}

				if maxTries <= 0 {
					So("No PULL_DATA sent", ShouldBeNil)
				}
			})
		})
	})
}

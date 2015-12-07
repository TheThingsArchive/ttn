// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package gateway

import (
	. "github.com/smartystreets/goconvey/convey"
	"github.com/thethingsnetwork/core/lorawan/semtech"
	"net"
	"testing"
)

func TestStart(t *testing.T) {
	gatewayId := "MyGateway"
	routerAddr := "0.0.0.0:3000"
	gateway, err := New(gatewayId, routerAddr)
	chout, cherr := gateway.Start()

	udpAddr, err := net.ResolveUDPAddr("udp", routerAddr)
	if err != nil {
		t.Errorf("Unexpected error %+v\n", err)
		return
	}

	conn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		t.Errorf("Unexpected error %+v\n", err)
		return
	}

	Convey("Given two valid router adresses", t, func() {
		Convey("After having created a new gateway with one router", func() {
			Convey("Both channels should exist", func() {
				So(cherr, ShouldNotBeNil)
				So(chout, ShouldNotBeNil)
			})

			Convey("A valid packet should be forwarded", func() {
				packet := semtech.Packet{
					Version:    semtech.VERSION,
					Token:      []byte{0x1, 0x2},
					Identifier: semtech.PUSH_ACK,
				}
				raw, err := semtech.Marshal(&packet)
				if err != nil {
					t.Errorf("Unexpected error %+v\n", err)
					return
				}
				conn.Write(raw)
				So(<-chout, ShouldResemble, packet)
			})

			Convey("An invalid packet should raise an error", func() {
				conn.Write([]byte("Invalid"))
				So(<-cherr, ShouldNotBeNil)
			})
		})
	})
}

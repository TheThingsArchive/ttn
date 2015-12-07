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
	gateway, _ := New(gatewayId, routerAddr)
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

	Convey("Given a valid started gateway instance bound to a router", t, func() {
		Convey("Both channels should exist", func() {
			So(cherr, ShouldNotBeNil)
			So(chout, ShouldNotBeNil)
		})

		Convey("A connection should exist", func() {
			So(gateway.routers[routerAddr], ShouldNotBeNil)
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

		Convey("It should panic if started one more time", func() {
			So(func() {
				gateway.Start()
			}, ShouldPanic)
		})
	})

}

func TestStop(t *testing.T) {
	gatewayId := "MyGateway"
	routerAddr := "0.0.0.0:3000"
	gateway, _ := New(gatewayId, routerAddr)
	Convey("Given a gateway instance", t, func() {
		Convey("It should panic if stopped while not started", func() {
			So(func() { gateway.Stop() }, ShouldPanic)
		})

		Convey("It should stop correctly after having started", func() {
			gateway.Start()
			err := gateway.Stop()

			So(err, ShouldBeNil)
			So(gateway.cherr, ShouldBeNil)
			So(gateway.chout, ShouldBeNil)
			So(gateway.routers[routerAddr], ShouldBeNil)
		})
	})
}

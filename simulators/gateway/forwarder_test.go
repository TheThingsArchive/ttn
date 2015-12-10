// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package gateway

import (
	. "github.com/smartystreets/goconvey/convey"
	"github.com/thethingsnetwork/core/lorawan/semtech"
	"net"
	"testing"
	"time"
)

func TestStart(t *testing.T) {
	gatewayId := []byte("MyGateway")[:8]
	routerAddr := "0.0.0.0:3000"
	gateway, _ := New(gatewayId, routerAddr)
	chout, cherr, err := gateway.Start()
	defer gateway.Stop()

	udpAddr, e := net.ResolveUDPAddr("udp", routerAddr)
	if e != nil {
		t.Errorf("Unexpected error %+v\n", e)
		return
	}

	conn, e := net.DialUDP("udp", nil, udpAddr)
	if e != nil {
		t.Errorf("Unexpected error %+v\n", e)
		return
	}
	defer conn.Close()

	Convey("Given a valid started gateway instance bound to a router", t, func() {
		Convey("Both channels should exist", func() {
			So(cherr, ShouldNotBeNil)
			So(chout, ShouldNotBeNil)
			So(err, ShouldBeNil)
		})

		Convey("A connection should exist", func() {
			So(len(gateway.routers), ShouldEqual, 1)
		})

		Convey("A valid packet should be forwarded", func() {
			packet := semtech.Packet{
				Version:    semtech.VERSION,
				Token:      []byte{0x1, 0x2},
				Identifier: semtech.PUSH_ACK,
			}
			raw, e := semtech.Marshal(packet)
			if e != nil {
				t.Errorf("Unexpected error %+v\n", e)
				return
			}
			conn.Write(raw)
			var received semtech.Packet
			select {
			case received = <-chout:
			case <-time.After(time.Second):
			}
			So(received, ShouldResemble, packet)
		})

		Convey("An invalid packet should raise an error", func() {
			conn.Write([]byte("Invalid"))
			var err error
			select {
			case err = <-cherr:
			case <-time.After(time.Second):
			}
			So(err, ShouldNotBeNil)
		})

		Convey("It should fail if started one more time", func() {
			_, _, err := gateway.Start()
			So(err, ShouldNotBeNil)
		})
	})
}

func TestStop(t *testing.T) {
	gatewayId := []byte("MyGateway")[:8]
	routerAddr := "0.0.0.0:3000"
	gateway, _ := New(gatewayId, routerAddr)
	Convey("Given a gateway instance", t, func() {
		Convey("It should failed if stopped while not started", func() {
			err := gateway.Stop()
			So(err, ShouldNotBeNil)
		})

		Convey("It should stop correctly after having started", func() {
			_, _, err := gateway.Start()
			if err != nil {
				t.Errorf("Unexpected error %v\n", err)
				return
			}
			time.Sleep(200 * time.Millisecond)
			err = gateway.Stop()

			So(err, ShouldBeNil)
			So(gateway.quit, ShouldBeNil)
			So(len(gateway.routers), ShouldEqual, 0)
		})
	})
}

func TestForward(t *testing.T) {
	gatewayId := []byte("MyGateway")[:8]
	routerAddr1 := "0.0.0.0:3000"
	routerAddr2 := "0.0.0.0:3001"

	gateway, _ := New(gatewayId, routerAddr1, routerAddr2)
	chout, _, e := gateway.Start()

	if e != nil {
		t.Errorf("Unexpected error %v", e)
		return
	}
	defer gateway.Stop()

	Convey("Given a started gateway bound to two routers", t, func() {
		Convey("When forwarding a valid packet", func() {
			packet := semtech.Packet{
				Version:    semtech.VERSION,
				Token:      []byte{0x1, 0x2},
				Identifier: semtech.PUSH_ACK,
			}
			gateway.Forward(packet)

			Convey("It should be forwarded to both routers", func() {
				var received semtech.Packet
				select {
				case received = <-chout:
				case <-time.After(time.Second):
				}
				So(received, ShouldResemble, packet)
				select {
				case received = <-chout:
				case <-time.After(time.Second):
				}
				So(received, ShouldResemble, packet)
			})
		})

		Convey("When forwarding an invalid packet", func() {
			packet := semtech.Packet{
				Version:    semtech.VERSION,
				Identifier: semtech.PUSH_ACK,
			}
			err := gateway.Forward(packet)
			Convey("The gateway should return an error", func() {
				So(err, ShouldNotBeNil)
			})
		})
	})
}

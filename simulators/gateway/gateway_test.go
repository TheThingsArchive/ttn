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

func TestNew(t *testing.T) {
	Convey("The New method should return a valid gateway struct ready to use", t, func() {
		id := []byte("qwerty")[:8]
		router1 := "0.0.0.0:3000"
		router2 := "0.0.0.0:1337"

		Convey("Given an identifier and a router address", func() {
			gateway, err := New(id, router1)

			Convey("No error should have been trown", func() {
				So(err, ShouldBeNil)
			})
			if err != nil {
				return
			}

			Convey("The identifier should have been set correctly", func() {
				So(gateway.Id, ShouldResemble, id)
			})

			Convey("The list of configured routers should have been set correctly", func() {
				So(len(gateway.routers), ShouldEqual, 1)
			})
		})

		Convey("Given an identifier and several routers address", func() {
			gateway, err := New(id, router1, router2)

			Convey("No error should have been trown", func() {
				So(err, ShouldBeNil)
			})
			if err != nil {
				return
			}

			Convey("The identifier should have been set correctly", func() {
				So(gateway.Id, ShouldResemble, id)
			})

			Convey("The list of configured routers should have been set correctly", func() {
				So(len(gateway.routers), ShouldEqual, 2)
			})
		})

		Convey("Given a bad identifier and/or bad router addresses", func() {
			Convey("It should return an error for an empty id", func() {
				gateway, err := New(make([]byte, 0), router1)
				So(gateway, ShouldBeNil)
				So(err, ShouldNotBeNil)
			})

			Convey("It should return an error for an empty routers list", func() {
				gateway, err := New(id)
				So(gateway, ShouldBeNil)
				So(err, ShouldNotBeNil)
			})

			Convey("It should return an error for an invalid router address", func() {
				gateway, err := New(id, "invalid")
				So(gateway, ShouldBeNil)
				So(err, ShouldNotBeNil)
			})
		})
	})
}

func TestStat(t *testing.T) {
	id := []byte("gatewayId")[:8]
	routerAddr := "0.0.0.0:3000"
	routerAddr2 := "0.0.0.0:3001"
	gateway, e := New(id, routerAddr, routerAddr2)
	if e != nil {
		t.Errorf("Unexpected error %+v\n", e)
		return
	}

	addr, e := net.ResolveUDPAddr("udp", routerAddr)
	if e != nil {
		t.Errorf("Unexpected error %+v\n", e)
		return
	}

	conn, e := net.DialUDP("udp", nil, addr)
	if e != nil {
		t.Errorf("Unexpected error %+v\n", e)
		return
	}
	defer conn.Close()

	Convey("Given a non started gateway", t, func() {
		stats := gateway.Stats()
		Convey("The stats should be equal to default values", func() {
			So(*stats.Ackr, ShouldEqual, 0.0)
			So(*stats.Dwnb, ShouldEqual, 0)
			So(*stats.Rxfw, ShouldEqual, 0)
			So(*stats.Rxnb, ShouldEqual, 0)
			So(*stats.Rxok, ShouldEqual, 0)
			So(*stats.Txnb, ShouldEqual, 0)
		})
	})

	_, _, e = gateway.Start()
	if e != nil {
		t.Errorf("Unexpected error %+v\n", e)
		return
	}
	defer gateway.Stop()

	Convey("Given a fresh gateway (just started)", t, func() {
		stats := gateway.Stats()
		Convey("The stats should be equal to default values", func() {
			So(*stats.Ackr, ShouldEqual, 0.0)
			So(*stats.Dwnb, ShouldEqual, 0)
			So(*stats.Rxfw, ShouldEqual, 0)
			So(*stats.Rxnb, ShouldEqual, 0)
			So(*stats.Rxok, ShouldEqual, 0)
			So(*stats.Txnb, ShouldEqual, 0)
			So(stats.Time, ShouldNotBeNil)
			So(stats.Lati, ShouldNotBeNil)
			So(stats.Long, ShouldNotBeNil)
			So(stats.Alti, ShouldNotBeNil)
		})

		Convey("After having received a valid downlink packet", func() {
			raw, e := semtech.Marshal(semtech.Packet{
				Version:    semtech.VERSION,
				Token:      genToken(),
				Identifier: semtech.PUSH_ACK,
			})
			if e != nil {
				t.Errorf("Unexpected error %+v\n")
				return
			}
			conn.Write(raw)
			time.Sleep(100 * time.Millisecond)

			Convey("The downlink packet number should have been incremented", func() {
				dwnb := *gateway.Stats().Dwnb
				So(dwnb, ShouldEqual, *stats.Dwnb+1)
			})
		})

		Convey("After having received an invalid downlink packet", func() {
			raw, e := semtech.Marshal(semtech.Packet{
				Version:    semtech.VERSION,
				Token:      genToken(),
				Identifier: semtech.PUSH_ACK,
			})
			if e != nil {
				t.Errorf("Unexpected error %+v\n")
				return
			}
			conn.Write(raw)
			time.Sleep(100 * time.Millisecond)
			Convey("The downlink packet number should have been incremented", func() {
				dwnb := *gateway.Stats().Dwnb
				So(dwnb, ShouldEqual, *stats.Dwnb+1)
			})
		})

		Convey("After having forwarded a valid radio packet", func() {
			packet := semtech.Packet{
				Version:    semtech.VERSION,
				Token:      genToken(),
				Identifier: semtech.PULL_DATA,
				GatewayId:  gateway.Id,
			}
			e := gateway.Forward(packet)
			if e != nil {
				t.Errorf("Unexpected error %+v\n", e)
				return
			}
			time.Sleep(100 * time.Millisecond)
			Convey("The number of packets forwarded should have been incremented", func() {
				rxfw := *gateway.Stats().Rxfw
				So(rxfw, ShouldEqual, *stats.Rxfw+1)
			})

			Convey("The number of packets received should have been incremented", func() {
				rxnb := *gateway.Stats().Rxnb
				So(rxnb, ShouldEqual, *stats.Rxnb+1)
			})

			Convey("The number of packets received with a valid PHY CRC should have been incremented", func() {
				rxok := *gateway.Stats().Rxok
				So(rxok, ShouldEqual, *stats.Rxok+1)
			})

			Convey("The number of packets emitted should have been incremented", func() {
				txnb := *gateway.Stats().Txnb
				So(txnb, ShouldEqual, *stats.Txnb+uint(len(gateway.routers)))
			})
		})

		Convey("After having forwarded an invalid radio packet", func() {
			packet := semtech.Packet{
				Version:    semtech.VERSION,
				Token:      genToken(),
				Identifier: semtech.PULL_DATA,
				GatewayId:  gateway.Id[:4], // Invalid Gateway id
			}
			e := gateway.Forward(packet)
			if e == nil {
				t.Errorf("An error was expected")
				return
			}
			time.Sleep(100 * time.Millisecond)
			Convey("The number of packets forwarded shouldn't have moved", func() {
				rxfw := *gateway.Stats().Rxfw
				So(rxfw, ShouldEqual, *stats.Rxfw)
			})

			Convey("The number of packets received should have been incremented", func() {
				rxnb := *gateway.Stats().Rxnb
				So(rxnb, ShouldEqual, *stats.Rxnb+1)
			})

			Convey("The number of packets received with a valid PHY CRC should have been incremented", func() {
				rxok := *gateway.Stats().Rxok
				So(rxok, ShouldEqual, *stats.Rxok+1)
			})

			Convey("The number of packets emitted shouldn't have moved", func() {
				txnb := *gateway.Stats().Txnb
				So(txnb, ShouldEqual, *stats.Txnb)
			})
		})
	})
}

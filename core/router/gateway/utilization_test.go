// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package gateway

import (
	"testing"

	"github.com/TheThingsNetwork/api/gateway"
	pb_protocol "github.com/TheThingsNetwork/api/protocol"
	pb_lorawan "github.com/TheThingsNetwork/api/protocol/lorawan"
	pb "github.com/TheThingsNetwork/api/router"
	. "github.com/smartystreets/assertions"
)

func buildUplink(freq uint64) *pb.UplinkMessage {
	return &pb.UplinkMessage{Payload: make([]byte, 10), ProtocolMetadata: pb_protocol.RxMetadata{
		Protocol: &pb_protocol.RxMetadata_LoRaWAN{LoRaWAN: &pb_lorawan.Metadata{
			DataRate:   "SF7BW125",
			CodingRate: "4/5",
		}},
	}, GatewayMetadata: gateway.RxMetadata{
		Frequency: freq,
	}}
}

func buildDownlink(freq uint64) *pb.DownlinkMessage {
	return &pb.DownlinkMessage{Payload: make([]byte, 10), ProtocolConfiguration: pb_protocol.TxConfiguration{
		Protocol: &pb_protocol.TxConfiguration_LoRaWAN{LoRaWAN: &pb_lorawan.TxConfiguration{
			DataRate:   "SF7BW125",
			CodingRate: "4/5",
		}},
	}, GatewayConfiguration: gateway.TxConfiguration{
		Frequency: freq,
	}}
}

func TestRxUtilization(t *testing.T) {
	a := New(t)
	u := NewUtilization()
	err := u.AddRx(buildUplink(8680000000))
	a.So(err, ShouldBeNil)
	err = u.AddRx(buildUplink(8682000000))
	a.So(err, ShouldBeNil)

	rx, tx := u.Get()
	a.So(rx, ShouldAlmostEqual, 0)
	a.So(tx, ShouldAlmostEqual, 0)

	u.Tick() // 5 seconds later

	rx, tx = u.GetChannel(8680000000)
	a.So(rx, ShouldAlmostEqual, 0.041216/5.0) // 41 ms per second
	a.So(tx, ShouldAlmostEqual, 0)

	rx, tx = u.Get()
	a.So(rx, ShouldAlmostEqual, 0.082432/5.0) // two times 41 ms per second
	a.So(tx, ShouldAlmostEqual, 0)

	u.AddRx(buildUplink(8680000000))
	u.AddRx(buildUplink(8682000000))

	u.Tick() // 5 seconds later

	rx, tx = u.GetChannel(8680000000)
	a.So(rx, ShouldAlmostEqual, 0.041216/5.0) // still 41 ms per second
	a.So(tx, ShouldAlmostEqual, 0)

	rx, tx = u.Get()
	a.So(rx, ShouldAlmostEqual, 0.082432/5.0) // still two times 41 ms per second
	a.So(tx, ShouldAlmostEqual, 0)
}

func TestTxUtilization(t *testing.T) {
	a := New(t)
	u := NewUtilization()
	err := u.AddTx(buildDownlink(8680000000))
	a.So(err, ShouldBeNil)
	err = u.AddTx(buildDownlink(8682000000))
	a.So(err, ShouldBeNil)

	rx, tx := u.Get()
	a.So(rx, ShouldAlmostEqual, 0)
	a.So(tx, ShouldAlmostEqual, 0)

	u.Tick() // 5 seconds later

	rx, tx = u.GetChannel(8680000000)
	a.So(rx, ShouldAlmostEqual, 0)
	a.So(tx, ShouldAlmostEqual, 0.041216/5.0) // 41 ms per second

	rx, tx = u.Get()
	a.So(rx, ShouldAlmostEqual, 0)
	a.So(tx, ShouldAlmostEqual, 0.082432/5.0) // two times 41 ms per second
}

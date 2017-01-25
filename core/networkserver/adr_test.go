// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package networkserver

import (
	"runtime"
	"testing"

	pb_broker "github.com/TheThingsNetwork/ttn/api/broker"
	pb_gateway "github.com/TheThingsNetwork/ttn/api/gateway"
	pb_protocol "github.com/TheThingsNetwork/ttn/api/protocol"
	pb_lorawan "github.com/TheThingsNetwork/ttn/api/protocol/lorawan"
	"github.com/TheThingsNetwork/ttn/core/networkserver/device"
	"github.com/TheThingsNetwork/ttn/core/types"
	. "github.com/TheThingsNetwork/ttn/utils/testing"
	"github.com/brocaar/lorawan"
	. "github.com/smartystreets/assertions"
)

func adrInitUplinkMessage() *pb_broker.DeduplicatedUplinkMessage {
	message := &pb_broker.DeduplicatedUplinkMessage{
		Message: new(pb_protocol.Message),
	}
	message.Message.InitLoRaWAN().InitUplink()
	downlink := message.InitResponseTemplate()
	downlink.Message = new(pb_protocol.Message)
	downlink.Message.InitLoRaWAN().InitDownlink()
	message.ProtocolMetadata = &pb_protocol.RxMetadata{Protocol: &pb_protocol.RxMetadata_Lorawan{
		Lorawan: &pb_lorawan.Metadata{
			DataRate: "SF8BW125",
		},
	}}
	message.GatewayMetadata = []*pb_gateway.RxMetadata{
		&pb_gateway.RxMetadata{
			Snr: 10,
		},
	}
	return message
}

func adrInitDownlinkMessage() *pb_broker.DownlinkMessage {
	message := &pb_broker.DownlinkMessage{
		Message: new(pb_protocol.Message),
	}
	message.Message.InitLoRaWAN().InitDownlink()
	return message
}

func TestHandleUplinkADR(t *testing.T) {
	a := New(t)
	ns := &networkServer{
		devices: device.NewRedisDeviceStore(GetRedisClient(), "ns-test-handle-uplink-adr"),
	}
	ns.InitStatus()

	defer func() {
		keys, _ := GetRedisClient().Keys("*ns-test-handle-uplink-adr*").Result()
		for _, key := range keys {
			GetRedisClient().Del(key).Result()
		}
	}()

	// Setting ADR to true should start collecting frames
	{
		dev := &device.Device{AppEUI: types.AppEUI([8]byte{1}), DevEUI: types.DevEUI([8]byte{1})}
		message := adrInitUplinkMessage()
		message.Message.GetLorawan().GetMacPayload().Adr = true
		err := ns.handleUplinkADR(message, dev)
		a.So(err, ShouldBeNil)
		frames, _ := ns.devices.GetFrames(dev.AppEUI, dev.DevEUI)
		a.So(frames, ShouldHaveLength, 1)
		a.So(dev.ADR.DataRate, ShouldEqual, "SF8BW125")
	}

	// Resetting ADR to false should empty the frames
	{
		dev := &device.Device{AppEUI: types.AppEUI([8]byte{1}), DevEUI: types.DevEUI([8]byte{1})}
		message := adrInitUplinkMessage()
		err := ns.handleUplinkADR(message, dev)
		a.So(err, ShouldBeNil)
		frames, _ := ns.devices.GetFrames(dev.AppEUI, dev.DevEUI)
		a.So(frames, ShouldBeEmpty)
	}

	// Setting ADRAckReq to true should set the ACK and schedule a LinkADRReq
	{
		dev := &device.Device{AppEUI: types.AppEUI([8]byte{1}), DevEUI: types.DevEUI([8]byte{1})}
		message := adrInitUplinkMessage()
		message.Message.GetLorawan().GetMacPayload().Adr = true
		message.Message.GetLorawan().GetMacPayload().AdrAckReq = true
		err := ns.handleUplinkADR(message, dev)
		a.So(err, ShouldBeNil)
		resMac := message.ResponseTemplate.Message.GetLorawan().GetMacPayload()
		a.So(resMac.Ack, ShouldBeTrue)
		a.So(dev.ADR.SendReq, ShouldBeTrue)
	}
}

func TestHandleDownlinkADR(t *testing.T) {
	a := New(t)
	ns := &networkServer{
		devices: device.NewRedisDeviceStore(GetRedisClient(), "ns-test-handle-downlink-adr"),
	}
	ns.InitStatus()

	defer func() {
		keys, _ := GetRedisClient().Keys("*ns-test-handle-downlink-adr*").Result()
		for _, key := range keys {
			GetRedisClient().Del(key).Result()
		}
	}()

	dev := &device.Device{AppEUI: types.AppEUI([8]byte{1}), DevEUI: types.DevEUI([8]byte{1})}

	message := adrInitDownlinkMessage()
	var shouldReturnError = func() {
		err := ns.handleDownlinkADR(message, dev)
		a.So(err, ShouldNotBeNil)
		a.So(message.Message.GetLorawan().GetMacPayload().FOpts, ShouldBeEmpty)
		if a.Failed() {
			_, file, line, _ := runtime.Caller(1)
			t.Errorf("\n%s:%d", file, line)
		}
	}
	var nothingShouldHappen = func() {
		err := ns.handleDownlinkADR(message, dev)
		a.So(err, ShouldBeNil)
		a.So(message.Message.GetLorawan().GetMacPayload().FOpts, ShouldBeEmpty)
		if a.Failed() {
			_, file, line, _ := runtime.Caller(1)
			t.Errorf("\n%s:%d", file, line)
		}
	}

	// initially
	nothingShouldHappen()

	dev.ADR.SendReq = true
	nothingShouldHappen()

	for i := 0; i < 20; i++ {
		ns.devices.PushFrame(dev.AppEUI, dev.DevEUI, &device.Frame{SNR: 10, GatewayCount: 3, FCnt: uint32(i)})
	}
	nothingShouldHappen()

	dev.ADR.DataRate = "SF8BW125"
	nothingShouldHappen()

	dev.ADR.Band = "INVALID"
	shouldReturnError()

	dev.ADR.Band = "US_902_928"
	nothingShouldHappen()

	dev.ADR.Band = "EU_863_870"
	err := ns.handleDownlinkADR(message, dev)
	a.So(err, ShouldBeNil)
	fOpts := message.Message.GetLorawan().GetMacPayload().FOpts
	a.So(fOpts, ShouldHaveLength, 1)
	a.So(fOpts[0].Cid, ShouldEqual, lorawan.LinkADRReq)
	var payload lorawan.LinkADRReqPayload
	payload.UnmarshalBinary(fOpts[0].Payload)
	a.So(payload.DataRate, ShouldEqual, 5) // SF7BW125
	a.So(payload.TXPower, ShouldEqual, 1)  // 14
	for i := 0; i < 8; i++ {               // First 8 channels enabled
		a.So(payload.ChMask[i], ShouldBeTrue)
	}
	a.So(payload.ChMask[8], ShouldBeFalse) // 9th channel (FSK) disabled

	// Invalid case
	message = adrInitDownlinkMessage()
	dev.ADR.DataRate = "INVALID"
	shouldReturnError()

}

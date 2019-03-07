// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package networkserver

import (
	"math"
	"sort"
	"testing"

	pb_broker "github.com/TheThingsNetwork/api/broker"
	pb_gateway "github.com/TheThingsNetwork/api/gateway"
	pb_protocol "github.com/TheThingsNetwork/api/protocol"
	pb_lorawan "github.com/TheThingsNetwork/api/protocol/lorawan"
	"github.com/TheThingsNetwork/ttn/core/component"
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
	message.ProtocolMetadata = pb_protocol.RxMetadata{Protocol: &pb_protocol.RxMetadata_LoRaWAN{
		LoRaWAN: &pb_lorawan.Metadata{
			DataRate: "SF10BW125",
		},
	}}
	message.GatewayMetadata = []*pb_gateway.RxMetadata{
		&pb_gateway.RxMetadata{
			SNR: 10,
		},
	}
	return message
}

func adrInitDownlinkMessage() *pb_broker.DownlinkMessage {
	message := &pb_broker.DownlinkMessage{
		Message: new(pb_protocol.Message),
	}
	dl := message.Message.InitLoRaWAN().InitDownlink()
	dl.FOpts = []pb_lorawan.MACCommand{
		pb_lorawan.MACCommand{CID: uint32(lorawan.LinkCheckAns)},
	}
	return message
}

func buildFrames(fCnts ...int) []*device.Frame {
	sort.Sort(sort.Reverse(sort.IntSlice(fCnts)))
	frames := make([]*device.Frame, 0, len(fCnts))
	for _, fCnt := range fCnts {
		frames = append(frames, &device.Frame{
			FCnt: uint32(fCnt),
			SNR:  float32(math.Floor(math.Sin(float64(fCnt))*100) / 10),
		})
	}
	return frames
}

func TestMaxSNR(t *testing.T) {
	a := New(t)
	a.So(maxSNR(buildFrames()), ShouldEqual, 0)
	a.So(maxSNR(buildFrames(1, 2, 3, 4, 5, 6, 7, 8, 9, 10)), ShouldEqual, 9.8)
}

func TestADR(t *testing.T) {
	ns := &networkServer{
		Component: &component.Component{
			Ctx: GetLogger(t, "TestADR"),
		},
		devices: device.NewRedisDeviceStore(GetRedisClient(), "ns-test-adr"),
	}
	ns.InitStatus()

	defer func() {
		keys, _ := GetRedisClient().Keys("*ns-test-adr*").Result()
		for _, key := range keys {
			GetRedisClient().Del(key).Result()
		}
	}()

	for i, tt := range []struct {
		Name                        string
		Band                        string
		DesiredInitialDataRate      string
		DesiredInitialDataRateIndex int
		DesiredInitialTxPower       int
		DesiredInitialTxPowerIndex  int
		DesiredDataRate             string
		DesiredDataRateIndex        int
		DesiredTxPower              int
		DesiredTxPowerIndex         int
	}{
		{
			Name: "EU Device", Band: "EU_863_870",
			DesiredInitialDataRate: "SF8BW125", DesiredInitialDataRateIndex: 4,
			DesiredInitialTxPower: 14, DesiredInitialTxPowerIndex: 1,
			DesiredDataRate: "SF7BW125", DesiredDataRateIndex: 5,
			DesiredTxPower: 14, DesiredTxPowerIndex: 1,
		},
		{
			Name: "AS Device", Band: "AS_923",
			DesiredInitialDataRate: "SF8BW125", DesiredInitialDataRateIndex: 4,
			DesiredInitialTxPower: 14, DesiredInitialTxPowerIndex: 0,
			DesiredDataRate: "SF7BW125", DesiredDataRateIndex: 5,
			DesiredTxPower: 14, DesiredTxPowerIndex: 0,
		},
		{
			Name: "AS1 Device", Band: "AS_920_923",
			DesiredInitialDataRate: "SF8BW125", DesiredInitialDataRateIndex: 4,
			DesiredInitialTxPower: 14, DesiredInitialTxPowerIndex: 0,
			DesiredDataRate: "SF7BW125", DesiredDataRateIndex: 5,
			DesiredTxPower: 14, DesiredTxPowerIndex: 0,
		},
		{
			Name: "AS2 Device", Band: "AS_923_925",
			DesiredInitialDataRate: "SF8BW125", DesiredInitialDataRateIndex: 4,
			DesiredInitialTxPower: 14, DesiredInitialTxPowerIndex: 0,
			DesiredDataRate: "SF7BW125", DesiredDataRateIndex: 5,
			DesiredTxPower: 14, DesiredTxPowerIndex: 0,
		},
		{
			Name: "KR Device", Band: "KR_920_923",
			DesiredInitialDataRate: "SF8BW125", DesiredInitialDataRateIndex: 4,
			DesiredInitialTxPower: 14, DesiredInitialTxPowerIndex: 1,
			DesiredDataRate: "SF7BW125", DesiredDataRateIndex: 5,
			DesiredTxPower: 14, DesiredTxPowerIndex: 1,
		},
		{
			Name: "RU Device", Band: "RU_864_870",
			DesiredInitialDataRate: "SF8BW125", DesiredInitialDataRateIndex: 4,
			DesiredInitialTxPower: 14, DesiredInitialTxPowerIndex: 1,
			DesiredDataRate: "SF7BW125", DesiredDataRateIndex: 5,
			DesiredTxPower: 14, DesiredTxPowerIndex: 1,
		},
		{
			Name: "US Device", Band: "US_902_928",
			DesiredInitialDataRate: "SF8BW125", DesiredInitialDataRateIndex: 2,
			DesiredInitialTxPower: 20, DesiredInitialTxPowerIndex: 5,
			DesiredDataRate: "SF7BW125", DesiredDataRateIndex: 3,
			DesiredTxPower: 20, DesiredTxPowerIndex: 5,
		},
		{
			Name: "AU Device", Band: "AU_915_928",
			DesiredInitialDataRate: "SF8BW125", DesiredInitialDataRateIndex: 4,
			DesiredInitialTxPower: 20, DesiredInitialTxPowerIndex: 5,
			DesiredDataRate: "SF7BW125", DesiredDataRateIndex: 5,
			DesiredTxPower: 20, DesiredTxPowerIndex: 5,
		},
	} {
		t.Run(tt.Name, func(t *testing.T) {
			a := New(t)

			appEUI := types.AppEUI([8]byte{1})
			devEUI := types.DevEUI([8]byte{1, uint8(i)})
			dev := &device.Device{AppEUI: appEUI, DevEUI: devEUI}

			history, _ := ns.devices.Frames(appEUI, devEUI)

			uplink, downlink := adrInitUplinkMessage(), adrInitDownlinkMessage()
			uplink.ProtocolMetadata.GetLoRaWAN().FrequencyPlan = pb_lorawan.FrequencyPlan(pb_lorawan.FrequencyPlan_value[tt.Band])

			err := ns.handleUplinkADR(uplink, dev)
			a.So(err, ShouldBeNil)
			a.So(dev.ADR.SendReq, ShouldBeFalse)

			uplink.Message.GetLoRaWAN().GetMACPayload().ADR = true

			err = ns.handleUplinkADR(uplink, dev)
			a.So(err, ShouldBeNil)
			a.So(dev.ADR.SendReq, ShouldBeTrue)
			a.So(dev.ADR.DataRate, ShouldEqual, tt.DesiredInitialDataRate)
			a.So(dev.ADR.TxPower, ShouldEqual, tt.DesiredInitialTxPower)

			frames, err := history.Get()
			a.So(err, ShouldBeNil)
			a.So(frames, ShouldHaveLength, 1)

			err = ns.handleDownlinkADR(downlink, dev)
			a.So(err, ShouldBeNil)

			a.So(dev.ADR.SentInitial, ShouldBeTrue)

			fOpts := downlink.Message.GetLoRaWAN().GetMACPayload().FOpts
			fOpt := fOpts[1]
			if tt.Band == "US_902_928" || tt.Band == "AU_915_928" {
				fOpt = fOpts[2]

				var req lorawan.LinkADRReqPayload
				req.UnmarshalBinary(fOpts[1].Payload)

				switch tt.Band {
				case "US_902_928":
					a.So(req.DataRate, ShouldEqual, 4) // 500kHz channel, so DR4
				case "AU_915_928":
					a.So(req.DataRate, ShouldEqual, 6) // 500kHz channel, so DR6
				}

				a.So(req.TXPower, ShouldEqual, 0)               // Max tx power
				a.So(req.Redundancy.ChMaskCntl, ShouldEqual, 7) // Ch 64-71, All 125 kHz channels off
				a.So(req.ChMask[0], ShouldBeFalse)              // Channel 64 disabled
				a.So(req.ChMask[1], ShouldBeTrue)               // Channel 65 enabled
				for i := 2; i < 8; i++ {                        // Channels 66-71 disabled
					a.So(req.ChMask[i], ShouldBeFalse)
				}
			}

			a.So(fOpt.CID, ShouldEqual, lorawan.LinkADRReq)
			var req lorawan.LinkADRReqPayload
			req.UnmarshalBinary(fOpt.Payload)

			a.So(req.DataRate, ShouldEqual, tt.DesiredInitialDataRateIndex)
			a.So(req.TXPower, ShouldEqual, tt.DesiredInitialTxPowerIndex)

			if tt.Band == "US_902_928" || tt.Band == "AU_915_928" {
				a.So(req.Redundancy.ChMaskCntl, ShouldEqual, 0) // Channels 0..15
				for i := 0; i < 8; i++ {                        // First 8 channels disabled
					a.So(req.ChMask[i], ShouldBeFalse)
				}
				for i := 8; i < 16; i++ { // Second 8 channels enabled
					a.So(req.ChMask[i], ShouldBeTrue)
				}
			}

			for i := 0; i < 20; i++ {
				uplink.Message.GetLoRaWAN().GetMACPayload().FCnt++
				ns.handleUplinkADR(uplink, dev)
			}

			frames, err = history.Get()
			a.So(err, ShouldBeNil)
			a.So(frames, ShouldHaveLength, 20)

			a.So(dev.ADR.SendReq, ShouldBeTrue)
			a.So(dev.ADR.DataRate, ShouldEqual, tt.DesiredDataRate)
			a.So(dev.ADR.TxPower, ShouldEqual, tt.DesiredTxPower)

			err = ns.handleDownlinkADR(downlink, dev)
			a.So(err, ShouldBeNil)

			fOpts = downlink.Message.GetLoRaWAN().GetMACPayload().FOpts
			fOpt = fOpts[1]
			if tt.Band == "US_902_928" || tt.Band == "AU_915_928" {
				fOpt = fOpts[2]
			}

			a.So(fOpt.CID, ShouldEqual, lorawan.LinkADRReq)
			req.UnmarshalBinary(fOpt.Payload)

			a.So(req.DataRate, ShouldEqual, tt.DesiredDataRateIndex)
			a.So(req.TXPower, ShouldEqual, tt.DesiredTxPowerIndex)

			uplink.ProtocolMetadata.GetLoRaWAN().DataRate = tt.DesiredDataRate
			uplink.GatewayMetadata[0].SNR = 7.5

			err = ns.handleUplinkADR(uplink, dev)
			a.So(err, ShouldBeNil)

			a.So(dev.ADR.SendReq, ShouldBeFalse)
		})
	}
}

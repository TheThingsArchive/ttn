// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package networkserver

import (
	"testing"
	"time"

	pb_broker "github.com/TheThingsNetwork/ttn/api/broker"
	pb_gateway "github.com/TheThingsNetwork/ttn/api/gateway"
	pb_protocol "github.com/TheThingsNetwork/ttn/api/protocol"
	pb_lorawan "github.com/TheThingsNetwork/ttn/api/protocol/lorawan"
	"github.com/TheThingsNetwork/ttn/core/component"
	"github.com/TheThingsNetwork/ttn/core/networkserver/device"
	"github.com/TheThingsNetwork/ttn/core/types"
	. "github.com/TheThingsNetwork/ttn/utils/testing"
	"github.com/brocaar/lorawan"
	. "github.com/smartystreets/assertions"
)

func TestHandleUplink(t *testing.T) {
	a := New(t)
	ns := &networkServer{
		Component: &component.Component{
			Ctx: GetLogger(t, "TestHandleUplink"),
		},
		devices: device.NewRedisDeviceStore(GetRedisClient(), "ns-test-handle-uplink"),
	}
	ns.InitStatus()

	appEUI := types.AppEUI(getEUI(1, 2, 3, 4, 5, 6, 7, 8))
	devEUI := types.DevEUI(getEUI(1, 2, 3, 4, 5, 6, 7, 8))
	devAddr := getDevAddr(1, 2, 3, 4)

	// Device Not Found
	message := &pb_broker.DeduplicatedUplinkMessage{
		AppEui:  &appEUI,
		DevEui:  &devEUI,
		Payload: []byte{},
	}
	_, err := ns.HandleUplink(message)
	a.So(err, ShouldNotBeNil)

	ns.devices.Set(&device.Device{
		DevAddr: devAddr,
		AppEUI:  appEUI,
		DevEUI:  devEUI,
	})
	defer func() {
		ns.devices.Delete(appEUI, devEUI)
		frames, _ := ns.devices.Frames(appEUI, devEUI)
		frames.Clear()
	}()

	// Invalid Payload
	message = &pb_broker.DeduplicatedUplinkMessage{
		AppEui:  &appEUI,
		DevEui:  &devEUI,
		Payload: []byte{},
	}
	_, err = ns.HandleUplink(message)
	a.So(err, ShouldNotBeNil)

	phy := lorawan.PHYPayload{
		MHDR: lorawan.MHDR{
			MType: lorawan.UnconfirmedDataUp,
			Major: lorawan.LoRaWANR1,
		},
		MACPayload: &lorawan.MACPayload{
			FHDR: lorawan.FHDR{
				DevAddr: lorawan.DevAddr([4]byte{1, 2, 3, 4}),
				FCnt:    1,
				FCtrl: lorawan.FCtrl{
					ADR:       true,
					ADRACKReq: true,
				},
				FOpts: []lorawan.MACCommand{
					lorawan.MACCommand{CID: lorawan.LinkCheckReq},
				},
			},
		},
	}
	bytes, _ := phy.MarshalBinary()

	// Valid Uplink
	message = &pb_broker.DeduplicatedUplinkMessage{
		AppEui:           &appEUI,
		DevEui:           &devEUI,
		Payload:          bytes,
		ResponseTemplate: &pb_broker.DownlinkMessage{DownlinkOption: &pb_broker.DownlinkOption{}},
		GatewayMetadata: []*pb_gateway.RxMetadata{
			&pb_gateway.RxMetadata{},
		},
		ProtocolMetadata: &pb_protocol.RxMetadata{Protocol: &pb_protocol.RxMetadata_Lorawan{
			Lorawan: &pb_lorawan.Metadata{
				DataRate: "SF7BW125",
			},
		}},
	}
	res, err := ns.HandleUplink(message)
	a.So(err, ShouldBeNil)
	a.So(res.ResponseTemplate, ShouldNotBeNil)

	// LoRaWAN: Unmarshal
	var phyPayload lorawan.PHYPayload
	phyPayload.UnmarshalBinary(res.ResponseTemplate.Payload)
	macPayload, _ := phyPayload.MACPayload.(*lorawan.MACPayload)

	// ResponseTemplate DevAddr should match
	a.So([4]byte(macPayload.FHDR.DevAddr), ShouldEqual, [4]byte(devAddr))

	// ResponseTemplate should ACK the ADRACKReq
	a.So(macPayload.FHDR.FCtrl.ACK, ShouldBeTrue)
	a.So(macPayload.FHDR.FOpts, ShouldHaveLength, 1)
	a.So(macPayload.FHDR.FOpts[0].Payload, ShouldResemble, &lorawan.LinkCheckAnsPayload{GwCnt: 1, Margin: 7})

	// Frame Counter should have been updated
	dev, _ := ns.devices.Get(appEUI, devEUI)
	a.So(dev.FCntUp, ShouldEqual, 1)
	a.So(time.Now().Sub(dev.LastSeen), ShouldBeLessThan, 1*time.Second)
}

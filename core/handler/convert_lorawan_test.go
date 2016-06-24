// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package handler

import (
	"testing"

	pb_broker "github.com/TheThingsNetwork/ttn/api/broker"
	pb_protocol "github.com/TheThingsNetwork/ttn/api/protocol"
	pb_lorawan "github.com/TheThingsNetwork/ttn/api/protocol/lorawan"
	"github.com/TheThingsNetwork/ttn/core"
	"github.com/TheThingsNetwork/ttn/core/handler/device"
	"github.com/TheThingsNetwork/ttn/core/types"
	"github.com/TheThingsNetwork/ttn/mqtt"
	. "github.com/TheThingsNetwork/ttn/utils/testing"
	. "github.com/smartystreets/assertions"
)

var testDevEUI = types.DevEUI{1, 2, 3, 4, 5, 6, 7, 8}
var testAppEUI = types.AppEUI{1, 2, 3, 4, 5, 6, 7, 8}

func buildLorawanUplink(payload []byte) (*pb_broker.DeduplicatedUplinkMessage, *mqtt.UplinkMessage) {
	ttnUp := &pb_broker.DeduplicatedUplinkMessage{
		DevEui:  &testDevEUI,
		AppEui:  &testAppEUI,
		Payload: payload,
		ProtocolMetadata: &pb_protocol.RxMetadata{Protocol: &pb_protocol.RxMetadata_Lorawan{
			Lorawan: &pb_lorawan.Metadata{
				FCnt: 1,
			},
		}},
	}
	appUp := &mqtt.UplinkMessage{}
	return ttnUp, appUp
}

func TestConvertFromLoRaWAN(t *testing.T) {
	a := New(t)
	h := &handler{
		devices:   device.NewDeviceStore(),
		Component: &core.Component{Ctx: GetLogger(t, "TestConvertFromLoRaWAN")},
	}
	h.devices.Set(&device.Device{
		DevEUI: testDevEUI,
		AppEUI: testAppEUI,
	})
	ttnUp, appUp := buildLorawanUplink([]byte{0x40, 0x04, 0x03, 0x02, 0x01, 0x00, 0x01, 0x00, 0x01, 0x46, 0x55, 0x23, 0xf4, 0xf8, 0x45})
	err := h.ConvertFromLoRaWAN(h.Ctx, ttnUp, appUp)
	a.So(err, ShouldBeNil)
	a.So(appUp.Payload, ShouldResemble, []byte{0xaa, 0xbc})
	a.So(appUp.FCnt, ShouldEqual, 1)
}

func buildLorawanDownlink(payload []byte) (*mqtt.DownlinkMessage, *pb_broker.DownlinkMessage) {
	appDown := &mqtt.DownlinkMessage{
		DevEUI:  testDevEUI,
		AppEUI:  testAppEUI,
		Payload: []byte{0xaa, 0xbc},
	}
	ttnDown := &pb_broker.DownlinkMessage{
		Payload: []byte{96, 4, 3, 2, 1, 0, 1, 0, 1, 0, 0, 0, 0},
		DownlinkOption: &pb_broker.DownlinkOption{
			ProtocolConfig: &pb_protocol.TxConfiguration{Protocol: &pb_protocol.TxConfiguration_Lorawan{
				Lorawan: &pb_lorawan.TxConfiguration{
					FCnt: 1,
				},
			}},
		},
	}
	return appDown, ttnDown
}

func TestConvertToLoRaWAN(t *testing.T) {
	a := New(t)
	h := &handler{
		devices:   device.NewDeviceStore(),
		Component: &core.Component{Ctx: GetLogger(t, "TestConvertToLoRaWAN")},
	}
	h.devices.Set(&device.Device{
		DevEUI: testDevEUI,
		AppEUI: testAppEUI,
	})
	appDown, ttnDown := buildLorawanDownlink([]byte{0xaa, 0xbc})
	err := h.ConvertToLoRaWAN(h.Ctx, appDown, ttnDown)
	a.So(err, ShouldBeNil)
	a.So(ttnDown.Payload, ShouldResemble, []byte{0x60, 0x04, 0x03, 0x02, 0x01, 0x00, 0x01, 0x00, 0x01, 0xa1, 0x33, 0x00, 0x00, 0x00, 0x00})

	appDown, ttnDown = buildLorawanDownlink([]byte{0xaa, 0xbc})
	appDown.FPort = 8
	err = h.ConvertToLoRaWAN(h.Ctx, appDown, ttnDown)
	a.So(err, ShouldBeNil)
	a.So(ttnDown.Payload, ShouldResemble, []byte{0x60, 0x04, 0x03, 0x02, 0x01, 0x00, 0x01, 0x00, 0x08, 0xa1, 0x33, 0x00, 0x00, 0x00, 0x00})
}

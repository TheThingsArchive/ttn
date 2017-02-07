// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package networkserver

import (
	"testing"

	pb_broker "github.com/TheThingsNetwork/ttn/api/broker"
	pb_protocol "github.com/TheThingsNetwork/ttn/api/protocol"
	pb_lorawan "github.com/TheThingsNetwork/ttn/api/protocol/lorawan"
	"github.com/TheThingsNetwork/ttn/core/networkserver/device"
	"github.com/TheThingsNetwork/ttn/core/types"
	. "github.com/TheThingsNetwork/ttn/utils/testing"
	"github.com/brocaar/lorawan"
	. "github.com/smartystreets/assertions"
)

func TestHandleDownlink(t *testing.T) {
	a := New(t)
	ns := &networkServer{
		devices: device.NewRedisDeviceStore(GetRedisClient(), "test-handle-downlink"),
	}
	ns.InitStatus()

	appEUI := types.AppEUI(getEUI(1, 2, 3, 4, 5, 6, 7, 8))
	devEUI := types.DevEUI(getEUI(1, 2, 3, 4, 5, 6, 7, 8))
	devAddr := getDevAddr(1, 2, 3, 4)

	downlinkOption := &pb_broker.DownlinkOption{
		ProtocolConfig: &pb_protocol.TxConfiguration{Protocol: &pb_protocol.TxConfiguration_Lorawan{
			Lorawan: &pb_lorawan.TxConfiguration{},
		}},
	}

	// Device Not Found
	message := &pb_broker.DownlinkMessage{
		AppEui:         &appEUI,
		DevEui:         &devEUI,
		Payload:        []byte{},
		DownlinkOption: downlinkOption,
	}
	_, err := ns.HandleDownlink(message)
	a.So(err, ShouldNotBeNil)

	ns.devices.Set(&device.Device{
		DevAddr: devAddr,
		AppEUI:  appEUI,
		DevEUI:  devEUI,
	})
	defer func() {
		ns.devices.Delete(appEUI, devEUI)
	}()

	// Invalid Payload
	message = &pb_broker.DownlinkMessage{
		AppEui:         &appEUI,
		DevEui:         &devEUI,
		Payload:        []byte{},
		DownlinkOption: downlinkOption,
	}
	_, err = ns.HandleDownlink(message)
	a.So(err, ShouldNotBeNil)

	fPort := uint8(3)
	phy := lorawan.PHYPayload{
		MHDR: lorawan.MHDR{
			MType: lorawan.UnconfirmedDataDown,
			Major: lorawan.LoRaWANR1,
		},
		MACPayload: &lorawan.MACPayload{
			FPort: &fPort,
			FHDR: lorawan.FHDR{
				FCtrl: lorawan.FCtrl{
					ACK: true,
				},
				DevAddr: lorawan.DevAddr(devAddr),
			},
		},
	}
	bytes, _ := phy.MarshalBinary()

	message = &pb_broker.DownlinkMessage{
		AppEui:         &appEUI,
		DevEui:         &devEUI,
		Payload:        bytes,
		DownlinkOption: downlinkOption,
	}
	res, err := ns.HandleDownlink(message)
	a.So(err, ShouldBeNil)

	var phyPayload lorawan.PHYPayload
	phyPayload.UnmarshalBinary(res.Payload)
	macPayload, _ := phyPayload.MACPayload.(*lorawan.MACPayload)
	a.So(*macPayload.FPort, ShouldEqual, 3)
	a.So(macPayload.FHDR.DevAddr, ShouldEqual, lorawan.DevAddr{1, 2, 3, 4})
	a.So(macPayload.FHDR.FCnt, ShouldEqual, 0)                // The first Frame counter is zero
	a.So(phyPayload.MIC, ShouldNotEqual, [4]byte{0, 0, 0, 0}) // MIC should be set, we'll check it with actual examples in the integration test

	dev, _ := ns.devices.Get(appEUI, devEUI)
	a.So(dev.FCntDown, ShouldEqual, 1)

}

// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package networkserver

import (
	"testing"
	"time"

	pb_broker "github.com/TheThingsNetwork/ttn/api/broker"
	"github.com/TheThingsNetwork/ttn/core/networkserver/device"
	"github.com/TheThingsNetwork/ttn/core/types"
	"github.com/brocaar/lorawan"
	. "github.com/smartystreets/assertions"
)

func TestHandleUplink(t *testing.T) {
	a := New(t)
	ns := &networkServer{
		devices: device.NewDeviceStore(),
	}

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
			},
		},
	}
	bytes, _ := phy.MarshalBinary()

	// Valid Uplink
	message = &pb_broker.DeduplicatedUplinkMessage{
		AppEui:           &appEUI,
		DevEui:           &devEUI,
		Payload:          bytes,
		ResponseTemplate: &pb_broker.DownlinkMessage{},
	}
	res, err := ns.HandleUplink(message)
	a.So(err, ShouldBeNil)
	a.So(res.ResponseTemplate, ShouldNotBeNil)

	// LoRaWAN: Unmarshal
	var phyPayload lorawan.PHYPayload
	phyPayload.UnmarshalBinary(res.ResponseTemplate.Payload)
	macPayload, _ := phyPayload.MACPayload.(*lorawan.MACPayload)

	a.So([4]byte(macPayload.FHDR.DevAddr), ShouldEqual, [4]byte(devAddr))

	// Frame Counter should have been updated
	dev, _ := ns.devices.Get(appEUI, devEUI)
	a.So(dev.FCntUp, ShouldEqual, 1)
	a.So(time.Now().Sub(dev.LastSeen), ShouldBeLessThan, 1*time.Second)
}

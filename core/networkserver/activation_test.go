// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package networkserver

import (
	"testing"

	pb_broker "github.com/TheThingsNetwork/ttn/api/broker"
	pb_handler "github.com/TheThingsNetwork/ttn/api/handler"
	pb_protocol "github.com/TheThingsNetwork/ttn/api/protocol"
	pb_lorawan "github.com/TheThingsNetwork/ttn/api/protocol/lorawan"
	"github.com/TheThingsNetwork/ttn/core/networkserver/device"
	"github.com/TheThingsNetwork/ttn/core/types"
	"github.com/brocaar/lorawan"
	. "github.com/smartystreets/assertions"
)

func TestHandlePrepareActivation(t *testing.T) {
	a := New(t)
	ns := &networkServer{
		netID: [3]byte{0x00, 0x00, 0x13},
		prefixes: map[types.DevAddrPrefix][]string{
			types.DevAddrPrefix{DevAddr: [4]byte{0x26, 0x00, 0x00, 0x00}, Length: 7}: []string{
				"otaa",
				"local",
			},
		},
		devices: device.NewDeviceStore(),
	}

	appEUI := types.AppEUI(getEUI(2, 2, 3, 4, 5, 6, 7, 8))
	devEUI := types.DevEUI(getEUI(2, 2, 3, 4, 5, 6, 7, 8))

	// Device not registered
	resp, err := ns.HandlePrepareActivation(&pb_broker.DeduplicatedDeviceActivationRequest{
		ActivationMetadata: &pb_protocol.ActivationMetadata{Protocol: &pb_protocol.ActivationMetadata_Lorawan{
			Lorawan: &pb_lorawan.ActivationMetadata{
				CfList: []uint64{867100000, 867300000, 867500000, 867700000, 867900000},
			},
		}},
		ResponseTemplate: &pb_broker.DeviceActivationResponse{},
	})
	a.So(err, ShouldNotBeNil)

	// Constrained Device
	ns.devices.Set(&device.Device{AppEUI: appEUI, DevEUI: devEUI, Options: device.Options{
		ActivationConstraints: "private",
	}})
	resp, err = ns.HandlePrepareActivation(&pb_broker.DeduplicatedDeviceActivationRequest{
		DevEui: &devEUI,
		AppEui: &appEUI,
		ActivationMetadata: &pb_protocol.ActivationMetadata{Protocol: &pb_protocol.ActivationMetadata_Lorawan{
			Lorawan: &pb_lorawan.ActivationMetadata{
				CfList: []uint64{867100000, 867300000, 867500000, 867700000, 867900000},
			},
		}},
		ResponseTemplate: &pb_broker.DeviceActivationResponse{},
	})
	a.So(err, ShouldNotBeNil)

	// Device registered
	ns.devices.Set(&device.Device{AppEUI: appEUI, DevEUI: devEUI})
	resp, err = ns.HandlePrepareActivation(&pb_broker.DeduplicatedDeviceActivationRequest{
		DevEui: &devEUI,
		AppEui: &appEUI,
		ActivationMetadata: &pb_protocol.ActivationMetadata{Protocol: &pb_protocol.ActivationMetadata_Lorawan{
			Lorawan: &pb_lorawan.ActivationMetadata{
				CfList: []uint64{867100000, 867300000, 867500000, 867700000, 867900000},
			},
		}},
		ResponseTemplate: &pb_broker.DeviceActivationResponse{},
	})
	a.So(err, ShouldBeNil)
	devAddr := resp.ActivationMetadata.GetLorawan().DevAddr
	a.So(devAddr.IsEmpty(), ShouldBeFalse)
	a.So(devAddr[0]&254, ShouldEqual, 19<<1) // 7 MSB should be NetID

	var resPHY lorawan.PHYPayload
	resPHY.UnmarshalBinary(resp.ResponseTemplate.Payload)
	resMAC, _ := resPHY.MACPayload.(*lorawan.DataPayload)
	joinAccept := &lorawan.JoinAcceptPayload{}
	joinAccept.UnmarshalBinary(false, resMAC.Bytes)

	a.So(joinAccept.DevAddr[0]&254, ShouldEqual, 19<<1)
	a.So(*joinAccept.CFList, ShouldEqual, lorawan.CFList{867100000, 867300000, 867500000, 867700000, 867900000})
}

func TestHandleActivate(t *testing.T) {
	a := New(t)
	ns := &networkServer{
		devices: device.NewDeviceStore(),
	}
	ns.devices.Set(&device.Device{
		AppEUI: types.AppEUI(getEUI(0, 0, 0, 0, 0, 0, 3, 1)),
		DevEUI: types.DevEUI(getEUI(0, 0, 0, 0, 0, 0, 3, 1)),
	})

	_, err := ns.HandleActivate(&pb_handler.DeviceActivationResponse{})
	a.So(err, ShouldNotBeNil)

	_, err = ns.HandleActivate(&pb_handler.DeviceActivationResponse{
		ActivationMetadata: &pb_protocol.ActivationMetadata{},
	})
	a.So(err, ShouldNotBeNil)

	devAddr := getDevAddr(0, 0, 3, 1)
	var nwkSKey types.NwkSKey
	copy(nwkSKey[:], []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 3, 1})
	appEUI := types.AppEUI(getEUI(0, 0, 0, 0, 0, 0, 3, 1))
	devEUI := types.DevEUI(getEUI(0, 0, 0, 0, 0, 0, 3, 1))
	_, err = ns.HandleActivate(&pb_handler.DeviceActivationResponse{
		ActivationMetadata: &pb_protocol.ActivationMetadata{Protocol: &pb_protocol.ActivationMetadata_Lorawan{
			Lorawan: &pb_lorawan.ActivationMetadata{
				AppEui:  &appEUI,
				DevEui:  &devEUI,
				DevAddr: &devAddr,
				NwkSKey: &nwkSKey,
			},
		}},
	})
	a.So(err, ShouldBeNil)
}

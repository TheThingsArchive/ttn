// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package networkserver

import (
	"testing"
	"time"

	"gopkg.in/redis.v3"

	"github.com/brocaar/lorawan"
	. "github.com/smartystreets/assertions"

	pb_broker "github.com/TheThingsNetwork/ttn/api/broker"
	pb_handler "github.com/TheThingsNetwork/ttn/api/handler"
	pb "github.com/TheThingsNetwork/ttn/api/networkserver"
	pb_protocol "github.com/TheThingsNetwork/ttn/api/protocol"
	pb_lorawan "github.com/TheThingsNetwork/ttn/api/protocol/lorawan"
	"github.com/TheThingsNetwork/ttn/core/networkserver/device"
	"github.com/TheThingsNetwork/ttn/core/types"
)

func getDevAddr(bytes ...byte) (addr types.DevAddr) {
	copy(addr[:], bytes[:4])
	return
}

func getEUI(bytes ...byte) (eui types.EUI64) {
	copy(eui[:], bytes[:8])
	return
}

func TestNewNetworkServer(t *testing.T) {
	a := New(t)
	var client redis.Client

	// TTN NetID
	ns := NewRedisNetworkServer(&client, 19)
	a.So(ns, ShouldNotBeNil)
	a.So(ns.(*networkServer).netID, ShouldEqual, [3]byte{0, 0, 0x13})

	// Other NetID, same NwkID
	ns = NewRedisNetworkServer(&client, 66067)
	a.So(ns, ShouldNotBeNil)
	a.So(ns.(*networkServer).netID, ShouldEqual, [3]byte{0x01, 0x02, 0x13})
}

func TestUsePrefix(t *testing.T) {
	a := New(t)
	var client redis.Client
	ns := NewRedisNetworkServer(&client, 19)

	a.So(ns.UsePrefix([]byte{}, 0), ShouldNotBeNil)
	a.So(ns.UsePrefix([]byte{0x14}, 7), ShouldNotBeNil)
	a.So(ns.UsePrefix([]byte{0x26}, 7), ShouldBeNil)
	a.So(ns.(*networkServer).prefix, ShouldEqual, [4]byte{0x26, 0x00, 0x00, 0x00})
}

func TestHandleGetDevices(t *testing.T) {
	a := New(t)

	ns := &networkServer{
		devices: device.NewDeviceStore(),
	}

	// No Devices
	devAddr1 := getDevAddr(1, 2, 3, 4)
	res, err := ns.HandleGetDevices(&pb.DevicesRequest{
		DevAddr: &devAddr1,
		FCnt:    5,
	})
	a.So(err, ShouldBeNil)
	a.So(res.Results, ShouldBeEmpty)

	// Matching Device
	ns.devices.Set(&device.Device{
		DevAddr: getDevAddr(1, 2, 3, 4),
		AppEUI:  types.AppEUI(getEUI(1, 2, 3, 4, 5, 6, 7, 8)),
		DevEUI:  types.DevEUI(getEUI(1, 2, 3, 4, 5, 6, 7, 8)),
		NwkSKey: types.NwkSKey{1, 2, 3, 4, 5, 6, 7, 8, 1, 2, 3, 4, 5, 6, 7, 8},
		FCntUp:  5,
	})
	res, err = ns.HandleGetDevices(&pb.DevicesRequest{
		DevAddr: &devAddr1,
		FCnt:    5,
	})
	a.So(err, ShouldBeNil)
	a.So(res.Results, ShouldHaveLength, 1)

	// Non-Matching DevAddr
	devAddr2 := getDevAddr(5, 6, 7, 8)
	res, err = ns.HandleGetDevices(&pb.DevicesRequest{
		DevAddr: &devAddr2,
		FCnt:    5,
	})
	a.So(err, ShouldBeNil)
	a.So(res.Results, ShouldHaveLength, 0)

	// Non-Matching FCnt
	res, err = ns.HandleGetDevices(&pb.DevicesRequest{
		DevAddr: &devAddr1,
		FCnt:    4,
	})
	a.So(err, ShouldBeNil)
	a.So(res.Results, ShouldHaveLength, 0)

	// Non-Matching FCnt, but FCnt Check Disabled
	ns.devices.Set(&device.Device{
		DevAddr: getDevAddr(5, 6, 7, 8),
		AppEUI:  types.AppEUI(getEUI(5, 6, 7, 8, 1, 2, 3, 4)),
		DevEUI:  types.DevEUI(getEUI(5, 6, 7, 8, 1, 2, 3, 4)),
		NwkSKey: types.NwkSKey{1, 2, 3, 4, 5, 6, 7, 8, 1, 2, 3, 4, 5, 6, 7, 8},
		FCntUp:  5,
		Options: device.Options{
			DisableFCntCheck: true,
		},
	})
	res, err = ns.HandleGetDevices(&pb.DevicesRequest{
		DevAddr: &devAddr2,
		FCnt:    4,
	})
	a.So(err, ShouldBeNil)
	a.So(res.Results, ShouldHaveLength, 1)

	// 32 Bit Frame Counter (A)
	devAddr3 := getDevAddr(2, 2, 3, 4)
	ns.devices.Set(&device.Device{
		DevAddr: getDevAddr(2, 2, 3, 4),
		AppEUI:  types.AppEUI(getEUI(2, 2, 3, 4, 5, 6, 7, 8)),
		DevEUI:  types.DevEUI(getEUI(2, 2, 3, 4, 5, 6, 7, 8)),
		NwkSKey: types.NwkSKey{1, 2, 3, 4, 5, 6, 7, 8, 1, 2, 3, 4, 5, 6, 7, 8},
		FCntUp:  5 + (2 << 16),
		Options: device.Options{
			Uses32BitFCnt: true,
		},
	})
	res, err = ns.HandleGetDevices(&pb.DevicesRequest{
		DevAddr: &devAddr3,
		FCnt:    5,
	})
	a.So(err, ShouldBeNil)
	a.So(res.Results, ShouldHaveLength, 1)

	// 32 Bit Frame Counter (B)
	ns.devices.Set(&device.Device{
		DevAddr: devAddr3,
		AppEUI:  types.AppEUI(getEUI(2, 2, 3, 4, 5, 6, 7, 8)),
		DevEUI:  types.DevEUI(getEUI(2, 2, 3, 4, 5, 6, 7, 8)),
		NwkSKey: types.NwkSKey{1, 2, 3, 4, 5, 6, 7, 8, 1, 2, 3, 4, 5, 6, 7, 8},
		FCntUp:  (2 << 16) - 1,
		Options: device.Options{
			Uses32BitFCnt: true,
		},
	})
	res, err = ns.HandleGetDevices(&pb.DevicesRequest{
		DevAddr: &devAddr3,
		FCnt:    5,
	})
	a.So(err, ShouldBeNil)
	a.So(res.Results, ShouldHaveLength, 1)

}

func TestHandlePrepareActivation(t *testing.T) {
	a := New(t)
	ns := &networkServer{
		netID:        [3]byte{0x00, 0x00, 0x13},
		prefix:       [4]byte{0x26, 0x00, 0x00, 0x00},
		prefixLength: 7,
		devices:      device.NewDeviceStore(),
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

func TestHandleDownlink(t *testing.T) {
	a := New(t)
	ns := &networkServer{
		devices: device.NewDeviceStore(),
	}

	appEUI := types.AppEUI(getEUI(1, 2, 3, 4, 5, 6, 7, 8))
	devEUI := types.DevEUI(getEUI(1, 2, 3, 4, 5, 6, 7, 8))
	devAddr := getDevAddr(1, 2, 3, 4)

	// Device Not Found
	message := &pb_broker.DownlinkMessage{
		AppEui:  &appEUI,
		DevEui:  &devEUI,
		Payload: []byte{},
	}
	_, err := ns.HandleDownlink(message)
	a.So(err, ShouldNotBeNil)

	ns.devices.Set(&device.Device{
		DevAddr: devAddr,
		AppEUI:  appEUI,
		DevEUI:  devEUI,
	})

	// Invalid Payload
	message = &pb_broker.DownlinkMessage{
		AppEui:  &appEUI,
		DevEui:  &devEUI,
		Payload: []byte{},
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
			},
		},
	}
	bytes, _ := phy.MarshalBinary()

	message = &pb_broker.DownlinkMessage{
		AppEui:  &appEUI,
		DevEui:  &devEUI,
		Payload: bytes,
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

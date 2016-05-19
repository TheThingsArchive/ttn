package networkserver

import (
	"testing"

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

func TestHandleGetDevices(t *testing.T) {
	a := New(t)

	ns := &networkServer{
		devices: device.NewDeviceStore(),
	}

	// No Devices
	res, err := ns.HandleGetDevices(&pb.DevicesRequest{
		DevAddr: &types.DevAddr{1, 2, 3, 4},
		FCnt:    5,
	})
	a.So(err, ShouldBeNil)
	a.So(res.Results, ShouldBeEmpty)

	// Matching Device
	ns.devices.Set(&device.Device{
		DevAddr: types.DevAddr{1, 2, 3, 4},
		AppEUI:  types.AppEUI{1, 2, 3, 4, 5, 6, 7, 8},
		DevEUI:  types.DevEUI{1, 2, 3, 4, 5, 6, 7, 8},
		FCntUp:  5,
	})
	res, err = ns.HandleGetDevices(&pb.DevicesRequest{
		DevAddr: &types.DevAddr{1, 2, 3, 4},
		FCnt:    5,
	})
	a.So(err, ShouldBeNil)
	a.So(res.Results, ShouldHaveLength, 1)

	// Non-Matching DevAddr
	res, err = ns.HandleGetDevices(&pb.DevicesRequest{
		DevAddr: &types.DevAddr{5, 6, 7, 8},
		FCnt:    5,
	})
	a.So(err, ShouldBeNil)
	a.So(res.Results, ShouldHaveLength, 0)

	// Non-Matching FCnt
	res, err = ns.HandleGetDevices(&pb.DevicesRequest{
		DevAddr: &types.DevAddr{1, 2, 3, 4},
		FCnt:    4,
	})
	a.So(err, ShouldBeNil)
	a.So(res.Results, ShouldHaveLength, 0)

	// Non-Matching FCnt, but FCnt Check Disabled
	ns.devices.Set(&device.Device{
		DevAddr: types.DevAddr{5, 6, 7, 8},
		AppEUI:  types.AppEUI{5, 6, 7, 8, 1, 2, 3, 4},
		DevEUI:  types.DevEUI{5, 6, 7, 8, 1, 2, 3, 4},
		FCntUp:  5,
		Options: device.Options{
			DisableFCntCheck: true,
		},
	})
	res, err = ns.HandleGetDevices(&pb.DevicesRequest{
		DevAddr: &types.DevAddr{5, 6, 7, 8},
		FCnt:    4,
	})
	a.So(err, ShouldBeNil)
	a.So(res.Results, ShouldHaveLength, 1)

	// 32 Bit Frame Counter
	ns.devices.Set(&device.Device{
		DevAddr: types.DevAddr{2, 2, 3, 4},
		AppEUI:  types.AppEUI{2, 2, 3, 4, 5, 6, 7, 8},
		DevEUI:  types.DevEUI{2, 2, 3, 4, 5, 6, 7, 8},
		FCntUp:  5 + (2 << 16),
		Options: device.Options{
			Uses32BitFCnt: true,
		},
	})
	res, err = ns.HandleGetDevices(&pb.DevicesRequest{
		DevAddr: &types.DevAddr{2, 2, 3, 4},
		FCnt:    5,
	})
	a.So(err, ShouldBeNil)
	a.So(res.Results, ShouldHaveLength, 1)

}

func TestHandlePrepareActivation(t *testing.T) {
	a := New(t)
	ns := &networkServer{}
	resp, err := ns.HandlePrepareActivation(&pb_broker.DeduplicatedDeviceActivationRequest{})
	a.So(err, ShouldBeNil)
	devAddr := resp.ActivationMetadata.GetLorawan().DevAddr
	a.So(devAddr.IsEmpty(), ShouldBeFalse)
	a.So(devAddr[0]&254, ShouldEqual, 19<<1) // 7 MSB should be NetID
}

func TestHandleActivate(t *testing.T) {
	a := New(t)
	ns := &networkServer{
		devices: device.NewDeviceStore(),
	}

	_, err := ns.HandleActivate(&pb_handler.DeviceActivationResponse{})
	a.So(err, ShouldNotBeNil)

	_, err = ns.HandleActivate(&pb_handler.DeviceActivationResponse{
		ActivationMetadata: &pb_protocol.ActivationMetadata{},
	})
	a.So(err, ShouldNotBeNil)

	_, err = ns.HandleActivate(&pb_handler.DeviceActivationResponse{
		ActivationMetadata: &pb_protocol.ActivationMetadata{Protocol: &pb_protocol.ActivationMetadata_Lorawan{
			Lorawan: &pb_lorawan.ActivationMetadata{
				AppEui:  &types.AppEUI{0, 0, 0, 0, 0, 0, 3, 1},
				DevEui:  &types.DevEUI{0, 0, 0, 0, 0, 0, 3, 1},
				DevAddr: &types.DevAddr{0, 0, 3, 1},
				NwkSKey: &types.NwkSKey{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 3, 1},
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

	// Device Not Found
	message := &pb_broker.DeduplicatedUplinkMessage{
		AppEui:  &types.AppEUI{1, 2, 3, 4, 5, 6, 7, 8},
		DevEui:  &types.DevEUI{1, 2, 3, 4, 5, 6, 7, 8},
		Payload: []byte{},
	}
	_, err := ns.HandleUplink(message)
	a.So(err, ShouldNotBeNil)

	ns.devices.Set(&device.Device{
		DevAddr: types.DevAddr{1, 2, 3, 4},
		AppEUI:  types.AppEUI{1, 2, 3, 4, 5, 6, 7, 8},
		DevEUI:  types.DevEUI{1, 2, 3, 4, 5, 6, 7, 8},
	})

	// Invalid Payload
	message = &pb_broker.DeduplicatedUplinkMessage{
		AppEui:  &types.AppEUI{1, 2, 3, 4, 5, 6, 7, 8},
		DevEui:  &types.DevEUI{1, 2, 3, 4, 5, 6, 7, 8},
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
		AppEui:  &types.AppEUI{1, 2, 3, 4, 5, 6, 7, 8},
		DevEui:  &types.DevEUI{1, 2, 3, 4, 5, 6, 7, 8},
		Payload: bytes,
	}
	res, err := ns.HandleUplink(message)
	a.So(err, ShouldBeNil)
	a.So(res.ResponseTemplate, ShouldNotBeNil)

	// Frame Counter should have been updated
	dev, _ := ns.devices.Get(types.AppEUI{1, 2, 3, 4, 5, 6, 7, 8}, types.DevEUI{1, 2, 3, 4, 5, 6, 7, 8})
	a.So(dev.FCntUp, ShouldEqual, 1)
}

func TestHandleDownlink(t *testing.T) {
	a := New(t)
	ns := &networkServer{
		devices: device.NewDeviceStore(),
	}

	// Device Not Found
	message := &pb_broker.DownlinkMessage{
		AppEui:  &types.AppEUI{1, 2, 3, 4, 5, 6, 7, 8},
		DevEui:  &types.DevEUI{1, 2, 3, 4, 5, 6, 7, 8},
		Payload: []byte{},
	}
	_, err := ns.HandleDownlink(message)
	a.So(err, ShouldNotBeNil)

	ns.devices.Set(&device.Device{
		DevAddr: types.DevAddr{1, 2, 3, 4},
		AppEUI:  types.AppEUI{1, 2, 3, 4, 5, 6, 7, 8},
		DevEUI:  types.DevEUI{1, 2, 3, 4, 5, 6, 7, 8},
	})

	// Invalid Payload
	message = &pb_broker.DownlinkMessage{
		AppEui:  &types.AppEUI{1, 2, 3, 4, 5, 6, 7, 8},
		DevEui:  &types.DevEUI{1, 2, 3, 4, 5, 6, 7, 8},
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
		AppEui:  &types.AppEUI{1, 2, 3, 4, 5, 6, 7, 8},
		DevEui:  &types.DevEUI{1, 2, 3, 4, 5, 6, 7, 8},
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

	dev, _ := ns.devices.Get(types.AppEUI{1, 2, 3, 4, 5, 6, 7, 8}, types.DevEUI{1, 2, 3, 4, 5, 6, 7, 8})
	a.So(dev.FCntDown, ShouldEqual, 1)

}

// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package networkserver

import (
	"testing"

	pb "github.com/TheThingsNetwork/api/networkserver"
	"github.com/TheThingsNetwork/ttn/core/networkserver/device"
	"github.com/TheThingsNetwork/ttn/core/types"
	. "github.com/TheThingsNetwork/ttn/utils/testing"
	. "github.com/smartystreets/assertions"
)

func TestHandleGetDevices(t *testing.T) {
	a := New(t)

	ns := &networkServer{
		devices: device.NewRedisDeviceStore(GetRedisClient(), "ns-test-handle-get-devices"),
	}

	nwkSKey := types.NwkSKey{1, 2, 3, 4, 5, 6, 7, 8, 1, 2, 3, 4, 5, 6, 7, 8}

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
		NwkSKey: nwkSKey,
		FCntUp:  5,
	})
	defer func() {
		ns.devices.Delete(types.AppEUI(getEUI(1, 2, 3, 4, 5, 6, 7, 8)), types.DevEUI(getEUI(1, 2, 3, 4, 5, 6, 7, 8)))
	}()

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
		NwkSKey: nwkSKey,
		FCntUp:  5,
		Options: device.Options{
			DisableFCntCheck: true,
		},
	})
	defer func() {
		ns.devices.Delete(types.AppEUI(getEUI(5, 6, 7, 8, 1, 2, 3, 4)), types.DevEUI(getEUI(5, 6, 7, 8, 1, 2, 3, 4)))
	}()
	res, err = ns.HandleGetDevices(&pb.DevicesRequest{
		DevAddr: &devAddr2,
		FCnt:    4,
	})
	a.So(err, ShouldBeNil)
	a.So(res.Results, ShouldHaveLength, 1)

	// 32 Bit Frame Counter (A)
	ns.devices.Set(&device.Device{
		DevAddr: getDevAddr(2, 2, 3, 4),
		AppEUI:  types.AppEUI(getEUI(2, 2, 3, 4, 5, 6, 7, 8)),
		DevEUI:  types.DevEUI(getEUI(2, 2, 3, 4, 5, 6, 7, 8)),
		NwkSKey: nwkSKey,
		FCntUp:  5 + (2 << 16),
		Options: device.Options{
			Uses32BitFCnt: true,
		},
	})
	defer func() {
		ns.devices.Delete(types.AppEUI(getEUI(2, 2, 3, 4, 5, 6, 7, 8)), types.DevEUI(getEUI(2, 2, 3, 4, 5, 6, 7, 8)))
	}()
	devAddr3 := getDevAddr(2, 2, 3, 4)
	res, err = ns.HandleGetDevices(&pb.DevicesRequest{
		DevAddr: &devAddr3,
		FCnt:    5,
	})
	a.So(err, ShouldBeNil)
	a.So(res.Results, ShouldHaveLength, 1)

	// 32 Bit Frame Counter (B)
	ns.devices.Set(&device.Device{
		DevAddr: getDevAddr(2, 2, 3, 5),
		AppEUI:  types.AppEUI(getEUI(2, 2, 3, 4, 5, 3, 7, 8)),
		DevEUI:  types.DevEUI(getEUI(2, 2, 3, 4, 5, 3, 7, 8)),
		NwkSKey: nwkSKey,
		FCntUp:  (2 << 16) - 1,
		Options: device.Options{
			Uses32BitFCnt: true,
		},
	})
	defer func() {
		ns.devices.Delete(types.AppEUI(getEUI(2, 2, 3, 4, 5, 3, 7, 8)), types.DevEUI(getEUI(2, 2, 3, 4, 5, 3, 7, 8)))
	}()
	devAddr4 := getDevAddr(2, 2, 3, 5)
	res, err = ns.HandleGetDevices(&pb.DevicesRequest{
		DevAddr: &devAddr4,
		FCnt:    5,
	})
	a.So(err, ShouldBeNil)
	a.So(res.Results, ShouldHaveLength, 1)

}

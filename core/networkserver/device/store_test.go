// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package device

import (
	"testing"

	"github.com/TheThingsNetwork/ttn/core/types"
	. "github.com/TheThingsNetwork/ttn/utils/testing"
	. "github.com/smartystreets/assertions"
)

func TestDeviceStore(t *testing.T) {
	a := New(t)

	NewRedisDeviceStore(GetRedisClient(), "")

	s := NewRedisDeviceStore(GetRedisClient(), "networkserver-test-device-store")

	// Non-existing App
	err := s.Set(&Device{
		DevAddr: types.DevAddr{0, 0, 0, 1},
		DevEUI:  types.DevEUI{0, 0, 0, 0, 0, 0, 0, 1},
		AppEUI:  types.AppEUI{0, 0, 0, 0, 0, 0, 0, 1},
		NwkSKey: types.NwkSKey{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 5, 1},
	})
	a.So(err, ShouldBeNil)

	dev, err := s.Get(types.AppEUI{0, 0, 0, 0, 0, 0, 0, 1}, types.DevEUI{0, 0, 0, 0, 0, 0, 0, 1})
	a.So(err, ShouldBeNil)
	a.So(dev.DevAddr, ShouldEqual, types.DevAddr{0, 0, 0, 1})

	defer func() {
		s.Delete(types.AppEUI{0, 0, 0, 0, 0, 0, 0, 1}, types.DevEUI{0, 0, 0, 0, 0, 0, 0, 1})
	}()

	// Existing App
	err = s.Set(&Device{
		DevAddr: types.DevAddr{0, 0, 0, 1},
		DevEUI:  types.DevEUI{0, 0, 0, 0, 0, 0, 0, 2},
		AppEUI:  types.AppEUI{0, 0, 0, 0, 0, 0, 0, 1},
		NwkSKey: types.NwkSKey{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 5, 1},
	})
	a.So(err, ShouldBeNil)

	defer func() {
		s.Delete(types.AppEUI{0, 0, 0, 0, 0, 0, 0, 1}, types.DevEUI{0, 0, 0, 0, 0, 0, 0, 2})
	}()

	res, err := s.ListForAddress(types.DevAddr{0, 0, 0, 1})
	a.So(err, ShouldBeNil)
	a.So(res, ShouldHaveLength, 2)
	res, err = s.ListForAddress(types.DevAddr{0, 0, 0, 2})
	a.So(err, ShouldBeNil)
	a.So(res, ShouldHaveLength, 0)

	// Existing Device, New DevAddr
	err = s.Set(&Device{
		old: &Device{
			DevAddr: types.DevAddr{0, 0, 0, 1},
			DevEUI:  types.DevEUI{0, 0, 0, 0, 0, 0, 0, 2},
			AppEUI:  types.AppEUI{0, 0, 0, 0, 0, 0, 0, 1},
		},
		DevAddr: types.DevAddr{0, 0, 0, 3},
		DevEUI:  types.DevEUI{0, 0, 0, 0, 0, 0, 0, 2},
		AppEUI:  types.AppEUI{0, 0, 0, 0, 0, 0, 0, 1},
		NwkSKey: types.NwkSKey{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 5, 1},
	})
	a.So(err, ShouldBeNil)

	res, err = s.ListForAddress(types.DevAddr{0, 0, 0, 3})
	a.So(err, ShouldBeNil)
	a.So(res, ShouldHaveLength, 1)
	res, err = s.ListForAddress(types.DevAddr{0, 0, 0, 1})
	a.So(err, ShouldBeNil)
	a.So(res, ShouldHaveLength, 1)

	s.Set(&Device{
		old: &Device{
			DevAddr: types.DevAddr{0, 0, 0, 1},
			DevEUI:  types.DevEUI{0, 0, 0, 0, 0, 0, 0, 1},
			AppEUI:  types.AppEUI{0, 0, 0, 0, 0, 0, 0, 1},
		},
		DevAddr: types.DevAddr{0, 0, 0, 3},
		DevEUI:  types.DevEUI{0, 0, 0, 0, 0, 0, 0, 1},
		AppEUI:  types.AppEUI{0, 0, 0, 0, 0, 0, 0, 1},
		NwkSKey: types.NwkSKey{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 5, 2},
	})

	res, err = s.ListForAddress(types.DevAddr{0, 0, 0, 1})
	a.So(err, ShouldBeNil)
	a.So(res, ShouldHaveLength, 0)
	res, err = s.ListForAddress(types.DevAddr{0, 0, 0, 3})
	a.So(err, ShouldBeNil)
	a.So(res, ShouldHaveLength, 2)

	dev, err = s.Get(types.AppEUI{0, 0, 0, 0, 0, 0, 0, 2}, types.DevEUI{0, 0, 0, 0, 0, 0, 0, 2})
	a.So(err, ShouldNotBeNil)
	a.So(dev, ShouldBeNil)

	dev, err = s.Get(types.AppEUI{0, 0, 0, 0, 0, 0, 0, 1}, types.DevEUI{0, 0, 0, 0, 0, 0, 0, 3})
	a.So(err, ShouldNotBeNil)
	a.So(dev, ShouldBeNil)

	dev, err = s.Get(types.AppEUI{0, 0, 0, 0, 0, 0, 0, 1}, types.DevEUI{0, 0, 0, 0, 0, 0, 0, 1})
	a.So(err, ShouldBeNil)
	a.So(dev.DevAddr, ShouldEqual, types.DevAddr{0, 0, 0, 3})

	// List
	devices, err := s.List(nil)
	a.So(err, ShouldBeNil)
	a.So(devices, ShouldHaveLength, 2)

	err = s.Delete(types.AppEUI{0, 0, 0, 0, 0, 0, 0, 1}, types.DevEUI{0, 0, 0, 0, 0, 0, 0, 1})
	a.So(err, ShouldBeNil)

	res, err = s.ListForAddress(types.DevAddr{0, 0, 0, 3})
	a.So(err, ShouldBeNil)
	a.So(res, ShouldHaveLength, 1)
}

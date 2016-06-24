// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package device

import (
	"testing"

	"gopkg.in/redis.v3"

	"github.com/TheThingsNetwork/ttn/core/types"
	. "github.com/smartystreets/assertions"
)

func getRedisClient() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       1,  // use default DB
	})
}

func TestDeviceStore(t *testing.T) {
	a := New(t)

	stores := map[string]Store{
		"local": NewDeviceStore(),
		"redis": NewRedisDeviceStore(getRedisClient()),
	}

	for name, s := range stores {

		t.Logf("Testing %s store", name)

		// Non-existing App
		err := s.Set(&Device{
			DevAddr: types.DevAddr{0, 0, 0, 1},
			DevEUI:  types.DevEUI{0, 0, 0, 0, 0, 0, 0, 1},
			AppEUI:  types.AppEUI{0, 0, 0, 0, 0, 0, 0, 1},
			NwkSKey: types.NwkSKey{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 5, 1},
		})
		a.So(err, ShouldBeNil)

		// Existing App
		err = s.Set(&Device{
			DevAddr: types.DevAddr{0, 0, 0, 1},
			DevEUI:  types.DevEUI{0, 0, 0, 0, 0, 0, 0, 2},
			AppEUI:  types.AppEUI{0, 0, 0, 0, 0, 0, 0, 1},
			NwkSKey: types.NwkSKey{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 5, 1},
		})
		a.So(err, ShouldBeNil)

		res, err := s.GetWithAddress(types.DevAddr{0, 0, 0, 1})
		a.So(err, ShouldBeNil)
		a.So(res, ShouldHaveLength, 2)
		res, err = s.GetWithAddress(types.DevAddr{0, 0, 0, 2})
		a.So(err, ShouldBeNil)
		a.So(res, ShouldHaveLength, 0)

		// Existing Device, New DevAddr
		err = s.Set(&Device{
			DevAddr: types.DevAddr{0, 0, 0, 3},
			DevEUI:  types.DevEUI{0, 0, 0, 0, 0, 0, 0, 2},
			AppEUI:  types.AppEUI{0, 0, 0, 0, 0, 0, 0, 1},
			NwkSKey: types.NwkSKey{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 5, 1},
		})
		a.So(err, ShouldBeNil)

		res, err = s.GetWithAddress(types.DevAddr{0, 0, 0, 3})
		a.So(err, ShouldBeNil)
		a.So(res, ShouldHaveLength, 1)
		res, err = s.GetWithAddress(types.DevAddr{0, 0, 0, 1})
		a.So(err, ShouldBeNil)
		a.So(res, ShouldHaveLength, 1)

		s.Set(&Device{
			DevAddr: types.DevAddr{0, 0, 0, 3},
			DevEUI:  types.DevEUI{0, 0, 0, 0, 0, 0, 0, 1},
			AppEUI:  types.AppEUI{0, 0, 0, 0, 0, 0, 0, 1},
			NwkSKey: types.NwkSKey{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 5, 2},
		})

		res, err = s.GetWithAddress(types.DevAddr{0, 0, 0, 1})
		a.So(err, ShouldBeNil)
		a.So(res, ShouldHaveLength, 0)
		res, err = s.GetWithAddress(types.DevAddr{0, 0, 0, 3})
		a.So(err, ShouldBeNil)
		a.So(res, ShouldHaveLength, 2)

		dev, err := s.Get(types.AppEUI{0, 0, 0, 0, 0, 0, 0, 2}, types.DevEUI{0, 0, 0, 0, 0, 0, 0, 2})
		a.So(err, ShouldNotBeNil)
		a.So(dev, ShouldBeNil)

		dev, err = s.Get(types.AppEUI{0, 0, 0, 0, 0, 0, 0, 1}, types.DevEUI{0, 0, 0, 0, 0, 0, 0, 3})
		a.So(err, ShouldNotBeNil)
		a.So(dev, ShouldBeNil)

		dev, err = s.Get(types.AppEUI{0, 0, 0, 0, 0, 0, 0, 1}, types.DevEUI{0, 0, 0, 0, 0, 0, 0, 1})
		a.So(err, ShouldBeNil)
		a.So(dev.DevAddr, ShouldEqual, types.DevAddr{0, 0, 0, 3})

		// List
		devices, err := s.List()
		a.So(err, ShouldBeNil)
		a.So(devices, ShouldHaveLength, 2)

		err = s.Delete(types.AppEUI{0, 0, 0, 0, 0, 0, 0, 1}, types.DevEUI{0, 0, 0, 0, 0, 0, 0, 1})
		a.So(err, ShouldBeNil)

		res, err = s.GetWithAddress(types.DevAddr{0, 0, 0, 3})
		a.So(err, ShouldBeNil)
		a.So(res, ShouldHaveLength, 1)

		// Cleanup
		s.Delete(types.AppEUI{0, 0, 0, 0, 0, 0, 0, 1}, types.DevEUI{0, 0, 0, 0, 0, 0, 0, 2})
	}

}

func TestDeviceActivate(t *testing.T) {

	a := New(t)

	stores := map[string]Store{
		"local": NewDeviceStore(),
		"redis": NewRedisDeviceStore(getRedisClient()),
	}

	for name, s := range stores {

		t.Logf("Testing %s store", name)

		// Device not registered
		err := s.Activate(
			types.AppEUI{0, 0, 0, 0, 0, 0, 1, 1},
			types.DevEUI{0, 0, 0, 0, 0, 0, 1, 1},
			types.DevAddr{0, 0, 1, 1},
			types.NwkSKey{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 1},
		)
		a.So(err, ShouldNotBeNil)

		// Device registered
		s.Set(&Device{
			AppEUI: types.AppEUI{0, 0, 0, 0, 0, 0, 1, 1},
			DevEUI: types.DevEUI{0, 0, 0, 0, 0, 0, 1, 1},
		})
		err = s.Activate(
			types.AppEUI{0, 0, 0, 0, 0, 0, 1, 1},
			types.DevEUI{0, 0, 0, 0, 0, 0, 1, 1},
			types.DevAddr{0, 0, 1, 1},
			types.NwkSKey{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 1},
		)
		a.So(err, ShouldBeNil)

		// It should register the device
		dev, err := s.Get(types.AppEUI{0, 0, 0, 0, 0, 0, 1, 1}, types.DevEUI{0, 0, 0, 0, 0, 0, 1, 1})
		a.So(err, ShouldBeNil)
		a.So(dev.DevAddr, ShouldEqual, types.DevAddr{0, 0, 1, 1})

		// It should register the DevAddr
		res, err := s.GetWithAddress(types.DevAddr{0, 0, 1, 1})
		a.So(err, ShouldBeNil)
		a.So(res, ShouldHaveLength, 1)

		s.Set(&Device{
			DevAddr:  types.DevAddr{0, 0, 1, 1},
			DevEUI:   types.DevEUI{0, 0, 0, 0, 0, 0, 1, 1},
			AppEUI:   types.AppEUI{0, 0, 0, 0, 0, 0, 1, 1},
			FCntUp:   42,
			FCntDown: 42,
		})

		// Activate the same device again
		err = s.Activate(
			types.AppEUI{0, 0, 0, 0, 0, 0, 1, 1},
			types.DevEUI{0, 0, 0, 0, 0, 0, 1, 1},
			types.DevAddr{0, 0, 1, 2},
			types.NwkSKey{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 2},
		)
		a.So(err, ShouldBeNil)

		// It should reset the DevAddr, NwkSKey and Frame Counters
		dev, err = s.Get(types.AppEUI{0, 0, 0, 0, 0, 0, 1, 1}, types.DevEUI{0, 0, 0, 0, 0, 0, 1, 1})
		a.So(err, ShouldBeNil)
		a.So(dev.DevAddr, ShouldEqual, types.DevAddr{0, 0, 1, 2})
		a.So(dev.NwkSKey, ShouldEqual, types.NwkSKey{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 2})
		a.So(dev.FCntUp, ShouldEqual, 0)
		a.So(dev.FCntDown, ShouldEqual, 0)

		// Cleanup
		s.Delete(types.AppEUI{0, 0, 0, 0, 0, 0, 1, 1}, types.DevEUI{0, 0, 0, 0, 0, 0, 1, 1})

	}

}

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

		// Get non-existing
		dev, err := s.Get(types.AppEUI{0, 0, 0, 0, 0, 0, 0, 1}, types.DevEUI{0, 0, 0, 0, 0, 0, 0, 1})
		a.So(err, ShouldNotBeNil)
		a.So(dev, ShouldBeNil)

		// Create
		err = s.Set(&Device{
			DevAddr: types.DevAddr{0, 0, 0, 1},
			DevEUI:  types.DevEUI{0, 0, 0, 0, 0, 0, 0, 1},
			AppEUI:  types.AppEUI{0, 0, 0, 0, 0, 0, 0, 1},
		})
		a.So(err, ShouldBeNil)

		// Get existing
		dev, err = s.Get(types.AppEUI{0, 0, 0, 0, 0, 0, 0, 1}, types.DevEUI{0, 0, 0, 0, 0, 0, 0, 1})
		a.So(err, ShouldBeNil)
		a.So(dev, ShouldNotBeNil)

		// Create extra
		err = s.Set(&Device{
			DevAddr: types.DevAddr{0, 0, 0, 2},
			DevEUI:  types.DevEUI{0, 0, 0, 0, 0, 0, 0, 2},
			AppEUI:  types.AppEUI{0, 0, 0, 0, 0, 0, 0, 1},
		})
		a.So(err, ShouldBeNil)

		// List
		devices, err := s.List()
		a.So(err, ShouldBeNil)
		a.So(devices, ShouldHaveLength, 2)

		// Delete
		err = s.Delete(types.AppEUI{0, 0, 0, 0, 0, 0, 0, 1}, types.DevEUI{0, 0, 0, 0, 0, 0, 0, 1})
		a.So(err, ShouldBeNil)

		// Get deleted
		dev, err = s.Get(types.AppEUI{0, 0, 0, 0, 0, 0, 0, 1}, types.DevEUI{0, 0, 0, 0, 0, 0, 0, 1})
		a.So(err, ShouldNotBeNil)
		a.So(dev, ShouldBeNil)

		// Cleanup
		s.Delete(types.AppEUI{0, 0, 0, 0, 0, 0, 0, 1}, types.DevEUI{0, 0, 0, 0, 0, 0, 0, 2})
	}

}

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

	s := NewRedisDeviceStore(GetRedisClient(), "handler-test-device-store")

	// Get non-existing
	dev, err := s.Get("AppID-1", "DevID-1")
	a.So(err, ShouldNotBeNil)
	a.So(dev, ShouldBeNil)

	devs, err := s.ListForApp("AppID-1", nil)
	a.So(err, ShouldBeNil)
	a.So(devs, ShouldHaveLength, 0)

	// Create
	err = s.Set(&Device{
		DevAddr: types.DevAddr([4]byte{0, 0, 0, 1}),
		DevEUI:  types.DevEUI([8]byte{0, 0, 0, 0, 0, 0, 0, 1}),
		AppEUI:  types.AppEUI([8]byte{0, 0, 0, 0, 0, 0, 0, 1}),
		AppID:   "AppID-1",
		DevID:   "DevID-1",
	})
	a.So(err, ShouldBeNil)

	defer func() {
		s.Delete("AppID-1", "DevID-1")
	}()

	// Get existing
	dev, err = s.Get("AppID-1", "DevID-1")
	a.So(err, ShouldBeNil)
	a.So(dev, ShouldNotBeNil)

	devs, err = s.ListForApp("AppID-1", nil)
	a.So(err, ShouldBeNil)
	a.So(devs, ShouldHaveLength, 1)

	count, err := s.CountForApp("AppID-1")
	a.So(err, ShouldBeNil)
	a.So(count, ShouldEqual, 1)

	count, err = s.Count()
	a.So(err, ShouldBeNil)
	a.So(count, ShouldEqual, 1)

	// Create extra and update
	dev = &Device{
		DevAddr: types.DevAddr([4]byte{0, 0, 0, 2}),
		DevEUI:  types.DevEUI([8]byte{0, 0, 0, 0, 0, 0, 0, 2}),
		AppEUI:  types.AppEUI([8]byte{0, 0, 0, 0, 0, 0, 0, 1}),
		AppID:   "AppID-1",
		DevID:   "DevID-2",
	}
	err = s.Set(dev)
	a.So(err, ShouldBeNil)

	err = s.Set(&Device{
		old:     dev,
		DevAddr: types.DevAddr([4]byte{0, 0, 0, 3}),
		DevEUI:  types.DevEUI([8]byte{0, 0, 0, 0, 0, 0, 0, 3}),
		AppEUI:  types.AppEUI([8]byte{0, 0, 0, 0, 0, 0, 0, 2}),
		AppID:   "AppID-1",
		DevID:   "DevID-2",
	})
	a.So(err, ShouldBeNil)

	dev, err = s.Get("AppID-1", "DevID-2")
	a.So(err, ShouldBeNil)
	a.So(dev, ShouldNotBeNil)
	a.So(dev.DevEUI, ShouldEqual, types.DevEUI([8]byte{0, 0, 0, 0, 0, 0, 0, 3}))

	defer func() {
		s.Delete("AppID-1", "DevID-2")
	}()

	// List
	devices, err := s.List(nil)
	a.So(err, ShouldBeNil)
	a.So(devices, ShouldHaveLength, 2)

	count, err = s.Count()
	a.So(err, ShouldBeNil)
	a.So(count, ShouldEqual, 2)

	// Delete
	err = s.Delete("AppID-1", "DevID-1")
	a.So(err, ShouldBeNil)

	// Get deleted
	dev, err = s.Get("AppID-1", "DevID-1")
	a.So(err, ShouldNotBeNil)
	a.So(dev, ShouldBeNil)

	devs, err = s.ListForApp("AppID-1", nil)
	a.So(err, ShouldBeNil)
	a.So(devs, ShouldHaveLength, 1)

	count, err = s.CountForApp("AppID-1")
	a.So(err, ShouldBeNil)
	a.So(count, ShouldEqual, 1)

	count, err = s.Count()
	a.So(err, ShouldBeNil)
	a.So(count, ShouldEqual, 1)

}

func TestRedisDeviceStore_attrControl(t *testing.T) {
	a := New(t)

	store := NewRedisDeviceStore(GetRedisClient(), "handler-test-builtin-attribute")

	testMap1 := map[string]string{
		"hello": "bonjour",
		"test":  "TeSt",
	}
	in := &Device{Attributes: testMap1}
	store.attrFilter(in)
	a.So(in.Attributes, ShouldNotBeNil)
	a.So(in.Attributes["hello"], ShouldEqual, testMap1["hello"])
	a.So(in.Attributes["test"], ShouldEqual, testMap1["test"])

	//Past limit of 5
	testMap2 := map[string]string{
		"hello":   "bonjour",
		"test":    "TeSt",
		"beer":    "cold",
		"weather": "hot",
		"heart":   "pique",
		"square":  "trefle",
	}
	in.Attributes = testMap2
	store.attrFilter(in)
	a.So(len(in.Attributes), ShouldEqual, 5)

	//Past limit of 5 and builtin attributes
	store.SetBuiltinAttrList("ttn-battery:ttn-Model")
	testMap3 := map[string]string{
		"hello":       "bonjour",
		"test":        "TeSt",
		"beer":        "cold",
		"weather":     "hot",
		"heart":       "pique",
		"square":      "trefle",
		"ttn-battery": "quatre-ving-dix pourcent",
	}
	m := make(map[string]string, len(testMap3))
	for key, val := range testMap3 {
		m[key] = val
	}
	in.Attributes = m
	store.attrFilter(in)
	a.So(len(in.Attributes), ShouldEqual, 6)
	a.So(in.Attributes["ttn-Battery"], ShouldEqual, testMap3["ttn-Battery"])
}

func TestHandlerManager_attrControlKeyValidation(t *testing.T) {
	a := New(t)

	store := NewRedisDeviceStore(GetRedisClient(), "handler-test-builtin-attribute")
	testMap1 := map[string]string{
		"Hello": "bonjour",
		"test":  "TeSt",
		"youknowsometimesyoujustwanttoputareallylongnametobesurepeoplewillknowwhatallthislittlebytemean": "1",
		"": "too short!",
	}

	in := &Device{Attributes: testMap1}
	store.attrFilter(in)
	a.So(in.Attributes, ShouldNotBeNil)
	a.So(in.Attributes["Hello"], ShouldBeEmpty)
	a.So(in.Attributes[""], ShouldBeEmpty)
	a.So(
		in.Attributes["youknowsometimesyoujustwanttoputareallylongnametobesurepeoplewillknowallthislittlebytemean"],
		ShouldBeEmpty)
	a.So(in.Attributes["test"], ShouldEqual, testMap1["test"])
}

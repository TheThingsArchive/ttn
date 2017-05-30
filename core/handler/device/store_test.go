// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package device

import (
	"testing"

	pb "github.com/TheThingsNetwork/ttn/api/handler"
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

func TestRedisDeviceStore_SetAttributesList(t *testing.T) {
	a := New(t)

	store := NewRedisDeviceStore(GetRedisClient(), "handler-test-attribute")
	store.SetAttributesList("ttn-light:ttn-Humidity")
	a.So(store.attributesKeys["ttn-light"], ShouldBeTrue)
	a.So(store.attributesKeys["ttn-humidity"], ShouldBeFalse)
}

func TestRedisDeviceStore_sortAttributes_Add(t *testing.T) {
	a := New(t)

	store := NewRedisDeviceStore(GetRedisClient(), "handler-test-attribute")

	testMap1 := []*pb.Attribute{
		{"hello", "bonjour"},
		{"test", "TeSt"},
	}
	in := &Device{Attributes: testMap1}
	store.sortAttributes(in)
	a.So(in.Attributes, ShouldNotBeNil)
	a.So(in.Attributes[0].Key, ShouldEqual, testMap1[0].Key)
	a.So(in.Attributes[0].Val, ShouldEqual, testMap1[0].Val)
	a.So(in.Attributes[1].Key, ShouldEqual, testMap1[1].Key)

	// Past limit of 5
	testMap2 := []*pb.Attribute{
		{"hello", "bonjour"},
		{"test", "TeSt"},
		{"beer", "cold"},
		{"weather", "hot"},
		{"heart", "pique"},
		{"square", "trefle"},
	}
	in.Attributes = testMap2
	store.sortAttributes(in)
	a.So(len(in.Attributes), ShouldEqual, 5)
	a.So(in.Attributes[0].Key, ShouldEqual, testMap2[2].Key)
	a.So(in.Attributes[4].Key, ShouldEqual, testMap2[3].Key)

	// Past limit of 5 attributesKeys
	store.SetAttributesList("ttn-light:ttn-humidity")
	testMap3 := []*pb.Attribute{
		{"ttn-light", "quatre-ving-dix pourcent"},
		{"hello", "bonjour"},
		{"test", "TeSt"},
		{"beer", "cold"},
		{"weather", "hot"},
		{"heart", "pique"},
		{"square", "trefle"},
	}
	in.Attributes = testMap3
	store.sortAttributes(in)
	a.So(len(in.Attributes), ShouldEqual, 6)
	attr := &pb.Attribute{}
	for _, v := range in.Attributes {
		if v.Key == "ttn-light" {
			attr = v
		}
	}
	a.So(attr.Key, ShouldEqual, testMap3[0].Key)
	a.So(attr.Val, ShouldEqual, testMap3[0].Val)
}

func TestRedisDeviceStore_sortAttributes_KeyValidation(t *testing.T) {
	a := New(t)

	store := NewRedisDeviceStore(GetRedisClient(), "handler-test-attribute")
	testMap1 := []*pb.Attribute{
		{"Hello", "bonjour"},
		{"test", "TeSt"},
		{"youknowsometimesyoujustwanttoputareallylongnametobesurepeoplewillknowwhatallthislittlebytemean", "1"},
		{"", "too short!"},
	}

	in := &Device{Attributes: testMap1}
	store.sortAttributes(in)
	a.So(in.Attributes, ShouldNotBeNil)
	a.So(len(in.Attributes), ShouldEqual, 1)
}

func TestRedisDeviceStore_sortAttributes_Remove(t *testing.T) {
	a := New(t)

	store := NewRedisDeviceStore(GetRedisClient(), "handler-test-attribute")

	testMap1 := []*pb.Attribute{
		{"hello", "bonjour"},
		{"test", "TeSt"},
	}
	testMapRm := []*pb.Attribute{
		{"hello", ""},
		{"test", ""},
	}
	in := &Device{Attributes: testMapRm, old: &Device{Attributes: testMap1}}
	store.sortAttributes(in)
	a.So(in.Attributes, ShouldBeEmpty)

	testMap2 := []*pb.Attribute{
		{"hello", "bonjour"},
		{"test", "TeSt"},
	}
	testMapRm2 := []*pb.Attribute{
		{"hello", ""},
		{"hello", "coucou"},
		{"test", ""},
	}
	in.Attributes = testMapRm2
	in.old.Attributes = testMap2
	store.sortAttributes(in)
	a.So(len(in.Attributes), ShouldEqual, 1)
}

func TestRedisDeviceStore_attributesAdd(t *testing.T) {
	a := New(t)

	store := NewRedisDeviceStore(GetRedisClient(), "handler-test-attribute")

	dev := &Device{
		Attributes: []*pb.Attribute{
			{"Hello", "Bonjour"},
			{"adios", "goodbye"},
			{"youknowsometimesyoujustwanttoputareallylongnametobesurepeoplewillknowwhatallthislittlebytemean", "longKey"},
			{"longval", "youknowsometimesyoujustwanttoputareallylongnametobesurepeoplewillknowwhatallthislittlebytemean" +
				"youknowsometimesyoujustwanttoputareallylongnametobesurepeoplewillknowwhatallthislittlebytemean" +
				"youknowsometimesyoujustwanttoputareallylongnametobesurepeoplewillknowwhatallthislittlebytemean"},
		},
	}
	store.sortAttributes(dev)
	a.So(dev.Attributes, ShouldNotBeNil)
	a.So(len(dev.Attributes), ShouldEqual, 1)
	a.So(dev.Attributes[0].Val, ShouldEqual, "goodbye")
}

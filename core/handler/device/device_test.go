// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package device

import (
	"testing"

	pb "github.com/TheThingsNetwork/ttn/api/handler"
	"github.com/TheThingsNetwork/ttn/core/types"
	. "github.com/smartystreets/assertions"
)

func TestDeviceUpdate(t *testing.T) {
	a := New(t)
	device := &Device{
		DevID: "Device",
	}
	device.StartUpdate()
	a.So(device.old.DevID, ShouldEqual, device.DevID)
}

func TestDeviceClone(t *testing.T) {
	a := New(t)
	device := &Device{
		DevID: "Device",
		CurrentDownlink: &types.DownlinkMessage{
			PayloadRaw: []byte{1, 2, 3, 4},
		},
	}
	new := device.Clone()
	a.So(new.old, ShouldBeNil)
	a.So(new.DevID, ShouldEqual, device.DevID)
	a.So(new.CurrentDownlink, ShouldNotEqual, device.CurrentDownlink)
	a.So(new.CurrentDownlink.PayloadRaw, ShouldResemble, device.CurrentDownlink.PayloadRaw)
}

func TestDeviceChangedFields(t *testing.T) {
	a := New(t)
	device := &Device{
		DevID: "Device",
	}
	device.StartUpdate()
	device.DevID = "NewDevID"

	a.So(device.ChangedFields(), ShouldHaveLength, 1)
	a.So(device.ChangedFields(), ShouldContain, "DevID")
}

func TestDeviceGetLoRaWAN(t *testing.T) {
	device := &Device{
		DevID: "Device",
	}
	device.GetLoRaWAN()
}

func TestDevice_MapOldBuiltin(t *testing.T) {
	a := New(t)

	device := &Device{
		old: nil,
	}
	m, _ := device.MapOldBuiltin(nil)
	a.So(m, ShouldBeNil)

	device.old = &Device{
		Builtin: nil,
	}
	m, _ = device.MapOldBuiltin(nil)
	a.So(m, ShouldBeEmpty)

	device = &Device{
		old: &Device{
			Builtin: []*pb.Attribute{
				{"Hello", "Bonjour"},
			},
		},
	}
	m, i := device.MapOldBuiltin(nil)
	if !a.So(m, ShouldNotBeNil) {
		return
	}
	a.So(m["Hello"], ShouldEqual, "Bonjour")
	a.So(i, ShouldEqual, 1)

	m, i = device.MapOldBuiltin(map[string]bool{"Hello": true})
	if !a.So(m, ShouldNotBeNil) {
		return
	}
	a.So(m["Hello"], ShouldEqual, "Bonjour")
	a.So(i, ShouldBeZeroValue)
}

func TestDevice_DeleteEmptyBuiltin(t *testing.T) {
	a := New(t)

	device := &Device{
		Builtin: nil,
	}
	m, d := device.DeleteEmptyBuiltin(nil, nil)
	a.So(device.Builtin, ShouldBeNil)
	a.So(m, ShouldBeNil)
	a.So(d, ShouldBeZeroValue)

	device.Builtin = []*pb.Attribute{}
	m, d = device.DeleteEmptyBuiltin(nil, nil)
	a.So(device.Builtin, ShouldBeEmpty)
	a.So(m, ShouldBeNil)
	a.So(d, ShouldBeZeroValue)

	device.Builtin = []*pb.Attribute{{"Hello", ""}}
	m, d = device.DeleteEmptyBuiltin(nil, nil)
	a.So(device.Builtin, ShouldBeEmpty)
	a.So(m, ShouldBeNil)
	a.So(d, ShouldBeZeroValue)

	device.Builtin = []*pb.Attribute{{"Hello", ""}}
	m, d = device.DeleteEmptyBuiltin(nil, map[string]string{"Hello": "Bonjour"})
	a.So(device.Builtin, ShouldBeEmpty)
	a.So(m, ShouldBeEmpty)
	a.So(d, ShouldEqual, 1)

	device.Builtin = []*pb.Attribute{{"Hello", ""}}
	m, d = device.DeleteEmptyBuiltin(map[string]bool{"Hello": true}, map[string]string{"Hello": "Bonjour"})
	a.So(device.Builtin, ShouldBeEmpty)
	a.So(m, ShouldBeEmpty)
	a.So(d, ShouldEqual, 0)

	device.Builtin = []*pb.Attribute{{"Hello", ""}, {"Adios", "Goodbye"}}
	m, d = device.DeleteEmptyBuiltin(nil, map[string]string{"Hello": "Bonjour"})
	a.So(m, ShouldBeEmpty)
	a.So(d, ShouldEqual, 1)
	a.So(device.Builtin[0].Val, ShouldEqual, "Goodbye")
}

func TestDevice_BuiltinFromMap(t *testing.T) {
	a := New(t)

	device := &Device{}

	testMap := map[string]string{"Hello": "Bonjour", "Adios": "Goodbye"}
	device.BuiltinFromMap(testMap)
	a.So(device.Builtin[0].Key, ShouldEqual, "Adios")
}

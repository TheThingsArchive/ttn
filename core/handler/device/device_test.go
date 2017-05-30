// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package device

import (
	"testing"

	pb_handler "github.com/TheThingsNetwork/ttn/api/handler"
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
	n := device.Clone()
	a.So(n.old, ShouldBeNil)
	a.So(n.DevID, ShouldEqual, device.DevID)
	a.So(n.CurrentDownlink, ShouldNotEqual, device.CurrentDownlink)
	a.So(n.CurrentDownlink.PayloadRaw, ShouldResemble, device.CurrentDownlink.PayloadRaw)
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

func TestDevice_MapOldAttributes(t *testing.T) {
	a := New(t)

	device := &Device{}
	m, _ := device.MapOldAttributes(nil)
	a.So(m, ShouldBeNil)

	device.old = &Device{
		Attributes: nil,
	}
	m, _ = device.MapOldAttributes(nil)
	a.So(m, ShouldBeEmpty)

	device = &Device{
		old: &Device{
			Attributes: []*pb_handler.Attribute{
				{"Hello", "Bonjour"},
			},
		},
	}
	m, i := device.MapOldAttributes(nil)
	if !a.So(m, ShouldNotBeNil) {
		return
	}
	a.So(m["Hello"], ShouldEqual, "Bonjour")
	a.So(i, ShouldEqual, 1)

	m, i = device.MapOldAttributes(map[string]bool{"Hello": true})
	if !a.So(m, ShouldNotBeNil) {
		return
	}
	a.So(m["Hello"], ShouldEqual, "Bonjour")
	a.So(i, ShouldBeZeroValue)
}

func TestDevice_AttributesFromMap(t *testing.T) {
	a := New(t)

	device := &Device{}

	testMap := map[string]string{"Hello": "Bonjour", "Adios": "Goodbye"}
	device.AttributesFromMap(testMap)
	a.So(device.Attributes[0].Key, ShouldEqual, "Adios")
}

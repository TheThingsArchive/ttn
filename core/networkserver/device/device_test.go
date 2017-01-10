// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package device

import (
	"testing"

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

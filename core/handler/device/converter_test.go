// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package device

import (
	"testing"
	"time"

	pb_handler "github.com/TheThingsNetwork/ttn/api/handler"
	"github.com/TheThingsNetwork/ttn/core/types"
	. "github.com/smartystreets/assertions"
)

var testDev = &Device{
	AppEUI: [8]byte{0x10},
	AppID:  "test-app",
	DevEUI: [8]byte{0x10},
	DevID:  "test-dev",

	Description: "testing",
	Latitude:    52.3746961,
	Longitude:   4.8285748,
	Altitude:    255,

	Options: Options{ActivationConstraints: "local"},

	AppKey:        [16]byte{0x10},
	UsedDevNonces: []DevNonce{},
	UsedAppNonces: []AppNonce{},

	DevAddr: types.DevAddr{byte(0x10)},
	NwkSKey: [16]byte{0x10},
	AppSKey: [16]byte{0x10},
	FCntUp:  255,

	CurrentDownlink: nil,

	CreatedAt: time.Now(),
	UpdatedAt: time.Now(),

	Attributes: []*pb_handler.Attribute{{"test", "test"}},
}

func TestDevice_ToPb(t *testing.T) {
	a := New(t)

	p := testDev.ToPb()
	a.So(p.AppId, ShouldEqual, testDev.AppID)
	a.So(p.DevId, ShouldEqual, testDev.DevID)
	a.So(p.Latitude, ShouldEqual, testDev.Latitude)
	a.So(p.Longitude, ShouldEqual, testDev.Longitude)
	a.So(p.Altitude, ShouldEqual, testDev.Altitude)
	a.So(p.Attributes[0].Val, ShouldEqual, testDev.Attributes[0].Val)
	a.So(p.Attributes[0].Key, ShouldEqual, testDev.Attributes[0].Key)
}

func TestDevice_FromPb(t *testing.T) {
	a := New(t)

	p := testDev.ToPb()
	dev := FromPb(p)
	a.So(dev.AppID, ShouldEqual, testDev.AppID)
	a.So(dev.DevID, ShouldEqual, testDev.DevID)
	a.So(dev.Latitude, ShouldEqual, testDev.Latitude)
	a.So(dev.Longitude, ShouldEqual, testDev.Longitude)
	a.So(dev.Altitude, ShouldEqual, testDev.Altitude)
	a.So(dev.Attributes[0].Val, ShouldEqual, testDev.Attributes[0].Val)
	a.So(dev.Attributes[0].Key, ShouldEqual, testDev.Attributes[0].Key)
}

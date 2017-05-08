// Copyright © 2017 The Things Industries B.V.

package device

import (
	"testing"

	"time"

	pb "github.com/TheThingsNetwork/ttn/api/handler"
	"github.com/TheThingsNetwork/ttn/core/types"

	. "github.com/smartystreets/assertions"
)

var test_dev = &Device{
	AppEUI: [8]byte{0x10},
	AppID:  "test-app",
	DevEUI: [8]byte{0x10},
	DevID:  "test-dev",

	Description: "testing",
	Latitude:    255,
	Longitude:   255,
	Altitude:    255,

	Options: Options{ActivationConstraints: "activate"},

	AppKey:        [16]byte{0x10},
	UsedDevNonces: []DevNonce{},
	UsedAppNonces: []AppNonce{},

	DevAddr: types.DevAddr{byte('E')},
	NwkSKey: [16]byte{0x10},
	AppSKey: [16]byte{0x10},
	FCntUp:  255,

	CurrentDownlink: nil,

	CreatedAt: time.Now(),
	UpdatedAt: time.Now(),

	Attributes: map[string]string{"test": "test"},
}

func TestDevice_ToPb(t *testing.T) {
	a := New(t)

	p := test_dev.ToPb()
	a.So(p.AppId, ShouldEqual, test_dev.AppID)
	a.So(p.DevId, ShouldEqual, test_dev.DevID)
	a.So(p.Latitude, ShouldEqual, test_dev.Latitude)
	a.So(p.Longitude, ShouldEqual, test_dev.Longitude)
	a.So(p.Altitude, ShouldEqual, test_dev.Altitude)
	a.So(p.Attributes, ShouldEqual, test_dev.Attributes)
}

func TestDevice_FromPb(t *testing.T) {
	a := New(t)

	p := test_dev.ToPb()
	dev := Device{}
	l := p.Device.(*pb.Device_LorawanDevice)
	lora := l.LorawanDevice
	dev.FromPb(p, lora)
	a.So(dev, ShouldEqual, test_dev)
}

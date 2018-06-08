// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package handler

import (
	"testing"
	"time"

	"github.com/TheThingsNetwork/ttn/core/handler/device"
	. "github.com/smartystreets/assertions"
)

func TestEventUpdatedFields(t *testing.T) {
	a := New(t)

	dev := &device.Device{
		AppID:     "test",
		DevID:     "test-dev",
		Latitude:  0.0,
		Longitude: 30.0,
		Altitude:  100,
		UpdatedAt: time.Now().Add(-time.Hour),
		UsedAppNonces: []device.AppNonce{
			[3]byte{00, 00, 01},
		},
		UsedDevNonces: []device.DevNonce{
			[2]byte{01, 00},
		},
		Attributes: map[string]string{
			"type": "testing",
		},
	}
	dev.StartUpdate()
	dev.Latitude = 10
	dev.Longitude = 35
	dev.Altitude = 25
	dev.UsedAppNonces = append(dev.UsedAppNonces, device.AppNonce{00, 00, 02})
	dev.UsedDevNonces = append(dev.UsedDevNonces, device.DevNonce{00, 02})
	dev.UpdatedAt = time.Now()
	eventUpdateData := eventUpdatedFields(dev)
	a.So(eventUpdateData.UpdatedAt.Equal(dev.UpdatedAt), ShouldBeTrue)
	a.So(eventUpdateData.Altitude, ShouldEqual, dev.Altitude)
	a.So(eventUpdateData.Longitude, ShouldEqual, dev.Longitude)
	a.So(eventUpdateData.Latitude, ShouldEqual, dev.Latitude)
}

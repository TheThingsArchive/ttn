// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package device

import (
	"testing"

	"github.com/TheThingsNetwork/ttn/core/types"
	"github.com/TheThingsNetwork/ttn/mqtt"
	. "github.com/smartystreets/assertions"
)

func getTestDevice() (device *Device, dmap map[string]string) {
	return &Device{
			DevEUI:        types.DevEUI{1, 2, 3, 4, 5, 6, 7, 8},
			AppEUI:        types.AppEUI{8, 7, 6, 5, 4, 3, 2, 1},
			AppID:         "AppID-1",
			DevAddr:       types.DevAddr{1, 2, 3, 4},
			AppKey:        types.AppKey{0, 1, 0, 2, 0, 3, 0, 4, 0, 5, 0, 6, 0, 7, 0, 8},
			NwkSKey:       types.NwkSKey{1, 1, 1, 2, 1, 3, 1, 4, 1, 5, 1, 6, 1, 7, 1, 8},
			AppSKey:       types.AppSKey{2, 1, 2, 2, 2, 3, 2, 4, 2, 5, 2, 6, 2, 7, 2, 8},
			UsedDevNonces: []DevNonce{DevNonce{1, 2}, DevNonce{3, 4}},
			UsedAppNonces: []AppNonce{AppNonce{1, 2, 3}, AppNonce{4, 5, 6}},
		}, map[string]string{
			"dev_eui":         "0102030405060708",
			"app_eui":         "0807060504030201",
			"app_id":          "AppID-1",
			"dev_addr":        "01020304",
			"app_key":         "00010002000300040005000600070008",
			"nwk_s_key":       "01010102010301040105010601070108",
			"app_s_key":       "02010202020302040205020602070208",
			"used_dev_nonces": "0102,0304",
			"used_app_nonces": "010203,040506",
			"next_downlink":   "null",
		}
}

func TestToStringMap(t *testing.T) {
	a := New(t)
	device, expected := getTestDevice()
	dmap, err := device.ToStringStringMap(DeviceProperties...)
	a.So(err, ShouldBeNil)
	a.So(dmap, ShouldResemble, expected)
}

func TestFromStringMap(t *testing.T) {
	a := New(t)
	device := &Device{}
	expected, dmap := getTestDevice()
	err := device.FromStringStringMap(dmap)
	a.So(err, ShouldBeNil)
	a.So(device, ShouldResemble, expected)
}

func TestNextDownlink(t *testing.T) {
	a := New(t)
	dev := &Device{
		NextDownlink: &mqtt.DownlinkMessage{
			Fields: map[string]interface{}{
				"string": "hello!",
				"int":    42,
				"bool":   true,
			},
		},
	}
	formatted, err := dev.formatProperty("next_downlink")
	a.So(err, ShouldBeNil)
	a.So(formatted, ShouldContainSubstring, `"fields":{"bool":true,"int":42,"string":"hello!"}`)

	dev = &Device{}
	err = dev.parseProperty("next_downlink", `{"fields":{"bool":true,"int":42,"string":"hello!"}}`)
	a.So(err, ShouldBeNil)
	a.So(dev.NextDownlink.Fields, ShouldNotBeNil)
	a.So(dev.NextDownlink.Fields["bool"], ShouldBeTrue)
	a.So(dev.NextDownlink.Fields["int"], ShouldEqual, 42)
	a.So(dev.NextDownlink.Fields["string"], ShouldEqual, "hello!")
}

// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package device

import (
	"testing"
	"time"

	"github.com/TheThingsNetwork/ttn/core/types"
	. "github.com/smartystreets/assertions"
)

func getTestDevice() (device *Device, dmap map[string]string) {
	return &Device{
			DevEUI:   types.DevEUI{1, 2, 3, 4, 5, 6, 7, 8},
			AppEUI:   types.AppEUI{8, 7, 6, 5, 4, 3, 2, 1},
			AppID:    "TestApp-1",
			DevAddr:  types.DevAddr{1, 2, 3, 4},
			NwkSKey:  types.NwkSKey{0, 1, 0, 2, 0, 3, 0, 4, 0, 5, 0, 6, 0, 7, 0, 8},
			FCntUp:   42,
			FCntDown: 24,
			LastSeen: time.Unix(0, 0).UTC(),
			Options:  Options{},
		}, map[string]string{
			"dev_eui":    "0102030405060708",
			"app_eui":    "0807060504030201",
			"app_id":     "TestApp-1",
			"dev_addr":   "01020304",
			"nwk_s_key":  "00010002000300040005000600070008",
			"f_cnt_up":   "42",
			"f_cnt_down": "24",
			"last_seen":  "1970-01-01T00:00:00Z",
			"options":    `{"disable_fcnt_check":false,"uses_32_bit_fcnt":false}`,
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

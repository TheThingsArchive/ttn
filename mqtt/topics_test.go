// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package mqtt

import (
	"testing"

	"github.com/TheThingsNetwork/ttn/core/types"
	. "github.com/smartystreets/assertions"
)

func TestParseDeviceTopic(t *testing.T) {
	a := New(t)

	topic := "0102030405060708/devices/0807060504030201/up"

	expected := &DeviceTopic{
		AppEUI: types.AppEUI{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08},
		DevEUI: types.DevEUI{0x08, 0x07, 0x06, 0x05, 0x04, 0x03, 0x02, 0x01},
		Type:   Uplink,
	}

	got, err := ParseDeviceTopic(topic)

	a.So(err, ShouldBeNil)
	a.So(got, ShouldResemble, expected)
}

func TestParseDeviceTopicInvalid(t *testing.T) {
	a := New(t)

	_, err := ParseDeviceTopic("000000000000000a/devices/0807060504030201/up") // AppEUI contains lowercase hex chars
	a.So(err, ShouldNotBeNil)

	_, err = ParseDeviceTopic("01020304050607/devices/0807060504030201/up") // AppEUI too short
	a.So(err, ShouldNotBeNil)

	_, err = ParseDeviceTopic("abcdefghijklmnop/devices/08070605040302/up") // AppEUI contains invalid characters
	a.So(err, ShouldNotBeNil)

	_, err = ParseDeviceTopic("0102030405060708/devices/000000000000000a/up") // DevEUI contains lowercase hex chars
	a.So(err, ShouldNotBeNil)

	_, err = ParseDeviceTopic("0102030405060708/devices/08070605040302/up") // DevEUI too short
	a.So(err, ShouldNotBeNil)

	_, err = ParseDeviceTopic("0102030405060708/devices/abcdefghijklmnop/up") // DevEUI contains invalid characters
	a.So(err, ShouldNotBeNil)

	_, err = ParseDeviceTopic("0102030405060708/fridges/0102030405060708/up") // We don't support fridges (at least, not specifically fridges)
	a.So(err, ShouldNotBeNil)

	_, err = ParseDeviceTopic("0102030405060708/devices/0102030405060708/emotions") // Devices usually don't publish emotions
	a.So(err, ShouldNotBeNil)
}

func TestTopicString(t *testing.T) {
	a := New(t)

	topic := &DeviceTopic{
		AppEUI: types.AppEUI{0x12, 0x34, 0x56, 0x78, 0x9a, 0xbc, 0xde, 0xf0},
		DevEUI: types.DevEUI{0x28, 0x27, 0x26, 0x25, 0x24, 0x23, 0x22, 0x21},
		Type:   Downlink,
	}

	expected := "123456789ABCDEF0/devices/2827262524232221/down"

	got := topic.String()

	a.So(got, ShouldResemble, expected)
}

func TestTopicParseAndString(t *testing.T) {
	a := New(t)

	expectedList := []string{
		// Uppercase (not lowercase)
		"0102030405060708/devices/ABCDABCD12345678/up",
		"0102030405060708/devices/ABCDABCD12345678/down",
		"0102030405060708/devices/ABCDABCD12345678/activations",
		// Numbers
		"0102030405060708/devices/0000000012345678/up",
		"0102030405060708/devices/0000000012345678/down",
		"0102030405060708/devices/0000000012345678/activations",
		// Wildcards
		"+/devices/+/up",
		"+/devices/+/down",
		"+/devices/+/activations",
		// Not Wildcard
		"0102030405060708/devices/0100000000000000/up",
		"0102030405060708/devices/0100000000000000/down",
		"0102030405060708/devices/0100000000000000/activations",
	}

	for _, expected := range expectedList {
		topic, err := ParseDeviceTopic(expected)
		a.So(err, ShouldBeNil)
		a.So(topic.String(), ShouldEqual, expected)
	}

}

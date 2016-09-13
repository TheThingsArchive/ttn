// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package mqtt

import (
	"testing"

	. "github.com/smartystreets/assertions"
)

func TestParseDeviceTopic(t *testing.T) {
	a := New(t)

	topic := "appid-1/devices/devid-1/up"

	expected := &DeviceTopic{
		AppID: "appid-1",
		DevID: "devid-1",
		Type:  Uplink,
	}

	got, err := ParseDeviceTopic(topic)

	a.So(err, ShouldBeNil)
	a.So(got, ShouldResemble, expected)
}

func TestParseDeviceTopicInvalid(t *testing.T) {
	a := New(t)

	_, err := ParseDeviceTopic("appid:Invalid/devices/dev/up")
	a.So(err, ShouldNotBeNil)

	_, err = ParseDeviceTopic("appid-1/devices/devid:Invalid/up") // DevEUI contains lowercase hex chars
	a.So(err, ShouldNotBeNil)

	_, err = ParseDeviceTopic("appid-1/fridges/devid-1/up") // We don't support fridges (at least, not specifically fridges)
	a.So(err, ShouldNotBeNil)

	_, err = ParseDeviceTopic("appid-1/devices/devid-1/emotions") // Devices usually don't publish emotions
	a.So(err, ShouldNotBeNil)
}

func TestTopicString(t *testing.T) {
	a := New(t)

	topic := &DeviceTopic{
		AppID: "appid-1",
		DevID: "devid-1",
		Type:  Downlink,
	}

	expected := "appid-1/devices/devid-1/down"

	got := topic.String()

	a.So(got, ShouldResemble, expected)
}

func TestTopicParseAndString(t *testing.T) {
	a := New(t)

	expectedList := []string{
		// Uppercase (not lowercase)
		"0102030405060708/devices/abcdabcd12345678/up",
		"0102030405060708/devices/abcdabcd12345678/down",
		"0102030405060708/devices/abcdabcd12345678/activations",
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

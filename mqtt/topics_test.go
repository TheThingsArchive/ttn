// Copyright © 2017 The Things Network
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
		Type:  DeviceUplink,
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

func TestDeviceTopicString(t *testing.T) {
	a := New(t)

	topic := &DeviceTopic{
		AppID: "appid-1",
		DevID: "devid-1",
		Type:  DeviceDownlink,
	}

	expected := "appid-1/devices/devid-1/down"

	got := topic.String()

	a.So(got, ShouldResemble, expected)
}

func TestDeviceTopicParseAndString(t *testing.T) {
	a := New(t)

	expectedList := []string{
		// Uppercase (not lowercase)
		"0102030405060708/devices/abcdabcd12345678/up",
		"0102030405060708/devices/abcdabcd12345678/up/value",
		"0102030405060708/devices/abcdabcd12345678/down",
		"0102030405060708/devices/abcdabcd12345678/events/activations",
		// Numbers
		"0102030405060708/devices/0000000012345678/up",
		"0102030405060708/devices/0000000012345678/up/value",
		"0102030405060708/devices/0000000012345678/down",
		"0102030405060708/devices/0000000012345678/events/activations",
		// Wildcards
		"+/devices/+/up",
		"+/devices/+/down",
		"+/devices/+/events/activations",
		// Not Wildcard
		"0102030405060708/devices/0100000000000000/up",
		"0102030405060708/devices/0100000000000000/up/value",
		"0102030405060708/devices/0100000000000000/down",
		"0102030405060708/devices/0100000000000000/events/activations",
	}

	for _, expected := range expectedList {
		topic, err := ParseDeviceTopic(expected)
		a.So(err, ShouldBeNil)
		a.So(topic.String(), ShouldEqual, expected)
	}

}

func TestParseAppTopicInvalid(t *testing.T) {
	a := New(t)

	_, err := ParseApplicationTopic("appid:Invalid/events")
	a.So(err, ShouldNotBeNil)

	_, err = ParseApplicationTopic("appid/randomstuff")
	a.So(err, ShouldNotBeNil)
}

func TestAppTopicString(t *testing.T) {
	a := New(t)

	topic := &ApplicationTopic{
		AppID: "appid-1",
		Type:  AppEvents,
	}

	a.So(topic.String(), ShouldResemble, "appid-1/events/#")

	topic = &ApplicationTopic{
		AppID: "appid-1",
		Type:  AppEvents,
		Field: "err",
	}

	a.So(topic.String(), ShouldResemble, "appid-1/events/err")
}

func TestAppTopicParseAndString(t *testing.T) {
	a := New(t)

	expectedList := []string{
		"+/events/#",
		"appid/events/#",
		"+/events/some-event",
		"appid/events/some-event",
		"+/events/some/event",
		"appid/events/some/event",
	}

	for _, expected := range expectedList {
		topic, err := ParseApplicationTopic(expected)
		a.So(err, ShouldBeNil)
		a.So(topic.String(), ShouldEqual, expected)
	}

}

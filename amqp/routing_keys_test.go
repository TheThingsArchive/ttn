// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package amqp

import (
	"testing"

	. "github.com/smartystreets/assertions"
)

func TestParseDeviceKey(t *testing.T) {
	a := New(t)

	key := "appid-1.devices.devid-1.up"

	expected := &DeviceKey{
		AppID: "appid-1",
		DevID: "devid-1",
		Type:  DeviceUplink,
	}

	got, err := ParseDeviceKey(key)

	a.So(err, ShouldBeNil)
	a.So(got, ShouldResemble, expected)
}

func TestParseDeviceKeyInvalid(t *testing.T) {
	a := New(t)

	_, err := ParseDeviceKey("appid:Invalid.devices.dev.up")
	a.So(err, ShouldNotBeNil)

	_, err = ParseDeviceKey("appid-1.devices.devid:Invalid.up") // DevID contains hex chars
	a.So(err, ShouldNotBeNil)

	_, err = ParseDeviceKey("appid-1.fridges.devid-1.up") // We don't support fridges (at least, not specifically fridges)
	a.So(err, ShouldNotBeNil)

	_, err = ParseDeviceKey("appid-1.devices.devid-1.emotions") // Devices usually don't publish emotions
	a.So(err, ShouldNotBeNil)
}

func TestDeviceKeyString(t *testing.T) {
	a := New(t)

	key := &DeviceKey{
		AppID: "appid-1",
		DevID: "devid-1",
		Type:  DeviceDownlink,
	}

	expected := "appid-1.devices.devid-1.down"

	got := key.String()

	a.So(got, ShouldResemble, expected)
}

func TestDeviceKeyParseAndString(t *testing.T) {
	a := New(t)

	expectedList := []string{
		// Uppercase (not lowercase)
		"0102030405060708.devices.abcdabcd12345678.up",
		"0102030405060708.devices.abcdabcd12345678.down",
		"0102030405060708.devices.abcdabcd12345678.events.activations",
		// Numbers
		"0102030405060708.devices.0000000012345678.up",
		"0102030405060708.devices.0000000012345678.down",
		"0102030405060708.devices.0000000012345678.events.activations",
		// Wildcards
		"*.devices.*.up",
		"*.devices.*.down",
		"*.devices.*.events.activations",
		// Not Wildcard
		"0102030405060708.devices.0100000000000000.up",
		"0102030405060708.devices.0100000000000000.down",
		"0102030405060708.devices.0100000000000000.events.activations",
	}

	for _, expected := range expectedList {
		key, err := ParseDeviceKey(expected)
		a.So(err, ShouldBeNil)
		a.So(key.String(), ShouldEqual, expected)
	}
}

func TestParseAppKey(t *testing.T) {
	a := New(t)

	key := "appid-1.devices.devid-1.up"

	expected := &DeviceKey{
		AppID: "appid-1",
		DevID: "devid-1",
		Type:  DeviceUplink,
	}

	got, err := ParseDeviceKey(key)

	a.So(err, ShouldBeNil)
	a.So(got, ShouldResemble, expected)
}

func TestParseAppKeyInvalid(t *testing.T) {
	a := New(t)

	_, err := ParseApplicationKey("appid:Invalid.events")
	a.So(err, ShouldNotBeNil)

	_, err = ParseApplicationKey("appid.randomstuff")
	a.So(err, ShouldNotBeNil)
}

func TestAppKeyString(t *testing.T) {
	a := New(t)

	key := &ApplicationKey{
		AppID: "appid-1",
		Type:  AppEvents,
	}

	a.So(key.String(), ShouldResemble, "appid-1.events.#")

	key = &ApplicationKey{
		AppID: "appid-1",
		Type:  AppEvents,
		Field: "err",
	}

	a.So(key.String(), ShouldResemble, "appid-1.events.err")
}

func TestAppKeyParseAndString(t *testing.T) {
	a := New(t)

	expectedList := []string{
		"*.events.#",
		"appid.events.#",
		"*.events.some-event",
		"appid.events.some-event",
		"*.events.some.event",
		"appid.events.some.event",
	}

	for _, expected := range expectedList {
		key, err := ParseApplicationKey(expected)
		a.So(err, ShouldBeNil)
		a.So(key.String(), ShouldEqual, expected)
	}
}

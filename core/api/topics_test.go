// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package api

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestGetType(t *testing.T) {
	Convey("Given the topic 0101/somevalue", t, func() {
		topic := "0101/somevalue"
		Convey("Then the type is somevalue", func() {
			tp, err := GetTopicType(topic)
			So(err, ShouldBeNil)
			So(tp, ShouldEqual, "somevalue")
		})
	})

	Convey("Given the topic 0101/devices/AA/up", t, func() {
		topic := "0101/devices/AA/up"
		Convey("Then the type is devices", func() {
			tp, err := GetTopicType(topic)
			So(err, ShouldBeNil)
			So(tp, ShouldEqual, Devices)
		})
	})

	Convey("Given the topic 0101", t, func() {
		topic := "0101"
		Convey("Then the type cannot be determined", func() {
			_, err := GetTopicType(topic)
			So(err, ShouldNotBeNil)
		})
	})
}

func TestDecodeDeviceTopic(t *testing.T) {
	Convey("Given the topic 0101/devices/AA/up", t, func() {
		topic := "0101/devices/AA/up"
		Convey("Then the type is devices", func() {
			tp, err := GetTopicType(topic)
			So(err, ShouldBeNil)
			So(tp, ShouldEqual, Devices)
		})

		Convey("Then the device topic can be decoded", func() {
			tp, err := DecodeDeviceTopic(topic)
			So(err, ShouldBeNil)

			Convey("And appEui is 0101, devEui is AA and type is Uplink", func() {
				So(tp.AppEUI, ShouldEqual, "0101")
				So(tp.DevEUI, ShouldEqual, "AA")
				So(tp.Type, ShouldEqual, Uplink)
			})
		})
	})

	Convey("Given the topic 0101/devices", t, func() {
		topic := "0101/devices"
		Convey("Then the device topic cannot be decoded", func() {
			_, err := DecodeDeviceTopic(topic)
			So(err, ShouldNotBeNil)
		})
	})

	Convey("Given the topic 0101/aa/bb/cc", t, func() {
		topic := "0101/aa/bb/cc"
		Convey("Then the device topic cannot be decoded", func() {
			_, err := DecodeDeviceTopic(topic)
			So(err, ShouldNotBeNil)
		})
	})
}

func TestEncodeDeviceTopic(t *testing.T) {
	Convey("Given the appEui 0202, devEui BB and type Downlink", t, func() {
		tp := &DeviceTopic{
			AppEUI: "0202",
			DevEUI: "BB",
			Type:   Downlink,
		}

		Convey("Then the topic is 0202/devices/BB/down", func() {
			topic, err := tp.Encode()
			So(err, ShouldBeNil)
			So(topic, ShouldEqual, "0202/devices/BB/down")
		})
	})
}

func TestRoundtrip(t *testing.T) {
	Convey("Given the input 0303/devices/CC/activations", t, func() {
		input := "0303/devices/CC/activations"

		Convey("Then the encoded output is the same as the decoded input", func() {
			tp, err := DecodeDeviceTopic(input)
			So(err, ShouldBeNil)

			output, err := tp.Encode()
			So(err, ShouldBeNil)
			So(output, ShouldEqual, input)
		})
	})
}

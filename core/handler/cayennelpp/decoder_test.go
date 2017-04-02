// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package cayennelpp

import (
	"testing"

	protocol "github.com/TheThingsNetwork/go-cayenne-lib/cayennelpp"
	. "github.com/smartystreets/assertions"
)

func TestDecode(t *testing.T) {
	a := New(t)

	buf := []byte{
		1, protocol.DigitalInput, 255,
		2, protocol.DigitalOutput, 100,
		3, protocol.AnalogInput, 21, 74,
		4, protocol.AnalogOutput, 234, 182,
		5, protocol.Luminosity, 1, 244,
		6, protocol.Presence, 50,
		7, protocol.Temperature, 255, 100,
		8, protocol.RelativeHumidity, 99,
		9, protocol.Accelerometer, 254, 88, 0, 15, 6, 130,
		10, protocol.BarometricPressure, 41, 239,
		11, protocol.Gyrometer, 1, 99, 2, 49, 254, 102,
		12, protocol.GPS, 7, 253, 135, 0, 190, 245, 0, 8, 106,
	}

	decoder := new(Decoder)
	fields, valid, err := decoder.Decode(buf, 1)
	a.So(err, ShouldBeNil)
	a.So(valid, ShouldBeTrue)
	a.So(fields, ShouldHaveLength, 12)
	a.So(fields["digital_in_1"], ShouldEqual, 255)
	a.So(fields["digital_out_2"], ShouldEqual, 100)
	a.So(fields["analog_in_3"], ShouldEqual, 54.5)
	a.So(fields["analog_out_4"], ShouldEqual, -54.5)
	a.So(fields["luminosity_5"], ShouldEqual, 500)
	a.So(fields["presence_6"], ShouldEqual, 50)
	a.So(fields["temperature_7"], ShouldEqual, -15.6)
	a.So(fields["relative_humidity_8"], ShouldEqual, 49.5)
	a.So(fields["accelerometer_9"], ShouldResemble, map[string]float32{
		"x": -0.424,
		"y": 0.015,
		"z": 1.666,
	})
	a.So(fields["barometric_pressure_10"], ShouldEqual, 1073.5)
	a.So(fields["gyrometer_11"], ShouldResemble, map[string]float32{
		"x": 3.55,
		"y": 5.61,
		"z": -4.10,
	})
	a.So(fields["gps_12"], ShouldResemble, map[string]float32{
		"latitude":  52.3655,
		"longitude": 4.8885,
		"altitude":  21.54,
	})
}

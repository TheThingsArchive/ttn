// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package cayennelpp

import (
	"testing"

	. "github.com/smartystreets/assertions"
)

func TestEncode(t *testing.T) {
	a := New(t)

	encoder := new(Encoder)

	// Happy flow
	{
		fields := make(map[string]interface{})
		fields["digital_in_1"] = 255
		fields["digital_out_2"] = 100
		fields["analog_in_3"] = 54.5
		fields["analog_out_4"] = -54.5
		fields["luminosity_5"] = 500
		fields["presence_6"] = 50
		fields["temperature_7"] = -15.6
		fields["relative_humidity_8"] = 49.5
		fields["accelerometer_9"] = map[string]float64{
			"x": -0.424,
			"y": 0.015,
			"z": 1.666,
		}
		fields["barometric_pressure_10"] = 1073.5
		fields["gyrometer_11"] = map[string]float64{
			"x": 3.55,
			"y": 5.61,
			"z": -4.10,
		}
		fields["gps_12"] = map[string]float64{
			"latitude":  52.3655,
			"longitude": 4.8885,
			"altitude":  21.54,
		}

		payload, err := encoder.Encode(fields, 1)
		a.So(err, ShouldBeNil)

		// Cannot test bytes here as the order is random, so testing roundtrip

		decoder := new(Decoder)
		fields, valid, err := decoder.Decode(payload, 1)
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

	// Test resilience against custom fields from the user. Should be fine
	{
		fields := map[string]interface{}{
			"custom":       8,
			"digital_in_8": "shouldn't be a string",
			"custom_5":     5,
		}
		payload, err := encoder.Encode(fields, 1)
		a.So(err, ShouldBeNil)
		a.So(payload, ShouldHaveLength, 0)
	}
}

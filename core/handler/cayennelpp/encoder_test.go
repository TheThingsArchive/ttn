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
		fields["value_2"] = float64(-50.51)

		payload, valid, err := encoder.Encode(fields, 1)
		a.So(err, ShouldBeNil)
		a.So(valid, ShouldBeTrue)
		a.So(payload, ShouldResemble, []byte{2, 236, 69})
	}

	// Test resilience against custom fields from the user. Should be fine
	{
		fields := map[string]interface{}{
			"custom":       8,
			"digital_in_8": "shouldn't be a string",
			"custom_5":     5,
			"accelerometer_1": map[string]interface{}{
				"x": "test",
			},
		}
		payload, valid, err := encoder.Encode(fields, 1)
		a.So(err, ShouldBeNil)
		a.So(valid, ShouldBeTrue)
		a.So(payload, ShouldBeEmpty)
	}
}

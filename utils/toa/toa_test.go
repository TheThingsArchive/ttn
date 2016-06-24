// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package toa

import (
	"testing"
	"time"

	. "github.com/smartystreets/assertions"
)

func TestCompute(t *testing.T) {
	a := New(t)

	var toa time.Duration
	var err error

	// Test different SFs
	sfTests := map[string]uint{
		"SF7BW125":  41216,
		"SF8BW125":  72192,
		"SF9BW125":  144384,
		"SF10BW125": 288768,
		"SF11BW125": 577536,
		"SF12BW125": 991232,
	}
	for dr, us := range sfTests {
		toa, err = Compute(10, dr, "4/5")
		a.So(err, ShouldBeNil)
		a.So(toa, ShouldAlmostEqual, time.Duration(us)*time.Microsecond)
	}

	// Test different BWs
	bwTests := map[string]uint{
		"SF7BW125": 41216,
		"SF7BW250": 20608,
		"SF7BW500": 10304,
	}
	for dr, us := range bwTests {
		toa, err = Compute(10, dr, "4/5")
		a.So(err, ShouldBeNil)
		a.So(toa, ShouldAlmostEqual, time.Duration(us)*time.Microsecond)
	}

	// Test different CRs
	crTests := map[string]uint{
		"4/5": 41216,
		"4/6": 45312,
		"4/7": 49408,
		"4/8": 53504,
	}
	for cr, us := range crTests {
		toa, err = Compute(10, "SF7BW125", cr)
		a.So(err, ShouldBeNil)
		a.So(toa, ShouldAlmostEqual, time.Duration(us)*time.Microsecond)
	}

	// Test different payload sizes
	plTests := map[uint]uint{
		13: 46336,
		14: 46336,
		15: 46336,
		16: 51456,
		17: 51456,
		18: 51456,
		19: 51456,
	}
	for size, us := range plTests {
		toa, err = Compute(size, "SF7BW125", "4/5")
		a.So(err, ShouldBeNil)
		a.So(toa, ShouldAlmostEqual, time.Duration(us)*time.Microsecond)
	}

}

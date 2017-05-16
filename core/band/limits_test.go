// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package band

import (
	"testing"
	"time"

	"github.com/TheThingsNetwork/go-utils/rate"
	. "github.com/smartystreets/assertions"
)

func TestUtilizationLimits(t *testing.T) {
	a := New(t)

	now := time.Now()

	limits := UtilizationLimits{
		10 * time.Minute: 0.02, // 12s in 10min
		time.Hour:        0.01, // 36s in 60min
	}

	counter := rate.NewCounter(time.Minute, time.Hour)
	counter.Add(now.Add(-15*time.Minute), uint64(10*time.Second))
	counter.Add(now.Add(-15*time.Second), uint64(10*time.Second))

	a.So(limits.Progress(counter), ShouldAlmostEqual, 10.0/12.0, 1e-6)
}

func TestDutyCycle(t *testing.T) {
	a := New(t)

	limits := BandLimits{
		SubBandLimits{
			appliesTo: func(_ uint64) bool {
				return true
			},
			limits: UtilizationLimits{
				5 * time.Minute: 0.05, // 15s in 5m
			},
		},
		SubBandLimits{
			appliesTo: func(f uint64) bool {
				switch f {
				case 868100000, 868300000:
					return true
				}
				return false
			},
			limits: UtilizationLimits{
				time.Minute: 0.01, // 0.6s in 1m
			},
		},
	}

	d := limits.New()
	a.So(d[0].utilization, ShouldNotBeNil)
	a.So(d[1].utilization, ShouldNotBeNil)

	a.So(d.Progress(868800000), ShouldEqual, 0)
	a.So(d.Progress(868100000), ShouldEqual, 0)
	a.So(d.Progress(868300000), ShouldEqual, 0)

	d.Add(868800000, 2*time.Second)
	a.So(d.Progress(868800000), ShouldAlmostEqual, 2.0/15)

	a.So(d.Progress(868100000), ShouldNotEqual, 0)

	d.Add(868100000, 2*time.Second)
	a.So(d.Progress(868100000), ShouldAlmostEqual, 2.0/0.6)
	a.So(d.Progress(868300000), ShouldAlmostEqual, 2.0/0.6)
}

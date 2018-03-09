// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package band

import (
	"testing"

	. "github.com/smartystreets/assertions"
)

func TestGuess(t *testing.T) {
	a := New(t)

	a.So(Guess(868100000), ShouldEqual, "EU_863_870")
	a.So(Guess(903900000), ShouldEqual, "US_902_928")
	a.So(Guess(779500000), ShouldEqual, "CN_779_787")
	a.So(Guess(433175000), ShouldEqual, "EU_433")
	a.So(Guess(916800000), ShouldEqual, "AU_915_928")
	a.So(Guess(470300000), ShouldEqual, "CN_470_510")
	a.So(Guess(923200000), ShouldEqual, "AS_923")
	a.So(Guess(922200000), ShouldEqual, "AS_920_923")
	a.So(Guess(923600000), ShouldEqual, "AS_923_925")
	a.So(Guess(922100000), ShouldEqual, "KR_920_923")
	a.So(Guess(865062500), ShouldEqual, "IN_865_867")

	a.So(Guess(922100001), ShouldEqual, "") // Not allowed
}

func TestGet(t *testing.T) {
	a := New(t)

	{
		fp, err := Get("EU_863_870")
		a.So(err, ShouldBeNil)
		a.So(fp.CFList, ShouldNotBeNil)
		a.So(fp.ADR, ShouldNotBeNil)
	}

	{
		fp, err := Get("US_902_928")
		a.So(err, ShouldBeNil)
		a.So(fp.CFList, ShouldBeNil)
		a.So(fp.ADR, ShouldNotBeNil)
	}

	{
		fp, err := Get("CN_779_787")
		a.So(err, ShouldBeNil)
		a.So(fp.CFList, ShouldBeNil)
		a.So(fp.ADR, ShouldBeNil)
	}

	{
		fp, err := Get("EU_433")
		a.So(err, ShouldBeNil)
		a.So(fp.CFList, ShouldBeNil)
		a.So(fp.ADR, ShouldBeNil)
	}

	{
		fp, err := Get("AU_915_928")
		a.So(err, ShouldBeNil)
		a.So(fp.CFList, ShouldBeNil)
		a.So(fp.ADR, ShouldNotBeNil)
	}

	{
		fp, err := Get("CN_470_510")
		a.So(err, ShouldBeNil)
		a.So(fp.CFList, ShouldBeNil)
		a.So(fp.ADR, ShouldBeNil)
	}

	{
		fp, err := Get("AS_923")
		a.So(err, ShouldBeNil)
		a.So(fp.CFList, ShouldBeNil)
		a.So(fp.ADR, ShouldNotBeNil)
	}

	{
		fp, err := Get("AS_920_923")
		a.So(err, ShouldBeNil)
		a.So(fp.CFList, ShouldNotBeNil)
		a.So(fp.ADR, ShouldNotBeNil)
	}

	{
		fp, err := Get("AS_923_925")
		a.So(err, ShouldBeNil)
		a.So(fp.CFList, ShouldNotBeNil)
		a.So(fp.ADR, ShouldNotBeNil)
	}

	{
		fp, err := Get("KR_920_923")
		a.So(err, ShouldBeNil)
		a.So(fp.CFList, ShouldNotBeNil)
		a.So(fp.ADR, ShouldBeNil)
	}

	{
		fp, err := Get("IN_865_867")
		a.So(err, ShouldBeNil)
		a.So(fp.CFList, ShouldBeNil)
		a.So(fp.ADR, ShouldBeNil)
	}

}

func TestGetDataRate(t *testing.T) {
	a := New(t)

	eu, _ := Get("EU_863_870")
	euRates := []string{"SF12BW125", "SF11BW125", "SF10BW125", "SF9BW125", "SF8BW125", "SF7BW125", "SF7BW250"}
	for expIdx, expRate := range euRates {
		idx, err := eu.GetDataRateIndexFor(expRate)
		a.So(err, ShouldBeNil)
		a.So(idx, ShouldEqual, expIdx)

		rate, err := eu.GetDataRateStringForIndex(expIdx)
		a.So(err, ShouldBeNil)
		a.So(rate, ShouldEqual, expRate)
	}
}

func TestGetTxPower(t *testing.T) {
	a := New(t)

	eu, _ := Get("EU_863_870")
	euPowers := []int{20, 14, 11, 8, 5, 2}
	for expIdx, expPower := range euPowers {
		idx, err := eu.GetTxPowerIndexFor(expPower)
		a.So(err, ShouldBeNil)
		a.So(idx, ShouldEqual, expIdx)
	}
}

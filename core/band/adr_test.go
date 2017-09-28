// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package band

import (
	"testing"

	. "github.com/smartystreets/assertions"
)

const defaultMargin = 10

func TestADRSettings(t *testing.T) {
	a := New(t)

	eu, _ := Get("EU_863_870")
	{
		dr, tx, err := eu.ADRSettings("SF12BW125", 14, -3, defaultMargin)
		a.So(err, ShouldBeNil)
		a.So(dr, ShouldEqual, "SF10BW125")
		a.So(tx, ShouldEqual, 14)
		_, err = eu.GetTxPowerIndexFor(tx)
		a.So(err, ShouldBeNil)
	}
	{
		dr, tx, err := eu.ADRSettings("SF7BW125", 14, 9, defaultMargin)
		a.So(err, ShouldBeNil)
		a.So(dr, ShouldEqual, "SF7BW125")
		a.So(tx, ShouldEqual, 8)
		_, err = eu.GetTxPowerIndexFor(tx)
		a.So(err, ShouldBeNil)
	}
	{
		dr, tx, err := eu.ADRSettings("SF7BW125", 4, 9, defaultMargin)
		a.So(err, ShouldBeNil)
		a.So(dr, ShouldEqual, "SF7BW125")
		a.So(tx, ShouldEqual, 2)
		_, err = eu.GetTxPowerIndexFor(tx)
		a.So(err, ShouldBeNil)
	}
	{
		dr, tx, err := eu.ADRSettings("SF7BW125", 9, -6, defaultMargin)
		a.So(err, ShouldBeNil)
		a.So(dr, ShouldEqual, "SF7BW125")
		a.So(tx, ShouldEqual, 14)
		_, err = eu.GetTxPowerIndexFor(tx)
		a.So(err, ShouldBeNil)
	}
	{
		dr, tx, err := eu.ADRSettings("SF7BW125", 8, -6, defaultMargin)
		a.So(err, ShouldBeNil)
		a.So(dr, ShouldEqual, "SF7BW125")
		a.So(tx, ShouldEqual, 14)
		_, err = eu.GetTxPowerIndexFor(tx)
		a.So(err, ShouldBeNil)
	}

	us, _ := Get("US_902_928")
	{
		dr, tx, err := us.ADRSettings("SF7BW125", 10, -6, defaultMargin)
		a.So(err, ShouldBeNil)
		a.So(dr, ShouldEqual, "SF7BW125")
		a.So(tx, ShouldEqual, 16)
		_, err = us.GetTxPowerIndexFor(tx)
		a.So(err, ShouldBeNil)
	}
	{
		dr, tx, err := us.ADRSettings("SF7BW125", 16, -6, defaultMargin)
		a.So(err, ShouldBeNil)
		a.So(dr, ShouldEqual, "SF7BW125")
		a.So(tx, ShouldEqual, 20)
		_, err = us.GetTxPowerIndexFor(tx)
		a.So(err, ShouldBeNil)
	}
	{
		dr, tx, err := us.ADRSettings("SF7BW125", 20, -6, defaultMargin)
		a.So(err, ShouldBeNil)
		a.So(dr, ShouldEqual, "SF7BW125")
		a.So(tx, ShouldEqual, 20)
		_, err = us.GetTxPowerIndexFor(tx)
		a.So(err, ShouldBeNil)
	}
	{
		dr, tx, err := us.ADRSettings("SF7BW125", 14, 6, defaultMargin)
		a.So(err, ShouldBeNil)
		a.So(dr, ShouldEqual, "SF7BW125")
		a.So(tx, ShouldEqual, 12)
		_, err = us.GetTxPowerIndexFor(tx)
		a.So(err, ShouldBeNil)
	}
	{
		dr, tx, err := us.ADRSettings("SF7BW125", 12, 9, defaultMargin)
		a.So(err, ShouldBeNil)
		a.So(dr, ShouldEqual, "SF7BW125")
		a.So(tx, ShouldEqual, 10)
		_, err = us.GetTxPowerIndexFor(tx)
		a.So(err, ShouldBeNil)
	}
	cn, _ := Get("CN_470_510")
	{
		_, _, err := cn.ADRSettings("SF10BW125", 14, -3, defaultMargin)
		a.So(err, ShouldNotBeNil)
	}

	// and some error cases
	{
		// Invalid datarate (there is no SF19)
		_, _, err := eu.ADRSettings("SF19BW125", 14, -3, defaultMargin)
		a.So(err, ShouldNotBeNil)
		_, _, err = us.ADRSettings("SF12BW125", 14, -3, defaultMargin)
		a.So(err, ShouldNotBeNil)
	}

}

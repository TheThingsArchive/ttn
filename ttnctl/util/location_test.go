// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package util

import (
	"testing"

	"github.com/TheThingsNetwork/go-account-lib/account"
	. "github.com/smartystreets/assertions"
)

func TestParseLocation(t *testing.T) {
	a := New(t)

	str := "10.5,33.4"
	loc := &account.Location{
		Latitude:  float64(10.5),
		Longitude: float64(33.4),
	}
	parsed, err := ParseLocation(str)
	a.So(err, ShouldBeNil)
	a.So(loc.Latitude, ShouldEqual, parsed.Latitude)
	a.So(loc.Longitude, ShouldEqual, parsed.Longitude)
}

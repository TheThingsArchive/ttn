// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package util

import (
	"testing"

	"github.com/TheThingsNetwork/ttn/core/account"
	s "github.com/smartystreets/assertions"
)

func TestParseLocation(t *testing.T) {
	a := s.New(t)

	str := "10.5,33.4"
	loc := &account.Location{
		Latitude:  float64(10.5),
		Longitude: float64(33.4),
	}
	parsed, err := ParseLocation(str)
	a.So(err, s.ShouldBeNil)
	a.So(loc.Latitude, s.ShouldEqual, parsed.Latitude)
	a.So(loc.Longitude, s.ShouldEqual, parsed.Longitude)
}

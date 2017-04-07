// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package cayennelpp

import (
	"testing"

	. "github.com/smartystreets/assertions"
)

func TestParseName(t *testing.T) {
	a := New(t)

	{
		key, channel, err := parseName("digital_in_8")
		a.So(err, ShouldBeNil)
		a.So(key, ShouldEqual, digitalInputKey)
		a.So(channel, ShouldEqual, 8)
	}

	{
		_, _, err := parseName("digital_in_-1")
		a.So(err, ShouldNotBeNil)
	}

	{
		_, _, err := parseName("_5")
		a.So(err, ShouldNotBeNil)
	}

	{
		_, _, err := parseName("test")
		a.So(err, ShouldNotBeNil)
	}

	{
		_, _, err := parseName("test_wrong")
		a.So(err, ShouldNotBeNil)
	}
}

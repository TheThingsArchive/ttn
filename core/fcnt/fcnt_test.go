// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package fcnt

import (
	"testing"

	. "github.com/smartystreets/assertions"
)

func TestGetFullFCnt(t *testing.T) {
	a := New(t)

	a.So(GetFull(0, 0), ShouldEqual, 0)
	a.So(GetFull(0, 1), ShouldEqual, 1)
	a.So(GetFull(0, 65535), ShouldEqual, 65535)

	a.So(GetFull(2000, 0), ShouldEqual, 65536)
	a.So(GetFull(2000, 1), ShouldEqual, 65537)

	a.So(GetFull(65536, 0), ShouldEqual, 65536)
	a.So(GetFull(65536, 1), ShouldEqual, 65537)

	a.So(GetFull(524287, 0), ShouldEqual, 524288)
	a.So(GetFull(524287, 1), ShouldEqual, 524289)

	a.So(GetFull(524288, 0), ShouldEqual, 524288)
	a.So(GetFull(524288, 1), ShouldEqual, 524289)
}

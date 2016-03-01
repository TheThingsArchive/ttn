// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package http

import (
	"testing"

	. "github.com/TheThingsNetwork/ttn/utils/testing"
	"github.com/smartystreets/assertions"
)

func TestBadRequest(t *testing.T) {
	rw := NewResponseWriter()
	BadRequest(&rw, "Test")
	a := assertions.New(t)
	a.So(rw.TheStatus, assertions.ShouldEqual, 400)
	a.So(string(rw.TheBody), assertions.ShouldEqual, "Test")
}

// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package http

import (
	"net/http"
	"testing"

	"github.com/TheThingsNetwork/ttn/utils/stats"
	. "github.com/TheThingsNetwork/ttn/utils/testing"
	"github.com/smartystreets/assertions"
)

func TestStatusPageURL(t *testing.T) {
	a := assertions.New(t)

	s := StatusPage{}

	a.So(s.URL(), assertions.ShouldEqual, "/status")
}

func TestStatusPageHandler(t *testing.T) {
	a := assertions.New(t)

	s := StatusPage{}

	// Only GET allowed
	r1, _ := http.NewRequest("POST", "/status", nil)
	r1.RemoteAddr = "127.0.0.1:12345"
	rw1 := NewResponseWriter()
	s.Handle(&rw1, r1)
	a.So(rw1.TheStatus, assertions.ShouldEqual, 405)

	// Initially Empty
	r3, _ := http.NewRequest("GET", "/status", nil)
	r3.RemoteAddr = "127.0.0.1:12345"
	rw3 := NewResponseWriter()
	s.Handle(&rw3, r3)
	a.So(rw3.TheStatus, assertions.ShouldEqual, 200)
	a.So(string(rw3.TheBody), assertions.ShouldEqual, "{}")

	// Create some stats
	stats.IncCounter("this.is-a-counter")
	stats.UpdateHistogram("and.this.is.a-histogram", 123)
	stats.MarkMeter("and.this.is.a-meter")
	stats.SetString("and.this.is.a-string", "with_key", "with_value")

	// Not Empty anymore
	r4, _ := http.NewRequest("GET", "/status", nil)
	r4.RemoteAddr = "127.0.0.1:12345"
	rw4 := NewResponseWriter()
	s.Handle(&rw4, r4)
	a.So(rw4.TheStatus, assertions.ShouldEqual, 200)
	a.So(string(rw4.TheBody), assertions.ShouldContainSubstring, "\"is-a-counter\"")
	a.So(string(rw4.TheBody), assertions.ShouldContainSubstring, "p_50")
	a.So(string(rw4.TheBody), assertions.ShouldContainSubstring, "rate_15")
	a.So(string(rw4.TheBody), assertions.ShouldContainSubstring, "with_key")
	a.So(string(rw4.TheBody), assertions.ShouldContainSubstring, "with_value")
}

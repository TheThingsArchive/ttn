// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package handlers

import (
	"net/http"
	"testing"

	. "github.com/TheThingsNetwork/ttn/core/adapters/http"
	. "github.com/TheThingsNetwork/ttn/utils/testing"
	"github.com/smartystreets/assertions"
)

func TestHealthzURL(t *testing.T) {
	a := assertions.New(t)

	h := Healthz{}

	a.So(h.URL(), assertions.ShouldEqual, "/healthz")
}

func TestHealthzHandle(t *testing.T) {
	a := assertions.New(t)

	h := Healthz{}

	req, _ := http.NewRequest("GET", "/healthz", nil)
	req.RemoteAddr = "127.0.0.1:12345"
	rw := NewResponseWriter()

	h.Handle(&rw, make(chan<- PktReq), make(chan<- RegReq), req)
	a.So(rw.TheStatus, assertions.ShouldEqual, 200)
	a.So(string(rw.TheBody), assertions.ShouldEqual, "ok")
}

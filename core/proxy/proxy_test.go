// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package proxy

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/smartystreets/assertions"
)

type testHandler struct {
	req *http.Request
	res http.ResponseWriter
}

func (h *testHandler) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	h.res = res
	h.req = req
}

func TestTokenProxier(t *testing.T) {
	a := New(t)

	hdl := &testHandler{}
	p := WithToken(hdl)

	req := httptest.NewRequest("GET", "/uri", bytes.NewBuffer([]byte{}))
	p.ServeHTTP(httptest.NewRecorder(), req)
	a.So(hdl.req.Header.Get("Grpc-Metadata-Token"), ShouldBeEmpty)

	req = httptest.NewRequest("GET", "/uri", bytes.NewBuffer([]byte{}))
	req.Header.Add("Authorization", "Key blabla")
	p.ServeHTTP(httptest.NewRecorder(), req)
	a.So(hdl.req.Header.Get("Grpc-Metadata-Token"), ShouldBeEmpty)

	req = httptest.NewRequest("GET", "/uri", bytes.NewBuffer([]byte{}))
	req.Header.Add("Authorization", "bearer token")
	p.ServeHTTP(httptest.NewRecorder(), req)
	a.So(hdl.req.Header.Get("Grpc-Metadata-Token"), ShouldEqual, "token")

	req = httptest.NewRequest("GET", "/uri", bytes.NewBuffer([]byte{}))
	req.Header.Add("Authorization", "Bearer token")
	p.ServeHTTP(httptest.NewRecorder(), req)
	a.So(hdl.req.Header.Get("Grpc-Metadata-Token"), ShouldEqual, "token")
}

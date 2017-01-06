// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package proxy

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/TheThingsNetwork/ttn/utils/testing"
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

func TestLogProxier(t *testing.T) {
	a := New(t)

	hdl := &testHandler{}
	p := WithLogger(hdl, GetLogger(t, ""))

	req := httptest.NewRequest("GET", "/uri", bytes.NewBuffer([]byte{}))
	p.ServeHTTP(httptest.NewRecorder(), req)
	a.So(hdl.req, ShouldNotBeNil)
	a.So(hdl.res, ShouldNotBeNil)
}

func TestPaginatedProxier(t *testing.T) {
	a := New(t)

	hdl := &testHandler{}
	p := WithPagination(hdl)
	req := httptest.NewRequest("GET", "/uri", nil)
	w := httptest.NewRecorder()
	p.ServeHTTP(w, req)
	a.So(hdl.req.Header.Get("Grpc-Metadata-Offset"), ShouldEqual, "")
	a.So(hdl.req.Header.Get("Grpc-Metadata-Limit"), ShouldEqual, "")
	a.So(w.Code, ShouldEqual, http.StatusOK)

	hdl = &testHandler{}
	p = WithPagination(hdl)
	req = httptest.NewRequest("GET", "/uri?offset=42", nil)
	w = httptest.NewRecorder()
	p.ServeHTTP(w, req)
	a.So(hdl.req.Header.Get("Grpc-Metadata-Offset"), ShouldEqual, "42")
	a.So(hdl.req.Header.Get("Grpc-Metadata-Limit"), ShouldEqual, "")
	a.So(w.Code, ShouldEqual, http.StatusOK)

	hdl = &testHandler{}
	p = WithPagination(hdl)
	req = httptest.NewRequest("GET", "/uri?limit=42", nil)
	w = httptest.NewRecorder()
	p.ServeHTTP(w, req)
	a.So(hdl.req.Header.Get("Grpc-Metadata-Offset"), ShouldEqual, "")
	a.So(hdl.req.Header.Get("Grpc-Metadata-Limit"), ShouldEqual, "42")
	a.So(w.Code, ShouldEqual, http.StatusOK)

	hdl = &testHandler{}
	p = WithPagination(hdl)
	req = httptest.NewRequest("GET", "/uri?offset=42&limit=42", nil)
	w = httptest.NewRecorder()
	p.ServeHTTP(w, req)
	a.So(hdl.req.Header.Get("Grpc-Metadata-Offset"), ShouldEqual, "42")
	a.So(hdl.req.Header.Get("Grpc-Metadata-Limit"), ShouldEqual, "42")
	a.So(w.Code, ShouldEqual, http.StatusOK)

	hdl = &testHandler{}
	p = WithPagination(hdl)
	req = httptest.NewRequest("GET", "/uri?offset=test", nil)
	w = httptest.NewRecorder()
	p.ServeHTTP(w, req)
	a.So(hdl.req, ShouldBeNil)
	a.So(w.Code, ShouldEqual, http.StatusBadRequest)

	hdl = &testHandler{}
	p = WithPagination(hdl)
	req = httptest.NewRequest("GET", "/uri?limit=test", nil)
	w = httptest.NewRecorder()
	p.ServeHTTP(w, req)
	a.So(hdl.req, ShouldBeNil)
	a.So(w.Code, ShouldEqual, http.StatusBadRequest)

	hdl = &testHandler{}
	p = WithPagination(hdl)
	req = httptest.NewRequest("GET", "/uri?offset=test&limit=test", nil)
	w = httptest.NewRecorder()
	p.ServeHTTP(w, req)
	a.So(hdl.req, ShouldBeNil)
	a.So(w.Code, ShouldEqual, http.StatusBadRequest)
}

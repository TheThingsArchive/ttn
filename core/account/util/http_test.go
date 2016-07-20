// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package util

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/smartystreets/assertions"
)

type OKResp struct {
	OK string `json:"ok"`
}

type FooResp struct {
	Foo string `json:"foo" valid:"required"`
}

func OKHandler(a *Assertion, method string) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		a.So(r.RequestURI, ShouldEqual, "/foo")
		a.So(r.Method, ShouldEqual, method)
		resp := OKResp{
			OK: "ok",
		}
		w.WriteHeader(http.StatusOK)
		encoder := json.NewEncoder(w)
		encoder.Encode(&resp)
	})
}

func FooHandler(a *Assertion, method string) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		a.So(r.RequestURI, ShouldEqual, "/foo")
		a.So(r.Method, ShouldEqual, method)
		resp := FooResp{
			Foo: "ok",
		}
		w.WriteHeader(http.StatusOK)
		encoder := json.NewEncoder(w)
		encoder.Encode(&resp)
	})
}

func RedirectHandler(a *Assertion, method string) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.RequestURI == "/foo" {
			w.Header().Set("Location", "/bar")
			w.WriteHeader(307)
		} else {
			a.So(r.RequestURI, ShouldEqual, "/bar")
			resp := FooResp{
				Foo: "ok",
			}
			w.WriteHeader(http.StatusOK)
			encoder := json.NewEncoder(w)
			encoder.Encode(&resp)
		}
	})
}

func EchoHandler(a *Assertion, method string) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		a.So(r.RequestURI, ShouldEqual, "/foo")
		a.So(r.Method, ShouldEqual, method)
		w.WriteHeader(http.StatusOK)
		defer r.Body.Close()
		body, err := ioutil.ReadAll(r.Body)
		a.So(err, ShouldBeNil)
		w.Write(body)
	})
}

func TestGET(t *testing.T) {
	a := New(t)
	server := httptest.NewServer(OKHandler(a, "GET"))
	defer server.Close()

	var resp OKResp
	err := GET(server.URL, "ok", "/foo", &resp)
	a.So(err, ShouldBeNil)
	a.So(resp.OK, ShouldEqual, "ok")
}

func TestGETDropResponse(t *testing.T) {
	a := New(t)
	server := httptest.NewServer(OKHandler(a, "GET"))
	defer server.Close()

	err := GET(server.URL, "ok", "/foo", nil)
	a.So(err, ShouldBeNil)
}

func TestGETIllegalResponse(t *testing.T) {
	a := New(t)
	server := httptest.NewServer(OKHandler(a, "GET"))
	defer server.Close()

	var resp FooResp
	err := GET(server.URL, "ok", "/foo", &resp)
	a.So(err, ShouldNotBeNil)
}

func TestGETIllegalResponseIgnore(t *testing.T) {
	a := New(t)
	server := httptest.NewServer(FooHandler(a, "GET"))
	defer server.Close()

	var resp OKResp
	err := GET(server.URL, "ok", "/foo", &resp)
	a.So(err, ShouldBeNil)
}

func TestGETRedirect(t *testing.T) {
	a := New(t)
	server := httptest.NewServer(RedirectHandler(a, "GET"))
	defer server.Close()

	var resp OKResp
	err := GET(server.URL, "ok", "/foo", &resp)
	a.So(err, ShouldBeNil)
}

func TestPUT(t *testing.T) {
	a := New(t)
	server := httptest.NewServer(EchoHandler(a, "PUT"))
	defer server.Close()

	var resp FooResp
	body := FooResp{
		Foo: "ok",
	}
	err := PUT(server.URL, "ok", "/foo", body, &resp)
	a.So(err, ShouldBeNil)
	a.So(resp.Foo, ShouldEqual, body.Foo)
}

func TestPUTIllegalRequest(t *testing.T) {
	a := New(t)
	server := httptest.NewServer(EchoHandler(a, "PUT"))
	defer server.Close()

	var resp FooResp
	body := FooResp{}
	err := PUT(server.URL, "ok", "/foo", body, &resp)
	a.So(err, ShouldNotBeNil)
}

func TestPUTIllegalResponse(t *testing.T) {
	a := New(t)
	server := httptest.NewServer(OKHandler(a, "PUT"))
	defer server.Close()

	var resp FooResp
	err := PUT(server.URL, "ok", "/foo", nil, &resp)
	a.So(err, ShouldNotBeNil)
}

func TestPUTRedirect(t *testing.T) {
	a := New(t)
	server := httptest.NewServer(RedirectHandler(a, "PUT"))
	defer server.Close()

	var resp FooResp
	err := PUT(server.URL, "ok", "/foo", nil, &resp)
	a.So(err, ShouldBeNil)
	a.So(resp.Foo, ShouldEqual, "ok")
}

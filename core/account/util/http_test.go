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

var (
	url   = "/foo"
	token = "token"
)

type OKResp struct {
	OK string `json:token`
}

type FooResp struct {
	Foo string `json:"foo" valid:"required"`
}

func OKHandler(a *Assertion, method string) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		a.So(r.RequestURI, ShouldEqual, url)
		a.So(r.Method, ShouldEqual, method)
		a.So(r.Header.Get("Authorization"), ShouldEqual, "bearer "+token)
		resp := OKResp{
			OK: token,
		}
		w.WriteHeader(http.StatusOK)
		encoder := json.NewEncoder(w)
		encoder.Encode(&resp)
	})
}

func FooHandler(a *Assertion, method string) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		a.So(r.RequestURI, ShouldEqual, url)
		a.So(r.Method, ShouldEqual, method)
		a.So(r.Header.Get("Authorization"), ShouldEqual, "bearer "+token)
		resp := FooResp{
			Foo: token,
		}
		w.WriteHeader(http.StatusOK)
		encoder := json.NewEncoder(w)
		encoder.Encode(&resp)
	})
}

func RedirectHandler(a *Assertion, method string) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		a.So(r.Header.Get("Authorization"), ShouldEqual, "bearer "+token)
		if r.RequestURI == url {
			w.Header().Set("Location", "/bar")
			w.WriteHeader(307)
		} else {
			a.So(r.RequestURI, ShouldEqual, "/bar")
			resp := FooResp{
				Foo: token,
			}
			w.WriteHeader(http.StatusOK)
			encoder := json.NewEncoder(w)
			encoder.Encode(&resp)
		}
	})
}

func EchoHandler(a *Assertion, method string) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		a.So(r.RequestURI, ShouldEqual, url)
		a.So(r.Method, ShouldEqual, method)
		a.So(r.Header.Get("Authorization"), ShouldEqual, "bearer "+token)
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
	err := GET(server.URL, token, url, &resp)
	a.So(err, ShouldBeNil)
	a.So(resp.OK, ShouldEqual, token)
}

func TestGETDropResponse(t *testing.T) {
	a := New(t)
	server := httptest.NewServer(OKHandler(a, "GET"))
	defer server.Close()

	err := GET(server.URL, token, url, nil)
	a.So(err, ShouldBeNil)
}

func TestGETIllegalResponse(t *testing.T) {
	a := New(t)
	server := httptest.NewServer(OKHandler(a, "GET"))
	defer server.Close()

	var resp FooResp
	err := GET(server.URL, token, url, &resp)
	a.So(err, ShouldNotBeNil)
}

func TestGETIllegalResponseIgnore(t *testing.T) {
	a := New(t)
	server := httptest.NewServer(FooHandler(a, "GET"))
	defer server.Close()

	var resp OKResp
	err := GET(server.URL, token, url, &resp)
	a.So(err, ShouldBeNil)
}

func TestGETRedirect(t *testing.T) {
	a := New(t)
	server := httptest.NewServer(RedirectHandler(a, "GET"))
	defer server.Close()

	var resp OKResp
	err := GET(server.URL, token, url, &resp)
	a.So(err, ShouldBeNil)
}

func TestPUT(t *testing.T) {
	a := New(t)
	server := httptest.NewServer(EchoHandler(a, "PUT"))
	defer server.Close()

	var resp FooResp
	body := FooResp{
		Foo: token,
	}
	err := PUT(server.URL, token, url, body, &resp)
	a.So(err, ShouldBeNil)
	a.So(resp.Foo, ShouldEqual, body.Foo)
}

func TestPUTIllegalRequest(t *testing.T) {
	a := New(t)
	server := httptest.NewServer(EchoHandler(a, "PUT"))
	defer server.Close()

	var resp FooResp
	body := FooResp{}
	err := PUT(server.URL, token, url, body, &resp)
	a.So(err, ShouldNotBeNil)
}

func TestPUTIllegalResponse(t *testing.T) {
	a := New(t)
	server := httptest.NewServer(OKHandler(a, "PUT"))
	defer server.Close()

	var resp FooResp
	err := PUT(server.URL, token, url, nil, &resp)
	a.So(err, ShouldNotBeNil)
}

func TestPUTRedirect(t *testing.T) {
	a := New(t)
	server := httptest.NewServer(RedirectHandler(a, "PUT"))
	defer server.Close()

	var resp FooResp
	err := PUT(server.URL, token, url, nil, &resp)
	a.So(err, ShouldBeNil)
	a.So(resp.Foo, ShouldEqual, token)
}

func TestPOST(t *testing.T) {
	a := New(t)
	server := httptest.NewServer(EchoHandler(a, "POST"))
	defer server.Close()

	var resp FooResp
	body := FooResp{
		Foo: token,
	}
	err := POST(server.URL, token, url, body, &resp)
	a.So(err, ShouldBeNil)
	a.So(resp.Foo, ShouldEqual, body.Foo)
}

func TestPOSTIllegalRequest(t *testing.T) {
	a := New(t)
	server := httptest.NewServer(EchoHandler(a, "POST"))
	defer server.Close()

	var resp FooResp
	body := FooResp{}
	err := POST(server.URL, token, url, body, &resp)
	a.So(err, ShouldNotBeNil)
}

func TestPOSTIllegalResponse(t *testing.T) {
	a := New(t)
	server := httptest.NewServer(OKHandler(a, "POST"))
	defer server.Close()

	var resp FooResp
	err := POST(server.URL, token, url, nil, &resp)
	a.So(err, ShouldNotBeNil)
}

func TestPOSTRedirect(t *testing.T) {
	a := New(t)
	server := httptest.NewServer(RedirectHandler(a, "POST"))
	defer server.Close()

	var resp FooResp
	err := POST(server.URL, token, url, nil, &resp)
	a.So(err, ShouldBeNil)
	a.So(resp.Foo, ShouldEqual, token)
}

func TestDELETE(t *testing.T) {
	a := New(t)
	server := httptest.NewServer(OKHandler(a, "DELETE"))
	defer server.Close()

	err := DELETE(server.URL, token, url)
	a.So(err, ShouldBeNil)
}

func TestDELETERedirect(t *testing.T) {
	a := New(t)
	server := httptest.NewServer(RedirectHandler(a, "DELETE"))
	defer server.Close()

	err := DELETE(server.URL, token, url)
	a.So(err, ShouldBeNil)
}

package util

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/smartystreets/assertions"
)

type OKResp struct {
	OK string `json:"ok"`
}

type FooResp struct {
	Foo string `json:"foo" validate:"nonzero"`
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

func TestGETWrongResponseType(t *testing.T) {
	a := New(t)
	server := httptest.NewServer(OKHandler(a, "GET"))
	defer server.Close()

	var resp FooResp
	err := GET(server.URL, "ok", "/foo", &resp)
	a.So(err, ShouldNotBeNil)
}

func TestGETWrongResponseTypeIgnore(t *testing.T) {
	a := New(t)
	server := httptest.NewServer(FooHandler(a, "GET"))
	defer server.Close()

	var resp OKResp
	err := GET(server.URL, "ok", "/foo", &resp)
	a.So(err, ShouldBeNil)
}

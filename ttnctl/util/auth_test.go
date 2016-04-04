package util

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	. "github.com/smartystreets/assertions"
)

func newTokenServer(a *Assertion) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		a.So(r.RequestURI, ShouldEqual, "/users/token")
		a.So(r.Method, ShouldEqual, "POST")

		username, password, ok := r.BasicAuth()
		a.So(ok, ShouldBeTrue)
		a.So(username, ShouldEqual, "ttnctl")
		a.So(password, ShouldEqual, "")

		grantType := r.FormValue("grant_type")
		if grantType == "password" {
			handleNewToken(a, w, r)
		} else if grantType == "refresh_token" {
			handleRefreshToken(a, w, r)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
}

func handleNewToken(a *Assertion, w http.ResponseWriter, r *http.Request) {
	var resp token
	if r.FormValue("username") == "jantje@test.org" && r.FormValue("password") == "secret" {
		resp = token{
			AccessToken:  "123",
			RefreshToken: "ABC",
			ExpiresIn:    3600,
		}
		w.WriteHeader(http.StatusOK)
	} else {
		resp = token{
			Error:            "invalid_credentials",
			ErrorDescription: "Invalid credentials",
		}
		w.WriteHeader(http.StatusForbidden)
	}

	encoder := json.NewEncoder(w)
	err := encoder.Encode(&resp)
	a.So(err, ShouldBeNil)
}

func handleRefreshToken(a *Assertion, w http.ResponseWriter, r *http.Request) {
	var resp token
	if r.FormValue("refresh_token") == "ABC" {
		resp = token{
			AccessToken:  "456",
			RefreshToken: "DEF",
			ExpiresIn:    3600,
		}
		w.WriteHeader(http.StatusOK)
	} else {
		resp = token{
			Error:            "invalid_grant",
			ErrorDescription: "Refresh token not found",
		}
		w.WriteHeader(http.StatusBadRequest)
	}

	encoder := json.NewEncoder(w)
	err := encoder.Encode(&resp)
	a.So(err, ShouldBeNil)
}

func TestLogin(t *testing.T) {
	a := New(t)
	server := newTokenServer(a)
	defer server.Close()

	_, err := Login(server.URL, "pietje@test.org", "secret")
	a.So(err, ShouldNotBeNil)
	a.So(err.Error(), ShouldEqual, "Invalid credentials")

	loginAuth, err := Login(server.URL, "jantje@test.org", "secret")
	a.So(err, ShouldBeNil)
	a.So(loginAuth, ShouldNotBeNil)
	a.So(loginAuth.AccessToken, ShouldEqual, "123")
	a.So(loginAuth.RefreshToken, ShouldEqual, "ABC")
	a.So(loginAuth.Email, ShouldEqual, "jantje@test.org")

	loadedAuth, err := LoadAuth(server.URL)
	a.So(err, ShouldBeNil)
	a.So(loadedAuth, ShouldNotBeNil)
	a.So(loginAuth, ShouldResemble, loadedAuth)

	// Check if we get this token on the HTTP request
	req, err := NewRequestWithAuth(server.URL, "GET", "http://external", nil)
	a.So(err, ShouldBeNil)
	a.So(req, ShouldNotBeNil)
	a.So(req.Header.Get("Authorization"), ShouldEqual, fmt.Sprintf("bearer %s", loadedAuth.AccessToken))

	Logout(server.URL)
}

func TestLogout(t *testing.T) {
	a := New(t)
	server := newTokenServer(a)
	defer server.Close()

	// Make sure we're not logged on
	err := Logout(server.URL)
	a.So(err, ShouldBeNil)
	loadedAuth, err := LoadAuth(server.URL)
	a.So(err, ShouldBeNil)
	a.So(loadedAuth, ShouldBeNil)

	// Login
	loginAuth, err := Login(server.URL, "jantje@test.org", "secret")
	a.So(err, ShouldBeNil)
	a.So(loginAuth, ShouldNotBeNil)

	// Logout
	err = Logout(server.URL)
	a.So(err, ShouldBeNil)
	loadedAuth, err = LoadAuth(server.URL)
	a.So(err, ShouldBeNil)
	a.So(loadedAuth, ShouldBeNil)

	// Make sure that we can't make an HTTP request
	_, err = NewRequestWithAuth(server.URL, "GET", "http://external", nil)
	a.So(err, ShouldNotBeNil)
}

func TestLoadWithRefresh(t *testing.T) {
	a := New(t)
	server := newTokenServer(a)
	defer server.Close()

	// Make sure we're not logged on
	err := Logout(server.URL)
	a.So(err, ShouldBeNil)

	// Save an expired token
	expires := time.Now().Add(time.Duration(-1) * time.Hour)
	savedAuth, err := saveAuth(server.URL, "jantje@test.org", "123", "ABC", expires)
	a.So(err, ShouldBeNil)

	// Refresh the token
	loadedAuth, err := LoadAuth(server.URL)
	a.So(err, ShouldBeNil)
	a.So(loadedAuth, ShouldNotBeNil)
	a.So(savedAuth, ShouldNotResemble, loadedAuth)
	a.So(loadedAuth.AccessToken, ShouldEqual, "456")
	a.So(loadedAuth.RefreshToken, ShouldEqual, "DEF")
	a.So(loadedAuth.Email, ShouldEqual, "jantje@test.org")

	Logout(server.URL)
}

func TestLoadWithInvalidRefresh(t *testing.T) {
	a := New(t)
	server := newTokenServer(a)
	defer server.Close()

	// Make sure we're not logged on
	err := Logout(server.URL)
	a.So(err, ShouldBeNil)

	// Save an expired token
	expires := time.Now().Add(time.Duration(-1) * time.Hour)
	_, err = saveAuth(server.URL, "pietje@test.org", "987", "ZYX", expires)
	a.So(err, ShouldBeNil)

	// Refresh the token
	loadedAuth, err := LoadAuth(server.URL)
	a.So(err, ShouldNotBeNil)
	a.So(err.Error(), ShouldEqual, "Refresh token not found")
	a.So(loadedAuth, ShouldBeNil)

	Logout(server.URL)
}

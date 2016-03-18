// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package http

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	. "github.com/TheThingsNetwork/ttn/utils/testing"
)

// MockHandler mocks the http.Handler interface
type MockHandler struct {
	Failures map[string]error
	InURL    struct {
		Called bool
	}
	OutURL struct {
		URL string
	}
	InHandle struct {
		Called bool
	}
}

// NewMockHandler constructs a new MockHandler object
func NewMockHandler() *MockHandler {
	return &MockHandler{
		Failures: make(map[string]error),
	}
}

// URL implements the http.Handler interface
func (m *MockHandler) URL() string {
	m.InURL.Called = true
	return m.OutURL.URL
}

// Handle implements the http.Handler interface
func (m *MockHandler) Handle(w http.ResponseWriter, req *http.Request) error {
	m.InHandle.Called = true
	return m.Failures["Handle"]
}

func TestBind(t *testing.T) {
	{
		Desc(t, "Bind a handler and handle a request")

		// Build
		options := Options{NetAddr: "0.0.0.0:3001", Timeout: time.Millisecond * 50}
		a := New(
			Components{Ctx: GetLogger(t, "Adapter")},
			options,
		)
		cli := http.Client{}
		hdl := NewMockHandler()
		hdl.OutURL.URL = "/mock"
		<-time.After(time.Millisecond * 50)

		// Operate
		a.Bind(hdl)
		_, _ = cli.Get(fmt.Sprintf("http://%s%s", options.NetAddr, hdl.URL()))

		// Check
		Check(t, true, hdl.InURL.Called, "Url() Calls")
		Check(t, true, hdl.InHandle.Called, "Handle() Calls")
	}

	// --------------------

	{
		Desc(t, "Bind a handler and handle a request | handle fails")

		// Build
		options := Options{NetAddr: "0.0.0.0:3002", Timeout: time.Millisecond * 50}
		a := New(
			Components{Ctx: GetLogger(t, "Adapter")},
			options,
		)
		cli := http.Client{}
		hdl := NewMockHandler()
		hdl.OutURL.URL = "/mock"
		hdl.Failures["Handle"] = fmt.Errorf("Mock Error")
		<-time.After(time.Millisecond * 50)

		// Operate
		a.Bind(hdl)
		_, _ = cli.Get(fmt.Sprintf("http://%s%s", options.NetAddr, hdl.URL()))

		// Check
		Check(t, true, hdl.InURL.Called, "Url() Calls")
		Check(t, true, hdl.InHandle.Called, "Handle() Calls")
	}
}

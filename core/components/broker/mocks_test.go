// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package broker

import (
	"github.com/TheThingsNetwork/ttn/core"
)

// NOTE All the code below could be generated

// MockDialer mocks the Dialer interface
type MockDialer struct {
	Failures        map[string]error
	InMarshalSafely struct {
		Called bool
	}
	OutMarshalSafely struct {
		Data []byte
	}
	InDial struct {
		Called bool
	}
	OutDial struct {
		Client core.HandlerClient
		Closer Closer
	}
}

// NewMockDialer instantiates a new MockDialer object
func NewMockDialer() *MockDialer {
	return &MockDialer{
		Failures: make(map[string]error),
	}
}

// MarshalSafely implements the Dialer interface
func (m *MockDialer) MarshalSafely() []byte {
	m.InMarshalSafely.Called = true
	return m.OutMarshalSafely.Data
}

// Dial implements the Dialer interface
func (m *MockDialer) Dial() (core.HandlerClient, Closer, error) {
	m.InDial.Called = true
	return m.OutDial.Client, m.OutDial.Closer, m.Failures["Dial"]
}

// MockCloser mocks the Closer interface
type MockCloser struct {
	Failures map[string]error
	InClose  struct {
		Called bool
	}
}

// NewMockCloser instantiates a new MockCloser object
func NewMockCloser() *MockCloser {
	return &MockCloser{
		Failures: make(map[string]error),
	}
}

// Close implements the Closer interface
func (m *MockCloser) Close() error {
	m.InClose.Called = true
	return m.Failures["Close"]
}

// MockNetworkController mocks the NetworkController interface
type MockNetworkController struct {
	Failures        map[string]error
	InLookupDevices struct {
		DevAddr []byte
	}
	OutLookupDevices struct {
		Entries []devEntry
	}
	InWholeCounter struct {
		DevCnt   uint32
		EntryCnt uint32
	}
	OutWholeCounter struct {
		FCnt uint32
	}
	InStoreDevice struct {
		DevAddr []byte
		Entry   devEntry
	}
	InUpdateFcnt struct {
		AppEUI  []byte
		DevEUI  []byte
		DevAddr []byte
		FCnt    uint32
	}
	InClose struct {
		Called bool
	}
}

// NewMockNetworkController constructs a new MockNetworkController object
func NewMockNetworkController() *MockNetworkController {
	return &MockNetworkController{
		Failures: make(map[string]error),
	}
}

// LookupDevices implements the NetworkController interface
func (m *MockNetworkController) LookupDevices(devAddr []byte) ([]devEntry, error) {
	m.InLookupDevices.DevAddr = devAddr
	return m.OutLookupDevices.Entries, m.Failures["LookupDevices"]
}

// WholeCounter implements the NetworkController interface
func (m *MockNetworkController) WholeCounter(devCnt uint32, entryCnt uint32) (uint32, error) {
	m.InWholeCounter.DevCnt = devCnt
	m.InWholeCounter.EntryCnt = entryCnt
	return m.OutWholeCounter.FCnt, m.Failures["WholeCounter"]
}

// StoreDevice implements the NetworkController interface
func (m *MockNetworkController) StoreDevice(devAddr []byte, entry devEntry) error {
	m.InStoreDevice.DevAddr = devAddr
	m.InStoreDevice.Entry = entry
	return m.Failures["StoreDevice"]
}

// UpdateFcnt implements the NetworkController interface
func (m *MockNetworkController) UpdateFCnt(appEUI []byte, devEUI []byte, devAddr []byte, fcnt uint32) error {
	m.InUpdateFcnt.AppEUI = appEUI
	m.InUpdateFcnt.DevEUI = devEUI
	m.InUpdateFcnt.DevAddr = devAddr
	m.InUpdateFcnt.FCnt = fcnt
	return m.Failures["UpdateFCnt"]
}

// Close implements the NetworkController interface
func (m *MockNetworkController) Close() error {
	m.InClose.Called = true
	return m.Failures["Close"]
}

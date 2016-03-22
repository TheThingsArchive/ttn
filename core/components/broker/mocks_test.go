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

// MockAppStorage mocks the AppStorage interface
type MockAppStorage struct {
	Failures map[string]error
	InRead   struct {
		AppEUI []byte
	}
	OutRead struct {
		Entry appEntry
	}
	InUpsert struct {
		Entry appEntry
	}
	InDone struct {
		Called bool
	}
}

// NewMockAppStorage creates a new MockAppStorage
func NewMockAppStorage() *MockAppStorage {
	return &MockAppStorage{
		Failures: make(map[string]error),
	}
}

// read implements the AppStorage interface
func (m *MockAppStorage) read(appEUI []byte) (appEntry, error) {
	m.InRead.AppEUI = appEUI
	return m.OutRead.Entry, m.Failures["read"]
}

// upsert implements the AppStorage interface
func (m *MockAppStorage) upsert(entry appEntry) error {
	m.InUpsert.Entry = entry
	return m.Failures["upsert"]
}

// done implements the AppStorage Interface
func (m *MockAppStorage) done() error {
	m.InDone.Called = true
	return m.Failures["done"]
}

// MockNetworkController mocks the NetworkController interface
type MockNetworkController struct {
	Failures map[string]error
	InRead   struct {
		DevAddr []byte
	}
	OutRead struct {
		Entries []devEntry
	}
	InUpsert struct {
		Entry devEntry
	}
	InReadNonces struct {
		AppEUI []byte
		DevEUI []byte
	}
	OutReadNonces struct {
		Entry noncesEntry
	}
	InUpsertNonces struct {
		Entry noncesEntry
	}
	InWholeCounter struct {
		DevCnt   uint32
		EntryCnt uint32
	}
	OutWholeCounter struct {
		FCnt uint32
	}
	InDone struct {
		Called bool
	}
}

// NewMockNetworkController creates a new MockNetworkController
func NewMockNetworkController() *MockNetworkController {
	return &MockNetworkController{
		Failures: make(map[string]error),
	}
}

// read implements the NetworkController interface
func (m *MockNetworkController) read(devAddr []byte) ([]devEntry, error) {
	m.InRead.DevAddr = devAddr
	return m.OutRead.Entries, m.Failures["read"]
}

// upsert implements the NetworkController interface
func (m *MockNetworkController) upsert(entry devEntry) error {
	m.InUpsert.Entry = entry
	return m.Failures["upsert"]
}

// readNonces implements the NetworkController interface
func (m *MockNetworkController) readNonces(appEUI, devEUI []byte) (noncesEntry, error) {
	m.InReadNonces.AppEUI = appEUI
	m.InReadNonces.DevEUI = devEUI
	return m.OutReadNonces.Entry, m.Failures["readNonces"]
}

// upsertNonces implements the NetworkController interface
func (m *MockNetworkController) upsertNonces(entry noncesEntry) error {
	m.InUpsertNonces.Entry = entry
	return m.Failures["upsertNonces"]
}

// wholeCnt implements the NetworkController interface
func (m *MockNetworkController) wholeCounter(devCnt, entryCnt uint32) (uint32, error) {
	m.InWholeCounter.DevCnt = devCnt
	m.InWholeCounter.EntryCnt = entryCnt
	return m.OutWholeCounter.FCnt, m.Failures["wholeCounter"]
}

// done implements the NetworkController Interface
func (m *MockNetworkController) done() error {
	m.InDone.Called = true
	return m.Failures["done"]
}

// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package handler

// NOTE: All the code below could be generated

// MockDevStorage mocks the DevStorage interface
type MockDevStorage struct {
	Failures     map[string]error
	InUpdateFCnt struct {
		AppEUI []byte
		DevEUI []byte
		FCnt   uint32
	}
	InLookup struct {
		AppEUI []byte
		DevEUI []byte
	}
	OutLookup struct {
		Entry devEntry
	}
	InStorePersonalized struct {
		AppEUI  []byte
		DevAddr []byte
		AppSKey [16]byte
		NwkSKey [16]byte
	}
	InClose struct {
		Called bool
	}
}

// NewMockDevStorage creates a new MockDevStorage
func NewMockDevStorage() *MockDevStorage {
	return &MockDevStorage{
		Failures: make(map[string]error),
	}
}

// UpdateFCnt implements the DevStorage interface
func (m *MockDevStorage) UpdateFCnt(appEUI []byte, devEUI []byte, fcnt uint32) error {
	m.InUpdateFCnt.AppEUI = appEUI
	m.InUpdateFCnt.DevEUI = devEUI
	m.InUpdateFCnt.FCnt = fcnt
	return m.Failures["UpdateFCnt"]
}

// Lookup implements the DevStorage interface
func (m *MockDevStorage) Lookup(appEUI []byte, devEUI []byte) (devEntry, error) {
	m.InLookup.AppEUI = appEUI
	m.InLookup.DevEUI = devEUI
	return m.OutLookup.Entry, m.Failures["Lookup"]
}

// StorePersonalized implements the DevStorage interface
func (m *MockDevStorage) StorePersonalized(appEUI []byte, devAddr []byte, appSKey, nwkSKey [16]byte) error {
	m.InStorePersonalized.AppEUI = appEUI
	m.InStorePersonalized.DevAddr = devAddr
	m.InStorePersonalized.AppSKey = appSKey
	m.InStorePersonalized.NwkSKey = nwkSKey
	return m.Failures["StorePersonalized"]
}

// Close implements the DevStorage Interface
func (m *MockDevStorage) Close() error {
	m.InClose.Called = true
	return m.Failures["Close"]
}

// MockPktStorage mocks the PktStorage interface
type MockPktStorage struct {
	Failures map[string]error
	InPull   struct {
		AppEUI []byte
		DevEUI []byte
	}
	OutPull struct {
		Entry pktEntry
	}
	InPush struct {
		AppEUI  []byte
		DevEUI  []byte
		Payload pktEntry
	}
	InClose struct {
		Called bool
	}
}

// NewMockPktStorage creates a new MockPktStorage
func NewMockPktStorage() *MockPktStorage {
	return &MockPktStorage{
		Failures: make(map[string]error),
	}
}

// Close implements the PktStorage Interface
func (m *MockPktStorage) Close() error {
	m.InClose.Called = true
	return m.Failures["Close"]
}

// Push implements the PktStorage interface
func (m *MockPktStorage) Push(appEUI []byte, devEUI []byte, payload pktEntry) error {
	m.InPush.AppEUI = appEUI
	m.InPush.DevEUI = devEUI
	m.InPush.Payload = payload
	return m.Failures["Push"]
}

// Push implements the PktStorage interface
func (m *MockPktStorage) Pull(appEUI []byte, devEUI []byte) (pktEntry, error) {
	m.InPull.AppEUI = appEUI
	m.InPull.DevEUI = devEUI
	return m.OutPull.Entry, m.Failures["Pull"]
}

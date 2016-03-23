// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package handler

// NOTE: All the code below could be generated

// MockDevStorage mocks the DevStorage interface
type MockDevStorage struct {
	Failures map[string]error
	InRead   struct {
		AppEUI []byte
		DevEUI []byte
	}
	OutRead struct {
		Entry devEntry
	}
	InReadAll struct {
		AppEUI []byte
	}
	OutReadAll struct {
		Entries []devEntry
	}
	InUpsert struct {
		Entry devEntry
	}
	InDone struct {
		Called bool
	}
}

// NewMockDevStorage creates a new MockDevStorage
func NewMockDevStorage() *MockDevStorage {
	return &MockDevStorage{
		Failures: make(map[string]error),
	}
}

// read implements the DevStorage interface
func (m *MockDevStorage) read(appEUI []byte, devEUI []byte) (devEntry, error) {
	m.InRead.AppEUI = appEUI
	m.InRead.DevEUI = devEUI
	return m.OutRead.Entry, m.Failures["read"]
}

// readAll implements the DevStorage interface
func (m *MockDevStorage) readAll(appEUI []byte) ([]devEntry, error) {
	m.InReadAll.AppEUI = appEUI
	return m.OutReadAll.Entries, m.Failures["readAll"]
}

// upsert implements the DevStorage interface
func (m *MockDevStorage) upsert(entry devEntry) error {
	m.InUpsert.Entry = entry
	return m.Failures["upsert"]
}

// done implements the DevStorage Interface
func (m *MockDevStorage) done() error {
	m.InDone.Called = true
	return m.Failures["done"]
}

// MockPktStorage mocks the PktStorage interface
type MockPktStorage struct {
	Failures  map[string]error
	InDequeue struct {
		AppEUI []byte
		DevEUI []byte
	}
	OutDequeue struct {
		Entry pktEntry
	}
	InPeek struct {
		AppEUI []byte
		DevEUI []byte
	}
	OutPeek struct {
		Entry pktEntry
	}
	InEnqueue struct {
		Entry pktEntry
	}
	InDone struct {
		Called bool
	}
}

// NewMockPktStorage creates a new MockPktStorage
func NewMockPktStorage() *MockPktStorage {
	return &MockPktStorage{
		Failures: make(map[string]error),
	}
}

// done implements the PktStorage Interface
func (m *MockPktStorage) done() error {
	m.InDone.Called = true
	return m.Failures["done"]
}

// enqueue implements the PktStorage interface
func (m *MockPktStorage) enqueue(entry pktEntry) error {
	m.InEnqueue.Entry = entry
	return m.Failures["enqueue"]
}

// dequeue implements the PktStorage interface
func (m *MockPktStorage) dequeue(appEUI []byte, devEUI []byte) (pktEntry, error) {
	m.InDequeue.AppEUI = appEUI
	m.InDequeue.DevEUI = devEUI
	return m.OutDequeue.Entry, m.Failures["dequeue"]
}

// peek implements the PktStorage interface
func (m *MockPktStorage) peek(appEUI []byte, devEUI []byte) (pktEntry, error) {
	m.InPeek.AppEUI = appEUI
	m.InPeek.DevEUI = devEUI
	return m.OutPeek.Entry, m.Failures["peek"]
}

// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package router

// MockBrkStorage mocks the router.BrkStorage interface
type MockBrkStorage struct {
	Failures map[string]error
	InRead   struct {
		DevAddr []byte
	}
	OutRead struct {
		Entries []brkEntry
	}
	InCreate struct {
		Entry brkEntry
	}
	InDone struct {
		Called bool
	}
}

// NewMockBrkStorage creates a new mock BrkStorage
func NewMockBrkStorage() *MockBrkStorage {
	return &MockBrkStorage{
		Failures: make(map[string]error),
	}
}

// read implements the router.BrkStorage interface
func (m *MockBrkStorage) read(devAddr []byte) ([]brkEntry, error) {
	m.InRead.DevAddr = devAddr
	return m.OutRead.Entries, m.Failures["read"]
}

// create implements the router.BrkStorage interface
func (m *MockBrkStorage) create(entry brkEntry) error {
	m.InCreate.Entry = entry
	return m.Failures["create"]
}

// done implements the router.BrkStorage interface
func (m *MockBrkStorage) done() error {
	m.InDone.Called = true
	return m.Failures["done"]
}

// MockGtwStorage mocks the router.GtwStorage interface
type MockGtwStorage struct {
	Failures map[string]error
	InRead   struct {
		DevAddr []byte
	}
	OutRead struct {
		Entry gtwEntry
	}
	InUpsert struct {
		Entry gtwEntry
	}
	InDone struct {
		Called bool
	}
}

// NewMockGtwStorage Upserts a new mock GtwStorage
func NewMockGtwStorage() *MockGtwStorage {
	return &MockGtwStorage{
		Failures: make(map[string]error),
	}
}

// read implements the router.GtwStorage interface
func (m *MockGtwStorage) read(devAddr []byte) (gtwEntry, error) {
	m.InRead.DevAddr = devAddr
	return m.OutRead.Entry, m.Failures["read"]
}

// Upsert implements the router.GtwStorage interface
func (m *MockGtwStorage) upsert(entry gtwEntry) error {
	m.InUpsert.Entry = entry
	return m.Failures["upsert"]
}

// done implements the router.GtwStorage interface
func (m *MockGtwStorage) done() error {
	m.InDone.Called = true
	return m.Failures["done"]
}

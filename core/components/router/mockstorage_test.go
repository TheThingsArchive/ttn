// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package router

import (
	"github.com/TheThingsNetwork/ttn/core"
)

// MockStorage mocks the router.Storage interface
type MockStorage struct {
	Failures map[string]error
	InLookup struct {
		DevAddr []byte
	}
	OutLookup struct {
		Entries []entry
	}
	InStore struct {
		DevAddr     []byte
		BrokerIndex int
	}
	InLookupStats struct {
		GID []byte
	}
	OutLookupStats struct {
		Metadata core.StatsMetadata
	}
	InUpdateStats struct {
		GID      []byte
		Metadata core.StatsMetadata
	}
	InClose struct {
		Called bool
	}
}

// NewMockStorage creates a new mock storage
func NewMockStorage() *MockStorage {
	return &MockStorage{
		Failures: make(map[string]error),
	}
}

// Lookup implements the router.Storage interface
func (m *MockStorage) Lookup(devAddr []byte) ([]entry, error) {
	m.InLookup.DevAddr = devAddr
	return m.OutLookup.Entries, m.Failures["Lookup"]
}

// Store implements the router.Storage interface
func (m *MockStorage) Store(devAddr []byte, brokerIndex int) error {
	m.InStore.DevAddr = devAddr
	m.InStore.BrokerIndex = brokerIndex
	return m.Failures["Store"]
}

// LookupStats implements the router.Storage interface
func (m *MockStorage) LookupStats(gid []byte) (core.StatsMetadata, error) {
	m.InLookupStats.GID = gid
	return m.OutLookupStats.Metadata, m.Failures["LookupStats"]
}

// UpdateStats implements the router.Storage interface
func (m *MockStorage) UpdateStats(gid []byte, metadata core.StatsMetadata) error {
	m.InUpdateStats.GID = gid
	m.InUpdateStats.Metadata = metadata
	return m.Failures["UpdateStats"]
}

// Close implements the router.Storage interface
func (m *MockStorage) Close() error {
	m.InClose.Called = true
	return m.Failures["Close"]
}

// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package router

import (
	"time"

	. "github.com/TheThingsNetwork/ttn/core"
	"github.com/brocaar/lorawan"
)

// MockStorage to fake the Storage interface
type mockStorage struct {
	Failures  map[string]error
	InLookup  lorawan.EUI64
	OutLookup entry
	InStore   RRegistration
}

func newMockStorage() *mockStorage {
	return &mockStorage{
		Failures: make(map[string]error),
		OutLookup: entry{
			Recipient: []byte("MockStorageRecipient"),
			until:     time.Date(2016, 2, 3, 14, 16, 22, 0, time.UTC),
		},
	}
}

func (s *mockStorage) Lookup(devEUI lorawan.EUI64) (entry, error) {
	s.InLookup = devEUI
	if s.Failures["Lookup"] != nil {
		return entry{}, s.Failures["Lookup"]
	}
	return s.OutLookup, nil
}

func (s *mockStorage) Store(reg RRegistration) error {
	s.InStore = reg
	return s.Failures["Store"]
}

func (s *mockStorage) Close() error {
	return s.Failures["Close"]
}

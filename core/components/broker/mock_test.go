// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package broker

import (
	. "github.com/TheThingsNetwork/ttn/core"
	"github.com/brocaar/lorawan"
)

type mockStorage struct {
	Failures  map[string]error
	InLookup  lorawan.EUI64
	OutLookup []entry
	InStore   BRegistration
}

func newMockStorage() *mockStorage {
	return &mockStorage{
		Failures: make(map[string]error),
		OutLookup: []entry{
			{
				Recipient: []byte("MockStorageRecipient"),
				AppEUI:    lorawan.EUI64([8]byte{1, 1, 2, 2, 3, 3, 1, 1}),
				DevEUI:    lorawan.EUI64([8]byte{2, 3, 4, 5, 6, 7, 5, 4}),
				NwkSKey:   lorawan.AES128Key([16]byte{1, 1, 1, 1, 1, 1, 1, 2, 2, 2, 3, 4, 4, 5, 5}),
			},
		},
	}
}

func (s *mockStorage) Lookup(devEUI lorawan.EUI64) ([]entry, error) {
	s.InLookup = devEUI
	if s.Failures["Lookup"] != nil {
		return nil, s.Failures["Lookup"]
	}
	return s.OutLookup, nil
}

func (s *mockStorage) Store(reg BRegistration) error {
	s.InStore = reg
	return s.Failures["Store"]
}

func (s *mockStorage) Close() error {
	return s.Failures["Close"]
}

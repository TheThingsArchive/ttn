// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package handler

import (
	. "github.com/TheThingsNetwork/ttn/core"
	"github.com/brocaar/lorawan"
)

// mockDevStorage implements the handler.DevStorage interface
//
// It declares a `Failures` attributes that can be used to
// simulate failures on demand, associating the name of the method
// which needs to fail with the actual failure.
//
// It also stores the last arguments of each function call in appropriated
// attributes. Because there's no computation going on, the expected / wanted
// responses should also be defined. Default values are provided but can be changed
// if needed.
type mockDevStorage struct {
	Failures            map[string]error
	OutLookup           devEntry
	InLookupAppEUI      lorawan.EUI64
	InLookupDevEUI      lorawan.EUI64
	InStorePersonalized HRegistration
	InStoreActivated    HRegistration
}

func newMockDevStorage() *mockDevStorage {
	return &mockDevStorage{
		Failures: make(map[string]error),
		OutLookup: devEntry{
			Recipient: []byte("MockDevStorageRecipient"),
			DevAddr:   lorawan.DevAddr([4]byte{9, 9, 1, 4}),
			AppSKey:   lorawan.AES128Key([16]byte{6, 6, 4, 3, 2, 2, 0, 9, 8, 7, 6, 3, 1, 9, 6, 14}),
			NwkSKey:   lorawan.AES128Key([16]byte{7, 2, 3, 3, 5, 6, 7, 0, 9, 0, 1, 2, 7, 4, 5, 5}),
		},
	}
}

func (s *mockDevStorage) Lookup(appEUI lorawan.EUI64, devEUI lorawan.EUI64) (devEntry, error) {
	s.InLookupAppEUI = appEUI
	s.InLookupDevEUI = devEUI
	if s.Failures["Lookup"] != nil {
		return devEntry{}, s.Failures["Lookup"]
	}
	return s.OutLookup, nil
}

func (s *mockDevStorage) StorePersonalized(r HRegistration) error {
	s.InStorePersonalized = r
	return s.Failures["StorePersonalized"]
}

func (s *mockDevStorage) StoreActivated(r HRegistration) error {
	s.InStoreActivated = r
	return s.Failures["StoreActivated"]
}

func (s *mockDevStorage) Close() error {
	return s.Failures["Close"]
}

// mockPktStorage implements the handler.PktStorage interface
//
// It declares a `Failures` attributes that can be used to
// simulate failures on demand, associating the name of the method
// which needs to fail with the actual failure.
//
// It also stores the last arguments of each function call in appropriated
// attributes. Because there's no computation going on, the expected / wanted
// responses should also be defined. Default values are provided but can be changed
// if needed.
type mockPktStorage struct {
	Failures     map[string]error
	OutPull      APacket
	InPullAppEUI lorawan.EUI64
	InPullDevEUI lorawan.EUI64
	InPush       APacket
}

func newMockPktStorage() *mockPktStorage {
	return &mockPktStorage{
		Failures: make(map[string]error),
	}
}

func (s *mockPktStorage) Push(p APacket) error {
	s.InPush = p
	return s.Failures["Push"]
}

func (s *mockPktStorage) Pull(appEUI lorawan.EUI64, devEUI lorawan.EUI64) (APacket, error) {
	s.InPullAppEUI = appEUI
	s.InPullDevEUI = devEUI
	if s.Failures["Pull"] != nil {
		return nil, s.Failures["Pull"]
	}
	return s.OutPull, nil
}

func (s *mockPktStorage) Close() error {
	return s.Failures["Close"]
}

// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package handler

import (
	//	"os"
	//	"reflect"
	"testing"
	//	"time"
	//
	. "github.com/TheThingsNetwork/ttn/core"
	//	"github.com/TheThingsNetwork/ttn/utils/errors"
	. "github.com/TheThingsNetwork/ttn/utils/errors/checks"
	//	"github.com/TheThingsNetwork/ttn/utils/pointer"
	. "github.com/TheThingsNetwork/ttn/utils/testing"
	"github.com/brocaar/lorawan"
)

func TestRegister(t *testing.T) {
	{
		Desc(t, "Registration valid HRegistration")

		devStorage := newMockDevStorage()
		pktStorage := newMockPktStorage()
		handler := New(devStorage, pktStorage, GetLogger(t, "Handler"))
		an := newMockAckNacker()
		r := newTestRegistration(
			[8]byte{1, 1, 1, 1, 1, 1, 1, 1},
			[8]byte{2, 2, 2, 2, 2, 2, 2, 2},
		)

		err := handler.Register(r, an)
		CheckErrors(t, nil, err)
	}
}

// ----- TYPE utilities

//
// MOCK DEV STORAGE
//
type mockDevStorage struct {
	Failures     map[string]error
	LookupEntry  devEntry
	Personalized HRegistration
	Activated    HRegistration
}

func newMockDevStorage(failures ...string) *mockDevStorage {
	return &mockDevStorage{
		Failures: make(map[string]error),
	}
}

func (s mockDevStorage) Lookup(appEUI lorawan.EUI64, devEUI lorawan.EUI64) (devEntry, error) {
	if s.Failures["Lookup"] != nil {
		return devEntry{}, s.Failures["Lookup"]
	}
	return s.LookupEntry, nil
}

func (s *mockDevStorage) StorePersonalized(r HRegistration) error {
	s.Personalized = r
	return s.Failures["StorePersonalized"]
}

func (s *mockDevStorage) StoreActivated(r HRegistration) error {
	s.Activated = r
	return s.Failures["StoreActivated"]
}

func (s mockDevStorage) Close() error {
	return s.Failures["Close"]
}

//
// MOCK PKT STORAGE
//
type mockPktStorage struct {
	Failures  map[string]error
	PullEntry APacket
	Pushed    APacket
}

func newMockPktStorage() *mockPktStorage {
	return &mockPktStorage{
		Failures: make(map[string]error),
	}
}

func (s *mockPktStorage) Push(p APacket) error {
	s.Pushed = p
	return s.Failures["Push"]
}

func (s *mockPktStorage) Pull(appEUI lorawan.EUI64, devEUI lorawan.EUI64) (APacket, error) {
	if s.Failures["Pull"] != nil {
		return nil, s.Failures["Pull"]
	}
	return s.PullEntry, nil
}

func (s *mockPktStorage) Close() error {
	return s.Failures["Close"]
}

//
// MOCK ACK/NACKER
//
type mockAckNacker struct{}

func newMockAckNacker() *mockAckNacker {
	return &mockAckNacker{}
}

func (an mockAckNacker) Ack(p Packet) error {
	return nil
}

func (an mockAckNacker) Nack() error {
	return nil
}

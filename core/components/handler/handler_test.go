// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package handler

import (
	"reflect"
	"testing"
	//	"time"

	. "github.com/TheThingsNetwork/ttn/core"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	. "github.com/TheThingsNetwork/ttn/utils/errors/checks"
	"github.com/TheThingsNetwork/ttn/utils/pointer"
	. "github.com/TheThingsNetwork/ttn/utils/testing"
	"github.com/brocaar/lorawan"
)

func TestRegister(t *testing.T) {
	{
		Desc(t, "Register valid HRegistration")

		devStorage := newMockDevStorage()
		pktStorage := newMockPktStorage()
		an := newMockAckNacker()
		handler := New(devStorage, pktStorage, GetLogger(t, "Handler"))
		r := newTestRegistration(
			[8]byte{1, 1, 1, 1, 1, 1, 1, 1},
			[8]byte{2, 2, 2, 2, 2, 2, 2, 2},
		)

		err := handler.Register(r, an)

		CheckErrors(t, nil, err)
		CheckPushed(t, nil, pktStorage.Pushed)
		CheckPersonalized(t, r, devStorage.Personalized)
	}

	// --------------------

	{
		Desc(t, "Register invalid HRegistration")

		devStorage := newMockDevStorage()
		pktStorage := newMockPktStorage()
		an := newMockAckNacker()
		handler := New(devStorage, pktStorage, GetLogger(t, "Handler"))

		err := handler.Register(nil, an)

		CheckErrors(t, pointer.String(string(errors.Structural)), err)
		CheckPushed(t, nil, pktStorage.Pushed)
		CheckPersonalized(t, nil, devStorage.Personalized)
	}

	// --------------------

	{
		Desc(t, "Register valid HRegistration | devStorage fails")

		devStorage := newMockDevStorage()
		pktStorage := newMockPktStorage()
		an := newMockAckNacker()
		handler := New(devStorage, pktStorage, GetLogger(t, "Handler"))
		r := newTestRegistration(
			[8]byte{1, 1, 1, 1, 1, 1, 1, 1},
			[8]byte{2, 2, 2, 2, 2, 2, 2, 2},
		)

		devStorage.Failures["StorePersonalized"] = errors.New(errors.Operational, "Mock Error")
		err := handler.Register(r, an)

		CheckErrors(t, pointer.String(string(errors.Operational)), err)
		CheckPushed(t, nil, pktStorage.Pushed)
		CheckPersonalized(t, r, devStorage.Personalized)
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

//
// MOCK ADAPTER
//
type mockAdapter struct {
	Failures       map[string]error
	SentPkt        Packet
	SentRecipients []Recipient
	SendData       []byte
	Recipient      Recipient
	NextPacket     []byte
	NextAckNacker  AckNacker
	NextReg        Registration
}

func newMockAdapter() *mockAdapter {
	return &mockAdapter{
		Failures: make(map[string]error),
	}
}

func (a mockAdapter) Send(p Packet, r ...Recipient) ([]byte, error) {
	a.SentPkt = p
	a.SentRecipients = r
	if a.Failures["Send"] != nil {
		return nil, a.Failures["Send"]
	}
	return a.SendData, nil
}

func (a mockAdapter) GetRecipient(raw []byte) (Recipient, error) {
	if a.Failures["GetRecipient"] != nil {
		return nil, a.Failures["Send"]
	}
	return a.Recipient, nil
}

func (a *mockAdapter) Next() ([]byte, AckNacker, error) {
	if a.Failures["Next"] != nil {
		return nil, nil, a.Failures["Next"]
	}
	return a.NextPacket, a.NextAckNacker, nil
}

func (a *mockAdapter) NextRegistration() (Registration, AckNacker, error) {
	if a.Failures["NextRegistration"] != nil {
		return nil, nil, a.Failures["NextRegistration"]
	}
	return a.NextReg, a.NextAckNacker, nil
}

// ----- CHECK utilities
func CheckPushed(t *testing.T, want APacket, got APacket) {
	if !reflect.DeepEqual(want, got) {
		Ko(t, "Pushed packet does not match expectations.\nWant: %s\nGot:  %s", want, got)
	}
	Ok(t, "Check Pushed")
}

func CheckPersonalized(t *testing.T, want HRegistration, got HRegistration) {
	if !reflect.DeepEqual(want, got) {
		Ko(t, "Personalized packet does not match expectations.\nWant: %s\nGot:  %s", want, got)
	}
	Ok(t, "Check Personalized")
}

// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package router

import (
	"time"

	. "github.com/TheThingsNetwork/ttn/core"
	"github.com/TheThingsNetwork/ttn/core/dutycycle"
	. "github.com/TheThingsNetwork/ttn/core/mocks"
	"github.com/brocaar/lorawan"
)

// MockStorage to fake the Storage interface
type mockStorage struct {
	Failures  map[string]error
	InLookup  lorawan.EUI64
	OutLookup []entry
	InStore   RRegistration
}

func newMockStorage() *mockStorage {
	return &mockStorage{
		Failures: make(map[string]error),
		OutLookup: []entry{
			{
				Recipient: []byte("MockStorageRecipient"),
				until:     time.Date(2016, 2, 3, 14, 16, 22, 0, time.UTC),
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

func (s *mockStorage) Store(reg RRegistration) error {
	s.InStore = reg
	return s.Failures["Store"]
}

func (s *mockStorage) Close() error {
	return s.Failures["Close"]
}

// mockDutyManager implements the dutycycle.DutyManager interface
type mockDutyManager struct {
	Failures     map[string]error
	InUpdateId   []byte
	InUpdateFreq float64
	InUpdateSize uint
	InUpdateDatr string
	InUpdateCodr string
	InLookupId   []byte
	OutLookup    dutycycle.Cycles
}

func newMockDutyManager() *mockDutyManager {
	return &mockDutyManager{
		Failures:  make(map[string]error),
		OutLookup: make(dutycycle.Cycles),
	}
}

func (m *mockDutyManager) Update(id []byte, freq float64, size uint, datr string, codr string) error {
	m.InUpdateId = id
	m.InUpdateFreq = freq
	m.InUpdateSize = size
	m.InUpdateDatr = datr
	m.InUpdateCodr = codr
	return m.Failures["Update"]
}

func (m *mockDutyManager) Lookup(id []byte) (dutycycle.Cycles, error) {
	m.InLookupId = id
	if m.Failures["Lookup"] != nil {
		return nil, m.Failures["Lookup"]
	}
	return m.OutLookup, nil
}

func (m *mockDutyManager) Close() error {
	return m.Failures["Close"]
}

// MockRouterAdapter extends functionality of the mocks.MockAdapter.
//
// A list of failures can be defined to handle successive call to a method (at each call, an error
// get out from the list)
type mockRouterAdapter struct {
	Failures            map[string][]error
	InSendPacket        Packet
	InSendRecipients    []Recipient
	InGetRecipient      []byte
	OutSend             []byte
	OutGetRecipient     Recipient
	OutNextPacket       []byte
	OutNextAckNacker    AckNacker
	OutNextRegReg       Registration
	OutNextRegAckNacker AckNacker
}

func newMockRouterAdapter() *mockRouterAdapter {
	return &mockRouterAdapter{
		Failures:            make(map[string][]error),
		OutSend:             []byte("MockAdapterSend"),
		OutGetRecipient:     NewMockRecipient(),
		OutNextPacket:       []byte("MockAdapterNextPacket"),
		OutNextAckNacker:    NewMockAckNacker(),
		OutNextRegReg:       NewMockHRegistration(),
		OutNextRegAckNacker: NewMockAckNacker(),
	}
}

func (a *mockRouterAdapter) Send(p Packet, r ...Recipient) ([]byte, error) {
	a.InSendPacket = p
	a.InSendRecipients = r
	if len(a.Failures["Send"]) > 0 {
		err := a.Failures["Send"][0]
		a.Failures["Send"] = a.Failures["Send"][1:]
		return nil, err
	}
	return a.OutSend, nil
}

func (a *mockRouterAdapter) GetRecipient(raw []byte) (Recipient, error) {
	a.InGetRecipient = raw
	if len(a.Failures["GetRecipient"]) > 0 {
		err := a.Failures["GetRecipient"][0]
		a.Failures["GetRecipient"] = a.Failures["GetRecipient"][1:]
		return nil, err
	}
	return a.OutGetRecipient, nil
}

func (a *mockRouterAdapter) Next() ([]byte, AckNacker, error) {
	if len(a.Failures["Next"]) > 0 {
		err := a.Failures["Next"][0]
		a.Failures["Next"] = a.Failures["Next"][1:]
		return nil, nil, err
	}
	return a.OutNextPacket, a.OutNextAckNacker, nil
}

func (a *mockRouterAdapter) NextRegistration() (Registration, AckNacker, error) {
	if len(a.Failures["NextRegistration"]) > 0 {
		err := a.Failures["NextRegistration"][0]
		a.Failures["NextRegistration"] = a.Failures["NextRegistration"][1:]
		return nil, nil, err
	}
	return a.OutNextRegReg, a.OutNextRegAckNacker, nil
}

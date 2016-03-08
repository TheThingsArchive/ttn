// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package mocks

import (
	"fmt"
	"reflect"
	"testing"

	. "github.com/TheThingsNetwork/ttn/core"
	"github.com/TheThingsNetwork/ttn/utils/pointer"
	. "github.com/TheThingsNetwork/ttn/utils/testing"
	"github.com/brocaar/lorawan"
)

// MockRecipient implements the core.Recipient interface
//
// It declares a `Failures` attributes that can be used to
// simulate failures on demand, associating the name of the method
// which needs to fail with the actual failure.
//
// It also stores the last arguments of each function call in appropriated
// attributes. Because there's no computation going on, the expected / wanted
// responses should also be defined. Default values are provided but can be changed
// if needed.
type MockRecipient struct {
	Failures          map[string]error
	InUnmarshalBinary []byte // Data received by UnmarshalBinary()
	OutMarshalBinary  []byte // Data spit out by MarshalBinary()
}

// NewMockRecipient constructs a new mock recipient.
func NewMockRecipient() *MockRecipient {
	return &MockRecipient{
		OutMarshalBinary: []byte("MockRecipientData"),
		Failures:         make(map[string]error),
	}
}

func (r *MockRecipient) MarshalBinary() ([]byte, error) {
	if r.Failures["MarshalBinary"] != nil {
		return nil, r.Failures["MarshalBinary"]
	}
	return r.OutMarshalBinary, nil
}

func (r *MockRecipient) UnmarshalBinary(data []byte) error {
	r.InUnmarshalBinary = data
	if r.Failures["UnmarshalBinary"] != nil {
		return r.Failures["UnmarshalBinary"]
	}
	return nil
}

type MockJSONRecipient struct {
	*MockRecipient
	InUnmarshalJSON []byte
	OutMarshalJSON  []byte
}

func NewMockJSONRecipient() *MockJSONRecipient {
	return &MockJSONRecipient{
		MockRecipient:  NewMockRecipient(),
		OutMarshalJSON: []byte(`{"out":"MockJSONRecipientData"}`),
	}
}

func (r *MockJSONRecipient) MarshalJSON() ([]byte, error) {
	if r.Failures["MarshalJSON"] != nil {
		return nil, r.Failures["MarshalJSON"]
	}
	return r.OutMarshalJSON, nil
}

func (r *MockJSONRecipient) UnmarshalJSON(data []byte) error {
	r.InUnmarshalJSON = data
	if r.Failures["UnmarshalJSON"] != nil {
		return r.Failures["UnmarshalJSON"]
	}
	return nil
}

// MockRegistration implements the core.Registration interface
//
// It also stores the last arguments of each function call in appropriated
// attributes. Because there's no computation going on, the expected / wanted
// responses should also be defined. Default values are provided but can be changed
// if needed.
type MockRegistration struct {
	Failures       map[string]error
	OutRecipient   Recipient
	OutMarshalJSON []byte
}

func NewMockRegistration() MockRegistration {
	return MockRegistration{
		Failures:       make(map[string]error),
		OutRecipient:   NewMockRecipient(),
		OutMarshalJSON: []byte(`{"out":"MockRegistration"}`),
	}
}

func (r MockRegistration) RawRecipient() []byte {
	data, _ := r.Recipient().MarshalBinary()
	return data
}

func (r MockRegistration) Recipient() Recipient {
	return r.OutRecipient
}

func (r MockRegistration) MarshalJSON() ([]byte, error) {
	if r.Failures["MarshalJSON"] != nil {
		return nil, r.Failures["MarshalJSON"]
	}
	return r.OutMarshalJSON, nil
}

// MockARegistration implements the core.ARegistration interface
//
// It also stores the last arguments of each function call in appropriated
// attributes. Because there's no computation going on, the expected / wanted
// responses should also be defined. Default values are provided but can be changed
// if needed.
type MockARegistration struct {
	MockRegistration
	OutAppEUI lorawan.EUI64
}

func NewMockARegistration() MockARegistration {
	return MockARegistration{
		MockRegistration: NewMockRegistration(),
		OutAppEUI:        lorawan.EUI64([8]byte{9, 0, 9, 2, 2, 2, 3, 4}),
	}
}

func (r MockARegistration) AppEUI() lorawan.EUI64 {
	return r.OutAppEUI
}

// MockRRegistration implements the core.RRegistration interface
//
// It also stores the last arguments of each function call in appropriated
// attributes. Because there's no computation going on, the expected / wanted
// responses should also be defined. Default values are provided but can be changed
// if needed.
type MockRRegistration struct {
	MockRegistration
	OutDevEUI lorawan.EUI64
}

func NewMockRRegistration() MockRRegistration {
	return MockRRegistration{
		MockRegistration: NewMockRegistration(),
		OutDevEUI:        lorawan.EUI64([8]byte{2, 2, 2, 2, 2, 2, 2, 2}),
	}
}

func (r MockRRegistration) DevEUI() lorawan.EUI64 {
	return r.OutDevEUI
}

// MockBRegistration implements the core.BRegistration interface
//
// It also stores the last arguments of each function call in appropriated
// attributes. Because there's no computation going on, the expected / wanted
// responses should also be defined. Default values are provided but can be changed
// if needed.
type MockBRegistration struct {
	MockRRegistration
	OutAppEUI  lorawan.EUI64
	OutNwkSKey lorawan.AES128Key
}

func NewMockBRegistration() MockBRegistration {
	return MockBRegistration{
		MockRRegistration: NewMockRRegistration(),
		OutAppEUI:         lorawan.EUI64([8]byte{1, 1, 1, 1, 1, 1, 1, 1}),
		OutNwkSKey:        lorawan.AES128Key([16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6}),
	}
}

func (r MockBRegistration) AppEUI() lorawan.EUI64 {
	return r.OutAppEUI
}

func (r MockBRegistration) NwkSKey() lorawan.AES128Key {
	return r.OutNwkSKey
}

// MockHRegistration implements the core.HRegistration interface
//
// It also stores the last arguments of each function call in appropriated
// attributes. Because there's no computation going on, the expected / wanted
// responses should also be defined. Default values are provided but can be changed
// if needed.
type MockHRegistration struct {
	MockBRegistration
	OutAppSKey lorawan.AES128Key
}

func NewMockHRegistration() MockHRegistration {
	return MockHRegistration{
		MockBRegistration: NewMockBRegistration(),
		OutAppSKey:        lorawan.AES128Key([16]byte{6, 5, 4, 3, 2, 1, 0, 9, 8, 7, 6, 5, 4, 3, 2, 1}),
	}
}

func (r MockHRegistration) AppSKey() lorawan.AES128Key {
	return r.OutAppSKey
}

// MockAckNacker implements the core.AckNacker interface
//
// It declares a `Failures` attributes that can be used to
// simulate failures on demand, associating the name of the method
// which needs to fail with the actual failure.
//
// It also stores the last arguments of each function call in appropriated
// attributes. Because there's no computation going on, the expected / wanted
// responses should also be defined. Default values are provided but can be changed
// if needed.
type MockAckNacker struct {
	InAck struct {
		Ack    *bool
		Packet Packet
		Error  error
	}
}

func NewMockAckNacker() *MockAckNacker {
	return &MockAckNacker{}
}

func (an *MockAckNacker) Ack(p Packet) error {
	an.InAck = struct {
		Ack    *bool
		Packet Packet
		Error  error
	}{
		Ack:    pointer.Bool(true),
		Packet: p,
	}
	return nil
}

func (an *MockAckNacker) Nack(err error) error {
	an.InAck = struct {
		Ack    *bool
		Packet Packet
		Error  error
	}{
		Ack:   pointer.Bool(false),
		Error: err,
	}
	return nil
}

// MockAdapter implements the core.Adapter interface
//
// It declares a `Failures` attributes that can be used to
// simulate failures on demand, associating the name of the method
// which needs to fail with the actual failure.
//
// It also stores the last arguments of each function call in appropriated
// attributes. Because there's no computation going on, the expected / wanted
// responses should also be defined. Default values are provided but can be changed
// if needed.
type MockAdapter struct {
	Failures            map[string]error
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

func NewMockAdapter() *MockAdapter {
	return &MockAdapter{
		Failures:            make(map[string]error),
		OutSend:             []byte("MockAdapterSend"),
		OutGetRecipient:     NewMockRecipient(),
		OutNextPacket:       []byte("MockAdapterNextPacket"),
		OutNextAckNacker:    NewMockAckNacker(),
		OutNextRegReg:       NewMockHRegistration(),
		OutNextRegAckNacker: NewMockAckNacker(),
	}
}

func (a *MockAdapter) Send(p Packet, r ...Recipient) ([]byte, error) {
	a.InSendPacket = p
	a.InSendRecipients = r
	if a.Failures["Send"] != nil {
		return nil, a.Failures["Send"]
	}
	return a.OutSend, nil
}

func (a *MockAdapter) GetRecipient(raw []byte) (Recipient, error) {
	a.InGetRecipient = raw
	if a.Failures["GetRecipient"] != nil {
		return nil, a.Failures["GetRecipient"]
	}
	return a.OutGetRecipient, nil
}

func (a *MockAdapter) Next() ([]byte, AckNacker, error) {
	if a.Failures["Next"] != nil {
		return nil, nil, a.Failures["Next"]
	}
	return a.OutNextPacket, a.OutNextAckNacker, nil
}

func (a *MockAdapter) NextRegistration() (Registration, AckNacker, error) {
	if a.Failures["NextRegistration"] != nil {
		return nil, nil, a.Failures["NextRegistration"]
	}
	return a.OutNextRegReg, a.OutNextRegAckNacker, nil
}

// MockSubscriber implements the core.Subscriber interface
//
// It declares a `Failures` attributes that can be used to
// simulate failures on demand, associating the name of the method
// which needs to fail with the actual failure.
//
// It also stores the last arguments of each function call in appropriated
// attributes. Because there's no computation going on, the expected / wanted
// responses should also be defined. Default values are provided but can be changed
// if needed.
type MockSubscriber struct {
	Failures                map[string]error
	InSubscribeRegistration Registration
}

func NewMockSubscriber() *MockSubscriber {
	return &MockSubscriber{
		Failures: make(map[string]error),
	}
}

func (s *MockSubscriber) Subscribe(reg Registration) error {
	s.InSubscribeRegistration = reg
	return s.Failures["Subscribe"]
}

// MockPacket implements the core.Packet interface
//
// It declares a `Failures` attributes that can be used to
// simulate failures on demand, associating the name of the method
// which needs to fail with the actual failure.
//
// It also stores the last arguments of each function call in appropriated
// attributes. Because there's no computation going on, the expected / wanted
// responses should also be defined. Default values are provided but can be changed
// if needed.
type MockPacket struct {
	Failures         map[string]error
	OutMarshalBinary []byte
	OutString        string
	OutDevEUI        lorawan.EUI64
}

func NewMockPacket() *MockPacket {
	return &MockPacket{
		Failures:         make(map[string]error),
		OutMarshalBinary: []byte("MockPacketBinary"),
		OutString:        "MockPacket",
		OutDevEUI:        lorawan.EUI64([8]byte{1, 2, 1, 2, 4, 4, 5, 5}),
	}
}

func (p *MockPacket) DevEUI() lorawan.EUI64 {
	return p.OutDevEUI
}

func (p *MockPacket) MarshalBinary() ([]byte, error) {
	if p.Failures["MarshalBinary"] != nil {
		return nil, p.Failures["MarshalBinary"]
	}
	return p.OutMarshalBinary, nil
}

func (p *MockPacket) String() string {
	return p.OutString
}

// ----- CHECK utilities

func Check(t *testing.T, want, got interface{}, name string) {
	if !reflect.DeepEqual(want, got) {
		Ko(t, "%s don't match expectations.\nWant: %+v\nGot:  %+v", name, want, got)
	}
	Ok(t, fmt.Sprintf("Check %s", name))
}

func CheckAcks(t *testing.T, want interface{}, gotItf interface{}) {
	got := gotItf.(struct {
		Ack    *bool
		Packet Packet
		Error  error
	})

	if got.Ack == nil {
		Ko(t, "Invalid ack got: %+v", got)
	}

	switch want.(type) {
	case bool:
		Check(t, want.(bool), *(got.Ack), "Acks")
	case Packet:
		Check(t, want.(Packet), got.Packet, "Acks")
	default:
		panic("Unexpect ack wanted")
	}
}

func CheckRecipients(t *testing.T, want []Recipient, got []Recipient) {
	Check(t, want, got, "Recipients")
}

func CheckSent(t *testing.T, want Packet, got Packet) {
	Check(t, want, got, "Sent")
}

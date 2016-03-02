// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package mock

import (
	. "github.com/TheThingsNetwork/ttn/core"
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

// MockRegistration implements the core.Recipient interface
//
// It also stores the last arguments of each function call in appropriated
// attributes. Because there's no computation going on, the expected / wanted
// responses should also be defined. Default values are provided but can be changed
// if needed.
type MockRegistration struct {
	OutAppEUI    lorawan.EUI64
	OutDevEUI    lorawan.EUI64
	OutNwkSKey   lorawan.AES128Key
	OutAppSKey   lorawan.AES128Key
	OutRecipient Recipient
}

func NewMockRegistration() *MockRegistration {
	return &MockRegistration{
		OutAppEUI:    lorawan.EUI64([8]byte{1, 1, 1, 1, 1, 1, 1, 1}),
		OutDevEUI:    lorawan.EUI64([8]byte{2, 2, 2, 2, 2, 2, 2, 2}),
		OutNwkSKey:   lorawan.AES128Key([16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6}),
		OutAppSKey:   lorawan.AES128Key([16]byte{6, 5, 4, 3, 2, 1, 0, 9, 8, 7, 6, 5, 4, 3, 2, 1}),
		OutRecipient: NewMockRecipient(),
	}
}

func (r *MockRegistration) Recipient() Recipient {
	return r.OutRecipient
}

func (r *MockRegistration) RawRecipient() []byte {
	data, _ := r.Recipient().MarshalBinary()
	return data
}

func (r *MockRegistration) AppEUI() lorawan.EUI64 {
	return r.OutAppEUI
}

func (r *MockRegistration) DevEUI() lorawan.EUI64 {
	return r.OutDevEUI
}

func (r *MockRegistration) NwkSKey() lorawan.AES128Key {
	return r.OutNwkSKey
}

func (r *MockRegistration) AppSKey() lorawan.AES128Key {
	return r.OutAppSKey
}

//func (r *MockRegistration) DevAddr() lorawan.DevAddr {
//	devAddr := lorawan.DevAddr{}
//	copy(devAddr[:], r.devEUI[4:])
//	return devAddr
//}

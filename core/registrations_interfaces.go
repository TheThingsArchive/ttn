// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package core

import (
	"encoding"
	"encoding/json"

	"github.com/brocaar/lorawan"
)

// Recipient represents the recipient manipulated by adapters
type Recipient interface {
	encoding.BinaryMarshaler
}

// JSONRecipient extends the actual Recipient to also support json marshalin/unmarshaling
type JSONRecipient interface {
	Recipient
	json.Marshaler
}

// Registration gives an elementary base for each other registration levels
type Registration interface {
	Recipient() Recipient
}

// RRegistration represents the first-level of registration, used by router and router adapters
type RRegistration interface {
	Registration
	DevEUI() lorawan.EUI64
}

// BRegistration represents the second-level of registrations, used by the broker and broker
// adapters
type BRegistration interface {
	RRegistration
	AppEUI() lorawan.EUI64
	NwkSKey() lorawan.AES128Key
}

// ARegistration represents another second-level of registrations, used betwen handler and broker to
// register application before OTAA
type ARegistration interface {
	Registration
	AppEUI() lorawan.EUI64
}

// HRegistration represents the third-level of registrations, used bt the handler and handler
// adapters
type HRegistration interface {
	BRegistration
	AppSKey() lorawan.AES128Key
}

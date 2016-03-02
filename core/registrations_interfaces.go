// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package core

import (
	"encoding"

	"github.com/brocaar/lorawan"
)

// Recipient represents the recipient manipulated by adapters
type Recipient interface {
	encoding.BinaryMarshaler
}

// Registration represents the first-level of registration, used by router and router adapters
type Registration interface {
	Recipient() Recipient
	DevEUI() lorawan.EUI64
}

// BRegistration represents the second-level of registrations, used by the broker and broker
// adapters
type BRegistration interface {
	Registration
	AppEUI() lorawan.EUI64
	NwkSKey() lorawan.AES128Key
}

// HRegistration represents the third-level of registrations, used bt the handler and handler
// adapters
type HRegistration interface {
	BRegistration
	AppSKey() lorawan.AES128Key
}

// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package core

import (
	"encoding"

	"github.com/brocaar/lorawan"
)

type Recipient interface {
	encoding.BinaryMarshaler
}

type Registration interface {
	Recipient() Recipient
	DevEUI() lorawan.EUI64
}

type BRegistration interface {
	Registration
	AppEUI() lorawan.EUI64
	NwkSKey() lorawan.AES128Key
}

type HRegistration interface {
	BRegistration
	AppSKey() lorawan.AES128Key
}

// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package refactor

import (
	"github.com/brocaar/lorawan"
)

type BRegistration interface {
	Registration
	AppEUI() lorawan.EUI64
	NwkSKey() lorawan.AES128Key
}

type HRegistration interface {
	BRegistration
	AppSKey() lorawan.AES128Key
}

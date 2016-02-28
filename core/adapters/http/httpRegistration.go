// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package http

import (
	"github.com/TheThingsNetwork/ttn/core"
	"github.com/brocaar/lorawan"
)

type httpRegistration struct {
	recipient core.Recipient
	devEUI    lorawan.EUI64
}

// Recipient implements the core.Registration inteface
func (r httpRegistration) Recipient() core.Recipient {
	return r.recipient
}

// DevEUI implements the core.RRegistration interface
func (r httpRegistration) DevEUI() lorawan.EUI64 {
	return r.devEUI
}

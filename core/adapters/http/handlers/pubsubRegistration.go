// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package handlers

import (
	"github.com/TheThingsNetwork/ttn/core"
	. "github.com/TheThingsNetwork/ttn/core/adapters/http"
	"github.com/brocaar/lorawan"
)

// type pubSubRegistration implements the core.Registration interface
type pubSubRegistration struct {
	recipient Recipient
	appEUI    lorawan.EUI64
	nwkSKey   lorawan.AES128Key
	devEUI    lorawan.EUI64
}

// Recipient implements the core.Registration interface
func (r pubSubRegistration) Recipient() core.Recipient {
	return r.recipient
}

// AppEUI implements the core.Registration interface
func (r pubSubRegistration) AppEUI() lorawan.EUI64 {
	return r.appEUI
}

// DevEUI implements the core.Registration interface
func (r pubSubRegistration) DevEUI() lorawan.EUI64 {
	return r.devEUI
}

// NwkSKey implements the core.Registration interface
func (r pubSubRegistration) NwkSKey() lorawan.AES128Key {
	return r.nwkSKey
}

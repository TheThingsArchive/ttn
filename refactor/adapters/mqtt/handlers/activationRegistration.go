// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package handlers

import (
	core "github.com/TheThingsNetwork/ttn/refactor"
	"github.com/brocaar/lorawan"
)

type activationRegistration struct {
	recipient core.Recipient
	devEUI    lorawan.EUI64
	appEUI    lorawan.EUI64
	nwkSKey   lorawan.AES128Key
	appSKey   lorawan.AES128Key
}

// Recipient implements the core.HRegistration interface
func (r activationRegistration) Recipient() core.Recipient {
	return r.recipient
}

// AppEUI implements the core.HRegistration interface
func (r activationRegistration) AppEUI() lorawan.EUI64 {
	return r.appEUI
}

// DevEUI implements the core.HRegistration interface
func (r activationRegistration) DevEUI() lorawan.EUI64 {
	return r.devEUI
}

// AppSKey implements the core.HRegistration interface
func (r activationRegistration) AppSKey() lorawan.AES128Key {
	return r.appSKey
}

// NwkSKey implements the core.HRegistration interface
func (r activationRegistration) NwkSKey() lorawan.AES128Key {
	return r.nwkSKey
}

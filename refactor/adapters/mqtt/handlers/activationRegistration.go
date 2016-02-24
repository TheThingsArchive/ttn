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

// Recipient implements the core.Registration interface
func (r activationRegistration) Recipient() core.Recipient {
	return r.recipient
}

// AppEUI implements the core.Registration interface
func (r activationRegistration) AppEUI() (lorawan.EUI64, error) {
	return r.appEUI, nil
}

// DevEUI implements the core.Registration interface
func (r activationRegistration) DevEUI() (lorawan.EUI64, error) {
	return r.devEUI, nil
}

// AppSKey implements the core.Registration interface
func (r activationRegistration) AppSKey() (lorawan.AES128Key, error) {
	return r.appSKey, nil
}

// NwkSKey implements the core.Registration interface
func (r activationRegistration) NwkSKey() (lorawan.AES128Key, error) {
	return r.nwkSKey, nil
}

// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package http

import (
	core "github.com/TheThingsNetwork/ttn/refactor"
	. "github.com/TheThingsNetwork/ttn/refactor/errors"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	"github.com/brocaar/lorawan"
)

type httpRegistration struct {
	recipient core.Recipient
	devEUI    *lorawan.EUI64
}

// Recipient implements the core.Registration inteface
func (r httpRegistration) Recipient() core.Recipient {
	return r.recipient
}

// AppEUI implements the core.Registration interface
func (r httpRegistration) AppEUI() (lorawan.EUI64, error) {
	return lorawan.EUI64{}, errors.New(ErrNotSupported, "AppEUI not supported on http registration")
}

// DevEUI implements the core.Registration interface
func (r httpRegistration) DevEUI() (lorawan.EUI64, error) {
	if r.devEUI == nil {
		return lorawan.EUI64{}, errors.New(ErrInvalidStructure, "DevEUI not accessible on this registration")
	}
	return *r.devEUI, nil
}

// AppSKey implements the core.Registration interface
func (r httpRegistration) AppSKey() (lorawan.AES128Key, error) {
	return lorawan.AES128Key{}, errors.New(ErrNotSupported, "AppSKey not supported on http registration")
}

// NwkSKey implements the core.Registration interface
func (r httpRegistration) NwkSKey() (lorawan.AES128Key, error) {
	return lorawan.AES128Key{}, errors.New(ErrNotSupported, "NextRegistration not supported on udp adapter")
}

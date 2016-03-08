// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package handler

import (
	"encoding/json"

	"github.com/TheThingsNetwork/ttn/core"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	"github.com/brocaar/lorawan"
)

// type brokerRegistration implements the core.BRegistration interface
type brokerRegistration struct {
	recipient core.Recipient
	appEUI    lorawan.EUI64
	nwkSKey   lorawan.AES128Key
	devEUI    lorawan.EUI64
}

// Recipient implements the core.BRegistration interface
func (r brokerRegistration) Recipient() core.Recipient {
	return r.recipient
}

// AppEUI implements the core.BRegistration interface
func (r brokerRegistration) AppEUI() lorawan.EUI64 {
	return r.appEUI
}

// DevEUI implements the core.BRegistration interface
func (r brokerRegistration) DevEUI() lorawan.EUI64 {
	return r.devEUI
}

// NwkSKey implements the core.BRegistration interface
func (r brokerRegistration) NwkSKey() lorawan.AES128Key {
	return r.nwkSKey
}

// MarshalJSON implements the encoding/json.Marshaler interface
func (r brokerRegistration) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(struct {
		AppEUI  lorawan.EUI64     `json:"app_eui"`
		NwkSKey lorawan.AES128Key `json:"nwks_key"`
		DevEUI  lorawan.EUI64     `json:"dev_eui"`
	}{
		AppEUI:  r.appEUI,
		NwkSKey: r.nwkSKey,
		DevEUI:  r.devEUI,
	})
	if err != nil {
		return nil, errors.New(errors.Structural, err)
	}
	return data, nil
}

// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package handlers

import (
	"testing"

	"github.com/TheThingsNetwork/ttn/core/adapters/http"
	mocks "github.com/TheThingsNetwork/ttn/core/mocks"
	"github.com/brocaar/lorawan"
)

func TestPubSubRegistration(t *testing.T) {
	recipient := http.NewRecipient("url", "method")
	devEUI := lorawan.EUI64([8]byte{1, 2, 3, 4, 5, 6, 7, 8})
	appEUI := lorawan.EUI64([8]byte{1, 43, 3, 4, 6, 6, 6, 8})
	nwkSKey := lorawan.AES128Key([16]byte{0, 0, 1, 1, 0, 0, 2, 2, 3, 3, 3, 3, 4, 4, 4, 4})

	reg := pubSubRegistration{
		recipient: recipient,
		devEUI:    devEUI,
		appEUI:    appEUI,
		nwkSKey:   nwkSKey,
	}

	mocks.Check(t, recipient, reg.Recipient(), "Recipients")
	mocks.Check(t, devEUI, reg.DevEUI(), "DevEUIs")
	mocks.Check(t, appEUI, reg.AppEUI(), "AppEUIs")
	mocks.Check(t, nwkSKey, reg.NwkSKey(), "NwkSKeys")
}

func TestApplicationsRegistration(t *testing.T) {
	recipient := http.NewRecipient("url", "method")
	appEUI := lorawan.EUI64([8]byte{1, 43, 3, 4, 6, 6, 6, 8})

	reg := applicationsRegistration{
		recipient: recipient,
		appEUI:    appEUI,
	}

	mocks.Check(t, recipient, reg.Recipient(), "Recipients")
	mocks.Check(t, appEUI, reg.AppEUI(), "AppEUIs")
}

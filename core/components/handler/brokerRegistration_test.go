// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package handler

import (
	"testing"

	mocks "github.com/TheThingsNetwork/ttn/core/mocks"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	errutil "github.com/TheThingsNetwork/ttn/utils/errors/checks"
	"github.com/TheThingsNetwork/ttn/utils/pointer"
	"github.com/brocaar/lorawan"
)

func TestRegistration(t *testing.T) {
	recipient := mocks.NewMockJSONRecipient()
	devEUI := lorawan.EUI64([8]byte{1, 2, 3, 4, 5, 6, 7, 8})
	appEUI := lorawan.EUI64([8]byte{1, 43, 3, 4, 6, 6, 6, 8})
	nwkSKey := lorawan.AES128Key([16]byte{0, 0, 1, 1, 0, 0, 2, 2, 3, 3, 3, 3, 4, 4, 4, 4})

	reg := brokerRegistration{
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

func TestRegistrationMarshalUnmarshal(t *testing.T) {
	{
		reg := brokerRegistration{
			recipient: mocks.NewMockJSONRecipient(),
			devEUI:    lorawan.EUI64([8]byte{1, 2, 3, 4, 5, 6, 7, 8}),
			appEUI:    lorawan.EUI64([8]byte{1, 43, 3, 4, 6, 6, 6, 8}),
			nwkSKey:   lorawan.AES128Key([16]byte{0, 0, 1, 1, 0, 0, 2, 2, 3, 3, 3, 3, 4, 4, 4, 4}),
		}

		_, err := reg.MarshalJSON()
		errutil.CheckErrors(t, nil, err)
	}

	{
		reg := brokerRegistration{
			recipient: mocks.NewMockJSONRecipient(),
			devEUI:    lorawan.EUI64([8]byte{1, 2, 3, 4, 5, 6, 7, 8}),
			appEUI:    lorawan.EUI64([8]byte{1, 43, 3, 4, 6, 6, 6, 8}),
			nwkSKey:   lorawan.AES128Key([16]byte{0, 0, 1, 1, 0, 0, 2, 2, 3, 3, 3, 3, 4, 4, 4, 4}),
		}
		reg.recipient.(*mocks.MockJSONRecipient).OutMarshalJSON = []byte("InvalidJSON")

		_, err := reg.MarshalJSON()
		errutil.CheckErrors(t, pointer.String(string(errors.Structural)), err)
	}
}

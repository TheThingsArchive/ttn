// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package broker

import (
	"github.com/TheThingsNetwork/ttn/core"
	"github.com/brocaar/lorawan"
)

type mockController struct {
	Failures         map[string]error
	InLookupDevices  lorawan.EUI64
	InLookupApp      lorawan.EUI64
	InStoreDevices   core.BRegistration
	InStoreApp       core.ARegistration
	InUpdateAppEUI   lorawan.EUI64
	InUpdateDevEUI   lorawan.EUI64
	InUpdateFCnt     uint32
	InUpdateDir      string
	OutLookupDevices []devEntry
	OutLookupApp     appEntry
}

func newMockController() *mockController {
	return &mockController{
		Failures: make(map[string]error),
		OutLookupApp: appEntry{
			Recipient: []byte("MockControllerRecipient"),
			AppEUI:    lorawan.EUI64([8]byte{5, 5, 5, 5, 5, 5, 6, 6}),
		},
		OutLookupDevices: []devEntry{
			{
				Recipient: []byte("MockControllerRecipient"),
				AppEUI:    lorawan.EUI64([8]byte{1, 1, 2, 2, 3, 3, 1, 1}),
				DevEUI:    lorawan.EUI64([8]byte{2, 3, 4, 5, 6, 7, 5, 4}),
				NwkSKey:   lorawan.AES128Key([16]byte{1, 1, 1, 1, 1, 1, 1, 2, 2, 2, 3, 4, 4, 5, 5}),
			},
		},
	}
}

func (s *mockController) LookupDevices(devEUI lorawan.EUI64) ([]devEntry, error) {
	s.InLookupDevices = devEUI
	if s.Failures["LookupDevices"] != nil {
		return nil, s.Failures["LookupDevices"]
	}
	return s.OutLookupDevices, nil
}

func (s *mockController) LookupApplication(appEUI lorawan.EUI64) (appEntry, error) {
	s.InLookupApp = appEUI
	if s.Failures["LookupApplication"] != nil {
		return appEntry{}, s.Failures["LookupApplication"]
	}
	return s.OutLookupApp, nil
}

func (s *mockController) StoreDevice(reg core.BRegistration) error {
	s.InStoreDevices = reg
	return s.Failures["StoreDevice"]
}

func (s *mockController) StoreApplication(reg core.ARegistration) error {
	s.InStoreApp = reg
	return s.Failures["StoreApplication"]
}

func (s *mockController) UpdateFCnt(appEUI lorawan.EUI64, devEUI lorawan.EUI64, fcnt uint32, dir string) error {
	s.InUpdateAppEUI = appEUI
	s.InUpdateDevEUI = devEUI
	s.InUpdateFCnt = fcnt
	s.InUpdateDir = dir
	return s.Failures["UpdateFCnt"]
}

func (s *mockController) Close() error {
	return s.Failures["Close"]
}

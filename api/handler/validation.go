// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package handler

import (
	"github.com/TheThingsNetwork/ttn/api"
	"github.com/TheThingsNetwork/ttn/utils/errors"
)

// Validate implements the api.Validator interface
func (m *DeviceActivationResponse) Validate() error {
	if err := api.NotNilAndValid(m.DownlinkOption, "DownlinkOption"); err != nil {
		return err
	}
	if err := api.NotNilAndValid(m.ActivationMetadata, "ActivationMetadata"); err != nil {
		return err
	}
	if m.Message != nil {
		if err := m.Message.Validate(); err != nil {
			return errors.NewErrInvalidArgument("Message", err.Error())
		}
	}
	return nil
}

// Validate implements the api.Validator interface
func (m *ApplicationIdentifier) Validate() error {
	if err := api.NotEmptyAndValidID(m.AppId, "AppId"); err != nil {
		return err
	}
	return nil
}

// Validate implements the api.Validator interface
func (m *Application) Validate() error {
	if err := api.NotEmptyAndValidID(m.AppId, "AppId"); err != nil {
		return err
	}
	return nil
}

// Validate implements the api.Validator interface
func (m *DeviceIdentifier) Validate() error {
	if err := api.NotEmptyAndValidID(m.AppId, "AppId"); err != nil {
		return err
	}
	if err := api.NotEmptyAndValidID(m.DevId, "DevId"); err != nil {
		return err
	}
	return nil
}

// Validate implements the api.Validator interface
func (m *Device) Validate() error {
	if err := api.NotEmptyAndValidID(m.AppId, "AppId"); err != nil {
		return err
	}
	if err := api.NotEmptyAndValidID(m.DevId, "DevId"); err != nil {
		return err
	}
	if err := api.NotNilAndValid(m.Device, "Device"); err != nil {
		return err
	}
	return nil
}

// Validate implements the api.Validator interface
func (m *Device_LorawanDevice) Validate() error {
	if err := api.NotNilAndValid(m.LorawanDevice, "LorawanDevice"); err != nil {
		return err
	}
	return nil
}

// Validate implements the api.Validator interface
func (m *SimulatedUplinkMessage) Validate() error {
	if err := api.NotEmptyAndValidID(m.AppId, "AppId"); err != nil {
		return err
	}
	if err := api.NotEmptyAndValidID(m.DevId, "DevId"); err != nil {
		return err
	}
	return nil
}

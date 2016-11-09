package handler

import "github.com/TheThingsNetwork/ttn/api"

// Validate implements the api.Validator interface
func (m *DeviceActivationResponse) Validate() error {
	if err := api.NotNilAndValid(m.DownlinkOption, "DownlinkOption"); err != nil {
		return err
	}
	if err := api.NotNilAndValid(m.ActivationMetadata, "ActivationMetadata"); err != nil {
		return err
	}
	if err := api.NotNilAndValid(m.Message, "Message"); err != nil {
		return err
	}
	return nil
}

// Validate implements the api.Validator interface
func (m *ApplicationIdentifier) Validate() error {
	if err := api.NotEmptyAndValidId(m.AppId, "AppId"); err != nil {
		return err
	}
	return nil
}

// Validate implements the api.Validator interface
func (m *Application) Validate() error {
	if err := api.NotEmptyAndValidId(m.AppId, "AppId"); err != nil {
		return err
	}
	return nil
}

// Validate implements the api.Validator interface
func (m *DeviceIdentifier) Validate() error {
	if err := api.NotEmptyAndValidId(m.AppId, "AppId"); err != nil {
		return err
	}
	if err := api.NotEmptyAndValidId(m.DevId, "DevId"); err != nil {
		return err
	}
	return nil
}

// Validate implements the api.Validator interface
func (m *Device) Validate() error {
	if err := api.NotEmptyAndValidId(m.AppId, "AppId"); err != nil {
		return err
	}
	if err := api.NotEmptyAndValidId(m.DevId, "DevId"); err != nil {
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

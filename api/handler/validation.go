package handler

import "github.com/TheThingsNetwork/ttn/api"

// Validate implements the api.Validator interface
func (m *DeviceActivationResponse) Validate() bool {
	if m.DownlinkOption == nil || !m.DownlinkOption.Validate() {
		return false
	}
	if m.ActivationMetadata == nil || !m.ActivationMetadata.Validate() {
		return false
	}
	if m.Message != nil && !m.Message.Validate() {
		return false
	}
	return true
}

// Validate implements the api.Validator interface
func (m *ApplicationIdentifier) Validate() bool {
	if m.AppId == "" || !api.ValidID(m.AppId) {
		return false
	}
	return true
}

// Validate implements the api.Validator interface
func (m *Application) Validate() bool {
	if m.AppId == "" || !api.ValidID(m.AppId) {
		return false
	}
	return true
}

// Validate implements the api.Validator interface
func (m *DeviceIdentifier) Validate() bool {
	if m.AppId == "" || !api.ValidID(m.AppId) {
		return false
	}
	if m.DevId == "" || !api.ValidID(m.DevId) {
		return false
	}
	return true
}

// Validate implements the api.Validator interface
func (m *Device) Validate() bool {
	if m.AppId == "" || !api.ValidID(m.AppId) {
		return false
	}
	if m.DevId == "" || !api.ValidID(m.DevId) {
		return false
	}
	if m.Device == nil || !api.Validate(m.Device) {
		return false
	}
	return true
}

// Validate implements the api.Validator interface
func (m *Device_LorawanDevice) Validate() bool {
	return m.LorawanDevice.Validate()
}

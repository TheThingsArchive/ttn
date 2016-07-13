package handler

// Validate implements the api.Validator interface
func (m *DeviceActivationResponse) Validate() bool {
	if m.AppId == "" {
		return false
	}
	if m.DownlinkOption == nil || !m.DownlinkOption.Validate() {
		return false
	}
	if m.ActivationMetadata == nil || !m.ActivationMetadata.Validate() {
		return false
	}
	return true
}

// Validate implements the api.Validator interface
func (m *ApplicationIdentifier) Validate() bool {
	if m.AppId == "" {
		return false
	}
	return true
}

// Validate implements the api.Validator interface
func (m *Application) Validate() bool {
	if m.AppId == "" {
		return false
	}
	return true
}

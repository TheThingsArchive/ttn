package broker

// Validate implements the api.Validator interface
func (m *DownlinkOption) Validate() bool {
	if m.Identifier == "" {
		return false
	}
	if m.GatewayEui == nil || m.GatewayEui.IsEmpty() {
		return false
	}
	if m.ProtocolConfig == nil || !m.ProtocolConfig.Validate() {
		return false
	}
	if m.GatewayConfig == nil || !m.GatewayConfig.Validate() {
		return false
	}
	return true
}

// Validate implements the api.Validator interface
func (m *UplinkMessage) Validate() bool {
	if m.DevEui == nil || m.DevEui.IsEmpty() {
		return false
	}
	if m.AppEui == nil || m.AppEui.IsEmpty() {
		return false
	}
	if m.ProtocolMetadata == nil || !m.ProtocolMetadata.Validate() {
		return false
	}
	if m.GatewayMetadata == nil || !m.GatewayMetadata.Validate() {
		return false
	}
	return true
}

// Validate implements the api.Validator interface
func (m *DownlinkMessage) Validate() bool {
	if m.DevEui == nil || m.DevEui.IsEmpty() {
		return false
	}
	if m.AppEui == nil || m.AppEui.IsEmpty() {
		return false
	}
	if m.DownlinkOption == nil || !m.DownlinkOption.Validate() {
		return false
	}
	return true
}

// Validate implements the api.Validator interface
func (m *DeduplicatedUplinkMessage) Validate() bool {
	if m.DevEui == nil || m.DevEui.IsEmpty() {
		return false
	}
	if m.AppEui == nil || m.AppEui.IsEmpty() {
		return false
	}
	if m.AppId == "" {
		return false
	}
	if m.ProtocolMetadata == nil || !m.ProtocolMetadata.Validate() {
		return false
	}
	return true
}

// Validate implements the api.Validator interface
func (m *DeviceActivationRequest) Validate() bool {
	if m.DevEui == nil || m.DevEui.IsEmpty() {
		return false
	}
	if m.AppEui == nil || m.AppEui.IsEmpty() {
		return false
	}
	if m.ProtocolMetadata == nil || !m.ProtocolMetadata.Validate() {
		return false
	}
	if m.GatewayMetadata == nil || !m.GatewayMetadata.Validate() {
		return false
	}
	if m.ActivationMetadata == nil || !m.ActivationMetadata.Validate() {
		return false
	}
	return true
}

// Validate implements the api.Validator interface
func (m *DeduplicatedDeviceActivationRequest) Validate() bool {
	if m.DevEui == nil || m.DevEui.IsEmpty() {
		return false
	}
	if m.AppEui == nil || m.AppEui.IsEmpty() {
		return false
	}
	if m.ProtocolMetadata == nil || !m.ProtocolMetadata.Validate() {
		return false
	}
	return true
}

// Validate implements the api.Validator interface
func (m *ApplicationHandlerRegistration) Validate() bool {
	if m.AppId == "" {
		return false
	}
	if m.HandlerId == "" {
		return false
	}
	return true
}

package router

// Validate implements the api.Validator interface
func (m *UplinkMessage) Validate() bool {
	if m.GatewayMetadata == nil || !m.GatewayMetadata.Validate() {
		return false
	}
	if m.ProtocolMetadata == nil || !m.ProtocolMetadata.Validate() {
		return false
	}
	return true
}

// Validate implements the api.Validator interface
func (m *DownlinkMessage) Validate() bool {
	if m.ProtocolConfiguration == nil || !m.ProtocolConfiguration.Validate() {
		return false
	}
	if m.GatewayConfiguration == nil || !m.GatewayConfiguration.Validate() {
		return false
	}
	return true
}

// Validate implements the api.Validator interface
func (m *DeviceActivationRequest) Validate() bool {
	if m.AppEui == nil || m.AppEui.IsEmpty() {
		return false
	}
	if m.DevEui == nil || m.DevEui.IsEmpty() {
		return false
	}
	if m.GatewayMetadata == nil || !m.GatewayMetadata.Validate() {
		return false
	}
	if m.ProtocolMetadata == nil || !m.ProtocolMetadata.Validate() {
		return false
	}
	return true
}

package gateway

// Validate implements the api.Validator interface
func (m *RxMetadata) Validate() bool {
	if m.GatewayEui == nil || m.GatewayEui.IsEmpty() {
		return false
	}
	return true
}

// Validate implements the api.Validator interface
func (m *TxConfiguration) Validate() bool {
	return true
}

// Validate implements the api.Validator interface
func (m *Status) Validate() bool {
	return true
}

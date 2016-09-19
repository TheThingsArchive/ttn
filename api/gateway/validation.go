package gateway

// Validate implements the api.Validator interface
func (m *RxMetadata) Validate() bool {
	if m.GatewayId == "" {
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

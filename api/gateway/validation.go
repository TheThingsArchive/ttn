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

// Validate implements the api.Validator interface
func (m *GPSMetadata) Validate() bool {
	return m != nil && !m.IsZero()
}

func (m GPSMetadata) IsZero() bool {
	return m.Latitude == 0 && m.Longitude == 0
}

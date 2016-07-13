package protocol

// Validate implements the api.Validator interface
func (m *RxMetadata) Validate() bool {
	switch {
	case m.GetLorawan() != nil:
		if !m.GetLorawan().Validate() {
			return false
		}
	}
	return true
}

// Validate implements the api.Validator interface
func (m *TxConfiguration) Validate() bool {
	switch {
	case m.GetLorawan() != nil:
		if !m.GetLorawan().Validate() {
			return false
		}
	}
	return true
}

// Validate implements the api.Validator interface
func (m *ActivationMetadata) Validate() bool {
	switch {
	case m.GetLorawan() != nil:
		if !m.GetLorawan().Validate() {
			return false
		}
	}
	return true
}

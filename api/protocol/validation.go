package protocol

import "github.com/TheThingsNetwork/ttn/api"

// Validate implements the api.Validator interface
func (m *RxMetadata) Validate() bool {
	if m.Protocol == nil || !api.Validate(m.Protocol) {
		return false
	}
	return true
}

// Validate implements the api.Validator interface
func (m *TxConfiguration) Validate() bool {
	if m.Protocol == nil || !api.Validate(m.Protocol) {
		return false
	}
	return true
}

// Validate implements the api.Validator interface
func (m *ActivationMetadata) Validate() bool {
	if m.Protocol == nil || !api.Validate(m.Protocol) {
		return false
	}
	return true
}

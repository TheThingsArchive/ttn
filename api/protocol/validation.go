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
func (m *RxMetadata_Lorawan) Validate() bool {
	return m.Lorawan.Validate()
}

// Validate implements the api.Validator interface
func (m *TxConfiguration) Validate() bool {
	if m.Protocol == nil || !api.Validate(m.Protocol) {
		return false
	}
	return true
}

// Validate implements the api.Validator interface
func (m *TxConfiguration_Lorawan) Validate() bool {
	return m.Lorawan.Validate()
}

// Validate implements the api.Validator interface
func (m *ActivationMetadata) Validate() bool {
	if m.Protocol == nil || !api.Validate(m.Protocol) {
		return false
	}
	return true
}

// Validate implements the api.Validator interface
func (m *ActivationMetadata_Lorawan) Validate() bool {
	return m.Lorawan.Validate()
}

// Validate implements the api.Validator interface
func (m *Message) Validate() bool {
	if m.Protocol == nil || !api.Validate(m.Protocol) {
		return false
	}
	return true
}

// Validate implements the api.Validator interface
func (m *Message_Lorawan) Validate() bool {
	return m.Lorawan.Validate()
}

package protocol

import (
	"github.com/TheThingsNetwork/ttn/api"
	"github.com/TheThingsNetwork/ttn/utils/errors"
)

// Validate implements the api.Validator interface
func (m *RxMetadata) Validate() error {
	if m.Protocol == nil  {
		return errors.NewErrInvalidArgument("Protocol", "can not be empty")
	}
	if err := api.Validate(m.Protocol); err != nil {
		return errors.NewErrInvalidArgument("Protocol", err.Error())
	}
	return nil
}

// Validate implements the api.Validator interface
func (m *RxMetadata_Lorawan) Validate() error {
	if err := m.Lorawan.Validate(); err != nil {
		return errors.NewErrInvalidArgument("Lorawan", err.Error())
	}
	return nil
}

// Validate implements the api.Validator interface
func (m *TxConfiguration) Validate() error {
	if m.Protocol == nil {
		return errors.New("RxMetadata.Protocol is nil")
	}
	return api.Validate(m.Protocol)
}

// Validate implements the api.Validator interface
func (m *TxConfiguration_Lorawan) Validate() error {
	return m.Lorawan.Validate()
}

// Validate implements the api.Validator interface
func (m *ActivationMetadata) Validate() error {
	if m.Protocol == nil {
		return errors.NewErrInvalidArgument("Protocol", "can not be empty")
	}
	if err := api.Validate(m.Protocol); err != nil {
		return errors.NewErrInvalidArgument("Protocol", err.Error())
	}

	return nil
}

// Validate implements the api.Validator interface
func (m *ActivationMetadata_Lorawan) Validate() error {
	return m.Lorawan.Validate()
}

// Validate implements the api.Validator interface
func (m *Message) Validate() error {
	if m.Protocol == nil {
		return errors.NewErrInvalidArgument("Protocol", "can not be empty")
	}
	if err := api.Validate(m.Protocol); err != nil {
		return errors.NewErrInvalidArgument("Protocol", err.Error())
	}

	return nil
}

// Validate implements the api.Validator interface
func (m *Message_Lorawan) Validate() error {
	return m.Lorawan.Validate()
}

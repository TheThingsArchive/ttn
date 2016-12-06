package broker

import (
	"github.com/TheThingsNetwork/ttn/api"
	"github.com/TheThingsNetwork/ttn/utils/errors"
)

// Validate implements the api.Validator interface
func (m *DownlinkOption) Validate() error {
	if m.Identifier == "" {
		return errors.NewErrInvalidArgument("Identifier", "can not be empty")
	}
	if m.GatewayId == "" {
		return errors.NewErrInvalidArgument("GatewayId", "can not be empty")
	}
	if err := api.NotNilAndValid(m.ProtocolConfig, "ProtocolConfig"); err != nil {
		return err
	}
	if err := api.NotNilAndValid(m.GatewayConfig, "GatewayConfig"); err != nil {
		return err
	}
	return nil
}

// Validate implements the api.Validator interface
func (m *UplinkMessage) Validate() error {
	if err := api.NotNilAndValid(m.ProtocolMetadata, "ProtocolMetadata"); err != nil {
		return err
	}
	if err := api.NotNilAndValid(m.GatewayMetadata, "GatewayMetadata"); err != nil {
		return err
	}
	if m.Message != nil {
		if err := m.Message.Validate(); err != nil {
			return errors.NewErrInvalidArgument("Message", err.Error())
		}
	}
	return nil
}

// Validate implements the api.Validator interface
func (m *DownlinkMessage) Validate() error {
	if err := api.NotEmptyAndValidID(m.DevId, "DevId"); err != nil {
		return err
	}
	if err := api.NotEmptyAndValidID(m.AppId, "AppId"); err != nil {
		return err
	}

	if err := api.NotNilAndValid(m.DownlinkOption, "DownlinkOption"); err != nil {
		return err
	}
	if m.Message != nil {
		if err := m.Message.Validate(); err != nil {
			return errors.NewErrInvalidArgument("Message", err.Error())
		}
	}
	return nil
}

// Validate implements the api.Validator interface
func (m *DeduplicatedUplinkMessage) Validate() error {
	if err := api.NotEmptyAndValidID(m.AppId, "AppId"); err != nil {
		return err
	}
	if err := api.NotEmptyAndValidID(m.DevId, "DevId"); err != nil {
		return err
	}
	if err := api.NotNilAndValid(m.ProtocolMetadata, "ProtocolMetadata"); err != nil {
		return err
	}
	if m.ResponseTemplate != nil {
		if err := m.ResponseTemplate.Validate(); err != nil {
			return errors.NewErrInvalidArgument("ResponseTemplate", err.Error())
		}
	}
	if m.Message != nil {
		if err := m.Message.Validate(); err != nil {
			return errors.NewErrInvalidArgument("Message", err.Error())
		}
	}
	return nil
}

// Validate implements the api.Validator interface
func (m *DeviceActivationRequest) Validate() error {
	if err := api.NotNilAndValid(m.ProtocolMetadata, "ProtocolMetadata"); err != nil {
		return err
	}
	if err := api.NotNilAndValid(m.GatewayMetadata, "GatewayMetadata"); err != nil {
		return err
	}
	if err := api.NotNilAndValid(m.ActivationMetadata, "ActivationMetadata"); err != nil {
		return err
	}
	if m.Message != nil {
		if err := m.Message.Validate(); err != nil {
			return errors.NewErrInvalidArgument("Message", err.Error())
		}
	}
	return nil
}

// Validate implements the api.Validator interface
func (m *DeduplicatedDeviceActivationRequest) Validate() error {
	if err := api.NotNilAndValid(m.ProtocolMetadata, "ProtocolMetadata"); err != nil {
		return err
	}
	if m.Message != nil {
		if err := m.Message.Validate(); err != nil {
			return errors.NewErrInvalidArgument("Message", err.Error())
		}
	}
	return nil
}

// Validate implements the api.Validator interface
func (m *ActivationChallengeRequest) Validate() error {
	return nil
}

// Validate implements the api.Validator interface
func (m *ApplicationHandlerRegistration) Validate() error {
	if err := api.NotEmptyAndValidID(m.AppId, "AppId"); err != nil {
		return err
	}
	if m.HandlerId == "" {
		return errors.NewErrInvalidArgument("HandlerId", "can not be empty")
	}
	return nil
}

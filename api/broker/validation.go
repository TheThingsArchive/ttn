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
	return nil
}

// Validate implements the api.Validator interface
func (m *DownlinkMessage) Validate() error {
	if m.DevId == "" {
		return errors.NewErrInvalidArgument("DevId", "can not be empty")
	}
	if api.ValidID(m.DevId) {
		return errors.NewErrInvalidArgument("DevId", "has wrong format " + m.DevId)
	}
	if m.AppId == "" {
		return errors.NewErrInvalidArgument("AppId", "can not be empty")
	}
	if api.ValidID(m.AppId) {
		return errors.NewErrInvalidArgument("AppId", "has wrong format " + m.AppId)
	}
	if err := api.NotNilAndValid(m.DownlinkOption, "DownlinkOption"); err != nil {
		return err
	}
	return nil
}

// Validate implements the api.Validator interface
func (m *DeduplicatedUplinkMessage) Validate() error {
	if err := api.NotEmptyAndValidId(m.AppId, "AppId"); err != nil {
		return err
	}
	if err := api.NotEmptyAndValidId(m.DevId, "DevId"); err != nil {
		return err
	}
	if err := api.NotNilAndValid(m.ProtocolMetadata, "ProtocolMetadata"); err != nil {
		return err
	}
	if err := api.NotNilAndValid(m.ResponseTemplate, "ResponseTemplate"); err != nil {
		return err
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

	return nil
}

// Validate implements the api.Validator interface
func (m *DeduplicatedDeviceActivationRequest) Validate() error {
	if err := api.NotNilAndValid(m.ProtocolMetadata, "ProtocolMetadata"); err != nil {
		return err
	}
	return nil
}

// Validate implements the api.Validator interface
func (m *ActivationChallengeRequest) Validate() error {
	return nil
}

// Validate implements the api.Validator interface
func (m *ApplicationHandlerRegistration) Validate() error {
	if err := api.NotEmptyAndValidId(m.AppId, "AppId"); err != nil {
		return err
	}
	if m.HandlerId == "" {
		return errors.NewErrInvalidArgument("HandlerId", "can not be empty")
	}
	return nil
}

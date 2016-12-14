package router

import (
	"github.com/TheThingsNetwork/ttn/api"
	"github.com/TheThingsNetwork/ttn/utils/errors"
)

// Validate implements the api.Validator interface
func (m *UplinkMessage) Validate() error {
	if err := api.NotNilAndValid(m.GatewayMetadata, "GatewayMetadata"); err != nil {
		return err
	}
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
func (m *DownlinkMessage) Validate() error {
	if err := api.NotNilAndValid(m.ProtocolConfiguration, "ProtocolConfiguration"); err != nil {
		return err
	}
	if err := api.NotNilAndValid(m.GatewayConfiguration, "GatewayConfiguration"); err != nil {
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
func (m *DeviceActivationRequest) Validate() error {
	if err := api.NotNilAndValid(m.GatewayMetadata, "GatewayMetadata"); err != nil {
		return err
	}
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

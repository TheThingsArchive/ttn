package gateway

import "github.com/TheThingsNetwork/ttn/utils/errors"

// Validate implements the api.Validator interface
func (m *RxMetadata) Validate() error {
	if m.GatewayId == "" {
		return errors.NewErrInvalidArgument("GatewayId", "can not be empty")
	}
	return nil
}

// Validate implements the api.Validator interface
func (m *TxConfiguration) Validate() error {
	return nil
}

// Validate implements the api.Validator interface
func (m *Status) Validate() error {
	return nil
}

// Validate implements the api.Validator interface
func (m *GPSMetadata) Validate() error {
	if m == nil || m.IsZero() {
		return errors.NewErrInvalidArgument("GPSMetadata", "can not be empty")
	}
	return nil
}

func (m GPSMetadata) IsZero() bool {
	return m.Latitude == 0 && m.Longitude == 0
}

// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

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

// Location metadata errors
var (
	ErrInvalidLatitude  = errors.NewErrInvalidArgument("LocationMetadata", "invalid latitude")
	ErrInvalidLongitude = errors.NewErrInvalidArgument("LocationMetadata", "invalid longitude")
	ErrLocationZero     = errors.NewErrInvalidArgument("LocationMetadata", "is zero, so should be nil")
)

// Validate implements the api.Validator interface
func (m LocationMetadata) Validate() error {
	if m.IsZero() {
		return ErrLocationZero
	}
	if m.Latitude >= 90-delta || m.Latitude <= -90+delta {
		return ErrInvalidLatitude
	}
	if m.Longitude > 180 || m.Longitude < -180 {
		return ErrInvalidLongitude
	}
	return nil
}

const delta = 0.01

// IsZero returns whether the location is close enough to zero (and should be nil)
func (m LocationMetadata) IsZero() bool {
	return (m.Latitude > -delta && m.Latitude < delta) && (m.Longitude > -delta && m.Longitude < delta)
}

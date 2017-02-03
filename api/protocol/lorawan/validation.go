// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package lorawan

import (
	"github.com/TheThingsNetwork/ttn/api"
	"github.com/TheThingsNetwork/ttn/core/types"
	"github.com/TheThingsNetwork/ttn/utils/errors"
)

// Validate implements the api.Validator interface
func (m *DeviceIdentifier) Validate() error {
	if m.AppEui == nil || m.AppEui.IsEmpty() {
		return errors.NewErrInvalidArgument("AppEui", "can not be empty")
	}
	if m.DevEui == nil || m.DevEui.IsEmpty() {
		return errors.NewErrInvalidArgument("DevEui", "can not be empty")
	}
	return nil
}

// Validate implements the api.Validator interface
func (m *Device) Validate() error {
	if m.AppEui == nil || m.AppEui.IsEmpty() {
		return errors.NewErrInvalidArgument("AppEui", "can not be empty")
	}
	if m.DevEui == nil || m.DevEui.IsEmpty() {
		return errors.NewErrInvalidArgument("DevEui", "can not be empty")
	}
	if err := api.NotEmptyAndValidID(m.AppId, "AppId"); err != nil {
		return err
	}
	if err := api.NotEmptyAndValidID(m.DevId, "DevId"); err != nil {
		return err
	}
	return nil
}

// Validate implements the api.Validator interface
func (m *Metadata) Validate() error {
	switch m.Modulation {
	case Modulation_LORA:
		if m.DataRate == "" {
			return errors.NewErrInvalidArgument("DataRate", "can not be empty")
		}
		if _, err := types.ParseDataRate(m.DataRate); err != nil {
			return errors.NewErrInvalidArgument("DataRate", err.Error())
		}
	case Modulation_FSK:
		if m.BitRate == 0 {
			return errors.NewErrInvalidArgument("BitRate", "can not be empty")
		}
	}
	if m.CodingRate == "" {
		return errors.NewErrInvalidArgument("CodingRate", "can not be empty")
	}
	return nil
}

// Validate implements the api.Validator interface
func (m *TxConfiguration) Validate() error {
	switch m.Modulation {
	case Modulation_LORA:
		if m.DataRate == "" {
			return errors.NewErrInvalidArgument("DataRate", "can not be empty")
		}
		if _, err := types.ParseDataRate(m.DataRate); err != nil {
			return errors.NewErrInvalidArgument("DataRate", err.Error())
		}
	case Modulation_FSK:
		if m.BitRate == 0 {
			return errors.NewErrInvalidArgument("BitRate", "can not be empty")
		}
	}
	if m.CodingRate == "" {
		return errors.NewErrInvalidArgument("CodingRate", "can not be empty")
	}
	return nil
}

// Validate implements the api.Validator interface
func (m *ActivationMetadata) Validate() error {
	if m.AppEui == nil || m.AppEui.IsEmpty() {
		return errors.NewErrInvalidArgument("AppEui", "can not be empty")
	}
	if m.DevEui == nil || m.DevEui.IsEmpty() {
		return errors.NewErrInvalidArgument("DevEui", "can not be empty")
	}
	if m.DevAddr != nil && m.DevAddr.IsEmpty() {
		return errors.NewErrInvalidArgument("DevAddr", "can not be empty")
	}
	if m.NwkSKey != nil && m.NwkSKey.IsEmpty() {
		return errors.NewErrInvalidArgument("NwkSKey", "can not be empty")
	}
	return nil
}

// Validate implements the api.Validator interface
func (m *Message) Validate() error {
	if m.Major != Major_LORAWAN_R1 {
		return errors.NewErrInvalidArgument("Major", "invalid value "+Major_LORAWAN_R1.String())
	}
	switch m.MType {
	case MType_JOIN_REQUEST:
		if m.GetJoinRequestPayload() == nil {
			return errors.NewErrInvalidArgument("JoinRequestPayload", "can not be empty")
		}
		if err := m.GetJoinRequestPayload().Validate(); err != nil {
			return errors.NewErrInvalidArgument("JoinRequestPayload", err.Error())
		}
	case MType_JOIN_ACCEPT:
		if m.GetJoinAcceptPayload() == nil {
			return errors.NewErrInvalidArgument("JoinAcceptPayload", "can not be empty")
		}
		if err := m.GetJoinAcceptPayload().Validate(); err != nil {
			return errors.NewErrInvalidArgument("JoinAcceptPayload", err.Error())
		}
	case MType_UNCONFIRMED_UP, MType_UNCONFIRMED_DOWN, MType_CONFIRMED_UP, MType_CONFIRMED_DOWN:
		if m.GetMacPayload() == nil {
			return errors.NewErrInvalidArgument("MacPayload", "can not be empty")
		}
		if err := m.GetMacPayload().Validate(); err != nil {
			return errors.NewErrInvalidArgument("MacPayload", err.Error())
		}
	default:
		return errors.NewErrInvalidArgument("MType", "unknown type "+m.MType.String())
	}

	return nil
}

// Validate implements the api.Validator interface
func (m *JoinRequestPayload) Validate() error {
	if m.AppEui.IsEmpty() {
		return errors.NewErrInvalidArgument("AppEui", "can not be empty")
	}
	if m.DevEui.IsEmpty() {
		return errors.NewErrInvalidArgument("DevEui", "can not be empty")
	}

	return nil
}

// Validate implements the api.Validator interface
func (m *JoinAcceptPayload) Validate() error {
	if len(m.Encrypted) != 0 {
		return nil
	}

	if m.CfList != nil && len(m.CfList.Freq) != 5 {
		return errors.NewErrInvalidArgument("CfList.Freq", "length must be 5")
	}

	if m.DevAddr.IsEmpty() {
		return errors.NewErrInvalidArgument("DevAddr", "can not be empty")
	}
	if m.NetId.IsEmpty() {
		return errors.NewErrInvalidArgument("NetId", "can not be empty")
	}

	return nil
}

// Validate implements the api.Validator interface
func (m *MACPayload) Validate() error {
	if m.DevAddr.IsEmpty() {
		return errors.NewErrInvalidArgument("DevAddr", "can not be empty")
	}
	return nil
}

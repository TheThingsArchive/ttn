package lorawan

import "github.com/TheThingsNetwork/ttn/api"

// Validate implements the api.Validator interface
func (m *DeviceIdentifier) Validate() bool {
	if m.AppEui == nil || m.AppEui.IsEmpty() {
		return false
	}
	if m.DevEui == nil || m.DevEui.IsEmpty() {
		return false
	}
	return true
}

// Validate implements the api.Validator interface
func (m *Device) Validate() bool {
	if m.AppEui == nil || m.AppEui.IsEmpty() {
		return false
	}
	if m.DevEui == nil || m.DevEui.IsEmpty() {
		return false
	}
	if m.AppId == "" || !api.ValidID(m.AppId) {
		return false
	}
	if m.DevId == "" || !api.ValidID(m.DevId) {
		return false
	}
	return true
}

// Validate implements the api.Validator interface
func (m *Metadata) Validate() bool {
	switch m.Modulation {
	case Modulation_LORA:
		if m.DataRate == "" {
			return false
		}
	case Modulation_FSK:
		if m.BitRate == 0 {
			return false
		}
	}
	if m.CodingRate == "" {
		return false
	}
	return true
}

// Validate implements the api.Validator interface
func (m *TxConfiguration) Validate() bool {
	switch m.Modulation {
	case Modulation_LORA:
		if m.DataRate == "" {
			return false
		}
	case Modulation_FSK:
		if m.BitRate == 0 {
			return false
		}
	}
	if m.CodingRate == "" {
		return false
	}
	return true
}

// Validate implements the api.Validator interface
func (m *ActivationMetadata) Validate() bool {
	if m.AppEui == nil || m.AppEui.IsEmpty() {
		return false
	}
	if m.DevEui == nil || m.DevEui.IsEmpty() {
		return false
	}
	if m.DevAddr != nil && m.DevAddr.IsEmpty() {
		return false
	}
	if m.NwkSKey != nil && m.NwkSKey.IsEmpty() {
		return false
	}
	return true
}

// Validate implements the api.Validator interface
func (m *Message) Validate() bool {
	if m.Major != Major_LORAWAN_R1 {
		return false
	}
	switch m.MType {
	case MType_JOIN_REQUEST:
		return m.GetJoinRequestPayload() != nil && m.GetJoinRequestPayload().Validate()
	case MType_JOIN_ACCEPT:
		return m.GetJoinAcceptPayload() != nil && m.GetJoinAcceptPayload().Validate()
	case MType_UNCONFIRMED_UP, MType_UNCONFIRMED_DOWN, MType_CONFIRMED_UP, MType_CONFIRMED_DOWN:
		return m.GetMacPayload() != nil && m.GetMacPayload().Validate()
	}
	return false
}

// Validate implements the api.Validator interface
func (m *JoinRequestPayload) Validate() bool {
	return len(m.AppEui) == 8 && len(m.DevEui) == 8 && len(m.DevNonce) == 2
}

// Validate implements the api.Validator interface
func (m *JoinAcceptPayload) Validate() bool {
	if len(m.Encrypted) != 0 {
		return true
	}
	if m.CfList != nil && len(m.CfList.Freq) != 5 {
		return false
	}
	return len(m.DevAddr) == 4 && len(m.AppNonce) == 3 && len(m.NetId) == 3
}

// Validate implements the api.Validator interface
func (m *MACPayload) Validate() bool {
	return len(m.DevAddr) == 4
}

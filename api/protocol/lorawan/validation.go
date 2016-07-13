package lorawan

// Validate implements the api.Validator interface
func (m *DeviceIdentifier) Validate() bool {
	if m.AppEui == nil || m.AppEui.IsEmpty() {
		return false
	}
	if m.DevEui == nil || m.DevEui.IsEmpty() {
		return false
	}
	if m.AppId == "" {
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
	if m.AppId == "" {
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
	if m.DevAddr == nil || m.DevAddr.IsEmpty() {
		return false
	}
	if m.NwkSKey == nil || m.NwkSKey.IsEmpty() {
		return false
	}
	return true
}

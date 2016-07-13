package networkserver

// Validate implements the api.Validator interface
func (m *DevicesRequest) Validate() bool {
	if m.DevAddr == nil || m.DevAddr.IsEmpty() {
		return false
	}
	return true
}

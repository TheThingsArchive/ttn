package discovery

import "github.com/TheThingsNetwork/ttn/api"

// Validate implements the api.Validator interface
func (m *Announcement) Validate() bool {
	if m.Id == "" || !api.ValidID(m.Id) {
		return false
	}
	switch m.ServiceName {
	case "router", "broker", "handler":
	default:
		return false
	}
	return true
}

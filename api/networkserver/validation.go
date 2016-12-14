package networkserver

import "github.com/TheThingsNetwork/ttn/utils/errors"

// Validate implements the api.Validator interface
func (m *DevicesRequest) Validate() error {
	if m.DevAddr == nil || m.DevAddr.IsEmpty() {
		return errors.NewErrInvalidArgument("DevAddr", "can not be empty")
	}
	return nil
}

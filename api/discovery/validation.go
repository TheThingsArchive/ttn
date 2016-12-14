package discovery

import (
	"github.com/TheThingsNetwork/ttn/api"
	"github.com/TheThingsNetwork/ttn/utils/errors"
)

// Validate implements the api.Validator interface
func (m *Announcement) Validate() error {
	if err := api.NotEmptyAndValidID(m.Id, "Id"); err != nil {
		return err
	}
	switch m.ServiceName {
	case "router", "broker", "handler":
	default:
		return errors.NewErrInvalidArgument("ServiceName", "expected one of router, broker, handler but was "+m.ServiceName)
	}
	return nil
}

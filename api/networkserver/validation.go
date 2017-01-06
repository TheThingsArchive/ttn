// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package networkserver

import "github.com/TheThingsNetwork/ttn/utils/errors"

// Validate implements the api.Validator interface
func (m *DevicesRequest) Validate() error {
	if m.DevAddr == nil || m.DevAddr.IsEmpty() {
		return errors.NewErrInvalidArgument("DevAddr", "can not be empty")
	}
	return nil
}

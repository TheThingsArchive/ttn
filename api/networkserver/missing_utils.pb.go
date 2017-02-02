// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package networkserver

import "github.com/TheThingsNetwork/ttn/core/types"

func (m *DevicesRequest) GetDevAddr() *types.DevAddr {
	if m != nil {
		return m.DevAddr
	}
	return nil
}

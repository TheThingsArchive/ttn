// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package lorawan

import "github.com/TheThingsNetwork/ttn/core/types"

func (m *DeviceIdentifier) GetAppEui() *types.AppEUI {
	if m != nil {
		return m.AppEui
	}
	return nil
}

func (m *DeviceIdentifier) GetDevEui() *types.DevEUI {
	if m != nil {
		return m.DevEui
	}
	return nil
}

func (m *Device) GetAppEui() *types.AppEUI {
	if m != nil {
		return m.AppEui
	}
	return nil
}

func (m *Device) GetDevEui() *types.DevEUI {
	if m != nil {
		return m.DevEui
	}
	return nil
}

func (m *Device) GetDevAddr() *types.DevAddr {
	if m != nil {
		return m.DevAddr
	}
	return nil
}

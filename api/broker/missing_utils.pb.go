// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package broker

import "github.com/TheThingsNetwork/ttn/core/types"

func (m *UplinkMessage) GetAppEui() *types.AppEUI {
	if m != nil {
		return m.AppEui
	}
	return nil
}

func (m *UplinkMessage) GetDevEui() *types.DevEUI {
	if m != nil {
		return m.DevEui
	}
	return nil
}

func (m *DownlinkMessage) GetAppEui() *types.AppEUI {
	if m != nil {
		return m.AppEui
	}
	return nil
}

func (m *DownlinkMessage) GetDevEui() *types.DevEUI {
	if m != nil {
		return m.DevEui
	}
	return nil
}

func (m *DeviceActivationRequest) GetAppEui() *types.AppEUI {
	if m != nil {
		return m.AppEui
	}
	return nil
}

func (m *DeviceActivationRequest) GetDevEui() *types.DevEUI {
	if m != nil {
		return m.DevEui
	}
	return nil
}

func (m *DeduplicatedUplinkMessage) GetAppEui() *types.AppEUI {
	if m != nil {
		return m.AppEui
	}
	return nil
}

func (m *DeduplicatedUplinkMessage) GetDevEui() *types.DevEUI {
	if m != nil {
		return m.DevEui
	}
	return nil
}

func (m *DeduplicatedDeviceActivationRequest) GetAppEui() *types.AppEUI {
	if m != nil {
		return m.AppEui
	}
	return nil
}

func (m *DeduplicatedDeviceActivationRequest) GetDevEui() *types.DevEUI {
	if m != nil {
		return m.DevEui
	}
	return nil
}

func (m *ActivationChallengeRequest) GetAppEui() *types.AppEUI {
	if m != nil {
		return m.AppEui
	}
	return nil
}

func (m *ActivationChallengeRequest) GetDevEui() *types.DevEUI {
	if m != nil {
		return m.DevEui
	}
	return nil
}

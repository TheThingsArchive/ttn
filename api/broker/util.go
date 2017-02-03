// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package broker

import "github.com/TheThingsNetwork/ttn/api/protocol"

func (m *DeduplicatedUplinkMessage) InitResponseTemplate() *DownlinkMessage {
	if m.ResponseTemplate == nil {
		m.ResponseTemplate = new(DownlinkMessage)
	}
	m.ResponseTemplate.Message = new(protocol.Message)
	m.ResponseTemplate.AppEui = m.AppEui
	m.ResponseTemplate.DevEui = m.DevEui
	m.ResponseTemplate.AppId = m.AppId
	m.ResponseTemplate.DevId = m.DevId
	return m.ResponseTemplate
}

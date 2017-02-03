// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package protocol

import "github.com/TheThingsNetwork/ttn/api/protocol/lorawan"

// InitLoRaWAN initializes a LoRaWAN message
func (m *Message) InitLoRaWAN() *lorawan.Message {
	m.Protocol = &Message_Lorawan{
		Lorawan: &lorawan.Message{},
	}
	return m.GetLorawan()
}

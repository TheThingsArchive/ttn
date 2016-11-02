// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package handler

import (
	"errors"

	pb_protocol "github.com/TheThingsNetwork/ttn/api/protocol"
	pb_lorawan "github.com/TheThingsNetwork/ttn/api/protocol/lorawan"
	"github.com/brocaar/lorawan"
)

// UnmarshalPayload unmarshals the Payload into Message if Message is nil
func (m *DeviceActivationResponse) UnmarshalPayload() error {
	if m.GetMessage() == nil && m.GetDownlinkOption() != nil && m.DownlinkOption.GetProtocolConfig() != nil && m.DownlinkOption.ProtocolConfig.GetLorawan() != nil {
		var phy lorawan.PHYPayload
		if err := phy.UnmarshalBinary(m.Payload); err != nil {
			return err
		}
		msg := pb_lorawan.MessageFromPHYPayload(phy)
		m.Message = &pb_protocol.Message{Protocol: &pb_protocol.Message_Lorawan{Lorawan: &msg}}
	}
	return nil
}

// MarshalPayload marshals the Message into Payload if Payload is nil
func (m *DeviceActivationResponse) MarshalPayload() error {
	if m.Payload == nil && m.GetMessage() != nil {
		if m.Message.GetLorawan() == nil {
			return errors.New("No LoRaWAN message to marshal")
		}
		phy := m.Message.GetLorawan().PHYPayload()
		bin, err := phy.MarshalBinary()
		if err != nil {
			return err
		}
		m.Payload = bin
	}
	return nil
}

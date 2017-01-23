// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package broker

import (
	"errors"

	pb_protocol "github.com/TheThingsNetwork/ttn/api/protocol"
	pb_lorawan "github.com/TheThingsNetwork/ttn/api/protocol/lorawan"
	"github.com/brocaar/lorawan"
)

func msgFromPayload(payload []byte) (*pb_protocol.Message, error) {
	var phy lorawan.PHYPayload
	if err := phy.UnmarshalBinary(payload); err != nil {
		return nil, err
	}
	msg := pb_lorawan.MessageFromPHYPayload(phy)
	return &pb_protocol.Message{Protocol: &pb_protocol.Message_Lorawan{Lorawan: &msg}}, nil
}

// UnmarshalPayload unmarshals the Payload into Message if Message is nil
func (m *UplinkMessage) UnmarshalPayload() error {
	if m.GetMessage() == nil && m.GetProtocolMetadata() != nil && m.ProtocolMetadata.GetLorawan() != nil {
		msg, err := msgFromPayload(m.Payload)
		if err != nil {
			return err
		}
		m.Message = msg
	}
	return nil
}

// UnmarshalPayload unmarshals the Payload into Message if Message is nil
func (m *DownlinkMessage) UnmarshalPayload() error {
	if m.GetMessage() == nil && m.GetDownlinkOption() != nil && m.DownlinkOption.GetProtocolConfig() != nil && m.DownlinkOption.ProtocolConfig.GetLorawan() != nil {
		msg, err := msgFromPayload(m.Payload)
		if err != nil {
			return err
		}
		m.Message = msg
		if lorawan := m.GetDownlinkOption().GetProtocolConfig().GetLorawan(); lorawan != nil {
			if lorawan.FCnt != 0 {
				if mac := m.Message.GetLorawan().GetMacPayload(); mac != nil {
					mac.FHDR.FCnt = lorawan.FCnt
				}
			}
		}
	}
	return nil
}

// UnmarshalPayload unmarshals the Payload into Message if Message is nil
func (m *DeviceActivationResponse) UnmarshalPayload() error {
	if m.GetMessage() == nil && m.GetDownlinkOption() != nil && m.DownlinkOption.GetProtocolConfig() != nil && m.DownlinkOption.ProtocolConfig.GetLorawan() != nil {
		msg, err := msgFromPayload(m.Payload)
		if err != nil {
			return err
		}
		m.Message = msg
	}
	return nil
}

// UnmarshalPayload unmarshals the Payload into Message if Message is nil
func (m *DeduplicatedUplinkMessage) UnmarshalPayload() error {
	if m.GetMessage() == nil && m.GetProtocolMetadata() != nil && m.ProtocolMetadata.GetLorawan() != nil {
		msg, err := msgFromPayload(m.Payload)
		if err != nil {
			return err
		}
		m.Message = msg
		if lorawan := m.GetProtocolMetadata().GetLorawan(); lorawan != nil {
			if lorawan.FCnt != 0 {
				if mac := m.Message.GetLorawan().GetMacPayload(); mac != nil {
					mac.FHDR.FCnt = lorawan.FCnt
				}
			}
		}
	}
	return nil
}

// UnmarshalPayload unmarshals the Payload into Message if Message is nil
func (m *DeviceActivationRequest) UnmarshalPayload() error {
	if m.GetMessage() == nil && m.GetProtocolMetadata() != nil && m.ProtocolMetadata.GetLorawan() != nil {
		msg, err := msgFromPayload(m.Payload)
		if err != nil {
			return err
		}
		m.Message = msg
	}
	return nil
}

// UnmarshalPayload unmarshals the Payload into Message if Message is nil
func (m *DeduplicatedDeviceActivationRequest) UnmarshalPayload() error {
	if m.GetMessage() == nil && m.GetProtocolMetadata() != nil && m.ProtocolMetadata.GetLorawan() != nil {
		msg, err := msgFromPayload(m.Payload)
		if err != nil {
			return err
		}
		m.Message = msg
	}
	return nil
}

// UnmarshalPayload unmarshals the Payload into Message if Message is nil
func (m *ActivationChallengeRequest) UnmarshalPayload() error {
	if m.GetMessage() == nil {
		msg, err := msgFromPayload(m.Payload)
		if err != nil {
			return err
		}
		m.Message = msg
	}
	return nil
}

// UnmarshalPayload unmarshals the Payload into Message if Message is nil
func (m *ActivationChallengeResponse) UnmarshalPayload() error {
	if m.GetMessage() == nil {
		msg, err := msgFromPayload(m.Payload)
		if err != nil {
			return err
		}
		m.Message = msg
	}
	return nil
}

func payloadFromMsg(msg *pb_protocol.Message) ([]byte, error) {
	if msg.GetLorawan() == nil {
		return nil, errors.New("No LoRaWAN message to marshal")
	}
	phy := msg.GetLorawan().PHYPayload()
	bin, err := phy.MarshalBinary()
	if err != nil {
		return nil, err
	}
	return bin, nil
}

// MarshalPayload marshals the Message into Payload if Payload is nil
func (m *UplinkMessage) MarshalPayload() error {
	if m.Payload == nil && m.GetMessage() != nil {
		bin, err := payloadFromMsg(m.Message)
		if err != nil {
			return err
		}
		m.Payload = bin
	}
	return nil
}

// MarshalPayload marshals the Message into Payload if Payload is nil
func (m *DownlinkMessage) MarshalPayload() error {
	if m.Payload == nil && m.GetMessage() != nil {
		bin, err := payloadFromMsg(m.Message)
		if err != nil {
			return err
		}
		m.Payload = bin
	}
	return nil
}

// MarshalPayload marshals the Message into Payload if Payload is nil
func (m *DeviceActivationResponse) MarshalPayload() error {
	if m.Payload == nil && m.GetMessage() != nil {
		bin, err := payloadFromMsg(m.Message)
		if err != nil {
			return err
		}
		m.Payload = bin
	}
	return nil
}

// MarshalPayload marshals the Message into Payload if Payload is nil
func (m *DeduplicatedUplinkMessage) MarshalPayload() error {
	if m.Payload == nil && m.GetMessage() != nil {
		bin, err := payloadFromMsg(m.Message)
		if err != nil {
			return err
		}
		m.Payload = bin
	}
	return nil
}

// MarshalPayload marshals the Message into Payload if Payload is nil
func (m *DeviceActivationRequest) MarshalPayload() error {
	if m.Payload == nil && m.GetMessage() != nil {
		bin, err := payloadFromMsg(m.Message)
		if err != nil {
			return err
		}
		m.Payload = bin
	}
	return nil
}

// MarshalPayload marshals the Message into Payload if Payload is nil
func (m *DeduplicatedDeviceActivationRequest) MarshalPayload() error {
	if m.Payload == nil && m.GetMessage() != nil {
		bin, err := payloadFromMsg(m.Message)
		if err != nil {
			return err
		}
		m.Payload = bin
	}
	return nil
}

// MarshalPayload marshals the Message into Payload if Payload is nil
func (m *ActivationChallengeRequest) MarshalPayload() error {
	if m.Payload == nil && m.GetMessage() != nil {
		bin, err := payloadFromMsg(m.Message)
		if err != nil {
			return err
		}
		m.Payload = bin
	}
	return nil
}

// MarshalPayload marshals the Message into Payload if Payload is nil
func (m *ActivationChallengeResponse) MarshalPayload() error {
	if m.Payload == nil && m.GetMessage() != nil {
		bin, err := payloadFromMsg(m.Message)
		if err != nil {
			return err
		}
		m.Payload = bin
	}
	return nil
}

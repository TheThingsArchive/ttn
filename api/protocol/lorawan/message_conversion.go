// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package lorawan

import (
	"github.com/TheThingsNetwork/ttn/core/types"
	"github.com/brocaar/lorawan"
	"github.com/brocaar/lorawan/band"
)

type payloader interface {
	Payload() lorawan.Payload
}

type macCommandPayload []byte

func (m macCommandPayload) MarshalBinary() (data []byte, err error) {
	return []byte(m), nil
}

func (m *macCommandPayload) UnmarshalBinary(data []byte) error {
	*m = data
	return nil
}

// MACCommand converts the MACCommand to a lorawan.MACCommand
func (m *MACCommand) MACCommand() (cmd lorawan.MACCommand) {
	cmd.CID = lorawan.CID(m.Cid)
	payload := macCommandPayload(m.Payload)
	cmd.Payload = &payload
	return
}

// MACCommandFromMACCommand creates a new MACCommand from a lorawan.MACCommand
func MACCommandFromMACCommand(cmd lorawan.MACCommand) (m MACCommand) {
	m.Cid = uint32(cmd.CID)
	if cmd.Payload != nil {
		m.Payload, _ = cmd.Payload.MarshalBinary()
	}
	return
}

// Payload converts the MACPayload to a lorawan.Payload
func (msg *Message_MacPayload) Payload() lorawan.Payload {
	m := *msg.MacPayload
	var mac lorawan.MACPayload
	mac.FHDR.DevAddr = lorawan.DevAddr(m.DevAddr)
	mac.FHDR.FCtrl.ADR = m.Adr
	mac.FHDR.FCtrl.ADRACKReq = m.AdrAckReq
	mac.FHDR.FCtrl.ACK = m.Ack
	mac.FHDR.FCtrl.FPending = m.FPending
	mac.FHDR.FCnt = m.FCnt
	for _, cmd := range m.FOpts {
		mac.FHDR.FOpts = append(mac.FHDR.FOpts, cmd.MACCommand())
	}
	if m.FPort > 0 {
		fPort := uint8(m.FPort)
		mac.FPort = &fPort
	}
	if len(m.FrmPayload) != 0 {
		mac.FRMPayload = []lorawan.Payload{&lorawan.DataPayload{Bytes: m.FrmPayload}}
		if mac.FPort == nil {
			fPort := uint8(0)
			mac.FPort = &fPort
		}
	}
	return &mac
}

// MACPayloadFromPayload creates a new MACPayload from a lorawan.Payload
func MACPayloadFromPayload(payload lorawan.Payload) (mac MACPayload) {
	if payload, ok := payload.(*lorawan.MACPayload); ok {
		mac.DevAddr = types.DevAddr(payload.FHDR.DevAddr)
		mac.Adr = payload.FHDR.FCtrl.ADR
		mac.AdrAckReq = payload.FHDR.FCtrl.ADRACKReq
		mac.Ack = payload.FHDR.FCtrl.ACK
		mac.FPending = payload.FHDR.FCtrl.FPending
		mac.FCnt = payload.FHDR.FCnt
		for _, cmd := range payload.FHDR.FOpts {
			mac.FOpts = append(mac.FOpts, MACCommandFromMACCommand(cmd))
		}
		if payload.FPort != nil {
			mac.FPort = int32(*payload.FPort)
		}
		if len(payload.FRMPayload) == 1 {
			if payload, ok := payload.FRMPayload[0].(*lorawan.DataPayload); ok {
				mac.FrmPayload = payload.Bytes
			}
		}
	}
	return
}

// Payload converts the JoinRequestPayload to a lorawan.Payload
func (msg *Message_JoinRequestPayload) Payload() lorawan.Payload {
	m := *msg.JoinRequestPayload
	var mac lorawan.JoinRequestPayload
	mac.AppEUI = lorawan.EUI64(m.AppEui)
	mac.DevEUI = lorawan.EUI64(m.DevEui)
	mac.DevNonce = m.DevNonce
	return &mac
}

// JoinRequestPayloadFromPayload creates a new JoinRequestPayload from a lorawan.Payload
func JoinRequestPayloadFromPayload(payload lorawan.Payload) (request JoinRequestPayload) {
	if payload, ok := payload.(*lorawan.JoinRequestPayload); ok {
		request.AppEui = types.AppEUI(payload.AppEUI)
		request.DevEui = types.DevEUI(payload.DevEUI)
		request.DevNonce = types.DevNonce(payload.DevNonce)
	}
	return
}

// Payload converts the JoinAcceptPayload to a lorawan.Payload
func (msg *Message_JoinAcceptPayload) Payload() lorawan.Payload {
	m := *msg.JoinAcceptPayload
	if len(m.Encrypted) != 0 {
		return &lorawan.DataPayload{Bytes: m.Encrypted}
	}
	var mac lorawan.JoinAcceptPayload
	mac.AppNonce = m.AppNonce
	mac.NetID = m.NetId
	mac.DevAddr = lorawan.DevAddr(m.DevAddr)
	mac.DLSettings.RX1DROffset = uint8(m.Rx1DrOffset)
	mac.DLSettings.RX2DataRate = uint8(m.Rx2Dr)
	mac.RXDelay = uint8(m.RxDelay)
	if m.CfList != nil && len(m.CfList.Freq) == 5 {
		mac.CFList = &lorawan.CFList{
			m.CfList.Freq[0],
			m.CfList.Freq[1],
			m.CfList.Freq[2],
			m.CfList.Freq[3],
			m.CfList.Freq[4],
		}
	}
	return &mac
}

// JoinAcceptPayloadFromPayload creates a new JoinAcceptPayload from a lorawan.Payload
func JoinAcceptPayloadFromPayload(payload lorawan.Payload) (accept JoinAcceptPayload) {
	if dataPayload, ok := payload.(*lorawan.DataPayload); ok {
		accept.Encrypted = dataPayload.Bytes
		joinAccept := &lorawan.JoinAcceptPayload{}
		joinAccept.UnmarshalBinary(false, dataPayload.Bytes)
		payload = joinAccept
	}

	if payload, ok := payload.(*lorawan.JoinAcceptPayload); ok {
		accept.AppNonce = types.AppNonce(payload.AppNonce)
		accept.NetId = types.NetID(payload.NetID)
		accept.DevAddr = types.DevAddr(payload.DevAddr)
		accept.DLSettings.Rx1DrOffset = uint32(payload.DLSettings.RX1DROffset)
		accept.DLSettings.Rx2Dr = uint32(payload.DLSettings.RX2DataRate)
		accept.RxDelay = uint32(payload.RXDelay)
		if payload.CFList != nil {
			accept.CfList = &CFList{
				Freq: payload.CFList[:],
			}
		}
	}
	return
}

// PHYPayload converts the Message to a lorawan.PHYPayload
func (m *Message) PHYPayload() (phy lorawan.PHYPayload) {
	phy.MHDR.Major = lorawan.Major(m.Major)
	phy.MHDR.MType = lorawan.MType(m.MType)
	if m.Payload != nil {
		phy.MACPayload = m.Payload.(payloader).Payload()
	}
	copy(phy.MIC[:], m.Mic)
	return
}

// PHYPayloadBytes converts the Message to a lorawan.PHYPayload, marshals it and returns the bytes
func (m *Message) PHYPayloadBytes() []byte {
	phy := m.PHYPayload()
	bytes, _ := phy.MarshalBinary()
	return bytes
}

// MessageFromPHYPayloadBytes converts lorawan.PHYPayload bytes to a Message
func MessageFromPHYPayloadBytes(payload []byte) (msg Message, err error) {
	var phy lorawan.PHYPayload
	err = phy.UnmarshalBinary(payload)
	if err != nil {
		return
	}
	msg = MessageFromPHYPayload(phy)
	return
}

// MessageFromPHYPayload converts a lorawan.PHYPayload to a Message
func MessageFromPHYPayload(phy lorawan.PHYPayload) Message {
	var m Message
	m.Major = Major(phy.MHDR.Major)
	m.MType = MType(phy.MHDR.MType)
	m.Mic = phy.MIC[:]
	switch m.MType {
	case MType_JOIN_REQUEST:
		payload := JoinRequestPayloadFromPayload(phy.MACPayload)
		m.Payload = &Message_JoinRequestPayload{JoinRequestPayload: &payload}
	case MType_JOIN_ACCEPT:
		payload := JoinAcceptPayloadFromPayload(phy.MACPayload)
		m.Payload = &Message_JoinAcceptPayload{JoinAcceptPayload: &payload}
	case MType_UNCONFIRMED_UP,
		MType_UNCONFIRMED_DOWN,
		MType_CONFIRMED_UP,
		MType_CONFIRMED_DOWN:
		payload := MACPayloadFromPayload(phy.MACPayload)
		m.Payload = &Message_MacPayload{MacPayload: &payload}
	}
	return m
}

// GetLoRaWANDataRate returns the band.Datarate for the current Metadata
func (m *Metadata) GetLoRaWANDataRate() (dataRate band.DataRate, err error) {
	switch m.Modulation {
	case Modulation_LORA:
		dataRate.Modulation = band.LoRaModulation
		dr, err := types.ParseDataRate(m.DataRate)
		if err != nil {
			return dataRate, err
		}
		dataRate.Bandwidth = int(dr.Bandwidth)
		dataRate.SpreadFactor = int(dr.SpreadingFactor)
	case Modulation_FSK:
		dataRate.Modulation = band.FSKModulation
		dataRate.BitRate = int(m.BitRate)
	}
	return
}

// SetDataRate sets the dataRate for the current Metadata based from a band.Datarate
func (c *TxConfiguration) SetDataRate(dataRate band.DataRate) error {
	switch dataRate.Modulation {
	case band.LoRaModulation:
		c.Modulation = Modulation_LORA
		datr, err := types.ConvertDataRate(dataRate)
		if err != nil {
			return err
		}
		c.DataRate = datr.String()
	case band.FSKModulation:
		c.Modulation = Modulation_FSK
		c.BitRate = uint32(dataRate.BitRate)
	}
	return nil
}

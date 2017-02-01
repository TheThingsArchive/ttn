// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package util

import (
	"github.com/TheThingsNetwork/ttn/utils/errors"

	"github.com/TheThingsNetwork/ttn/core/types"
	"github.com/TheThingsNetwork/ttn/utils/pointer"
	"github.com/brocaar/lorawan"
)

// Message struct is used to construct an uplink message
type Message struct {
	devAddr   types.DevAddr
	nwkSKey   types.NwkSKey
	appSKey   types.AppSKey
	FCnt      int
	FPort     int
	confirmed bool
	ack       bool
	Payload   []byte
}

// SetDevice with the LoRaWAN options
func (m *Message) SetDevice(devAddr types.DevAddr, nwkSKey types.NwkSKey, appSKey types.AppSKey) {
	m.devAddr = devAddr
	m.nwkSKey = nwkSKey
	m.appSKey = appSKey
}

// SetMessage with some options
func (m *Message) SetMessage(confirmed bool, ack bool, fCnt int, payload []byte) {
	m.confirmed = confirmed
	m.ack = ack
	m.FCnt = fCnt
	m.Payload = payload
}

// Bytes returns the bytes
func (m *Message) Bytes() []byte {
	if m.FPort == 0 && len(m.Payload) > 0 {
		m.FPort = 1
	}
	macPayload := &lorawan.MACPayload{}
	macPayload.FHDR = lorawan.FHDR{
		DevAddr: lorawan.DevAddr(m.devAddr),
		FCnt:    uint32(m.FCnt),
		FCtrl: lorawan.FCtrl{
			ACK: m.ack,
		},
	}
	macPayload.FPort = pointer.Uint8(uint8(m.FPort))
	if len(m.Payload) > 0 {
		macPayload.FRMPayload = []lorawan.Payload{&lorawan.DataPayload{Bytes: m.Payload}}
	}
	phyPayload := &lorawan.PHYPayload{}
	phyPayload.MHDR = lorawan.MHDR{
		MType: lorawan.UnconfirmedDataUp,
		Major: lorawan.LoRaWANR1,
	}
	if m.confirmed {
		phyPayload.MHDR.MType = lorawan.ConfirmedDataUp
	}
	phyPayload.MACPayload = macPayload
	phyPayload.EncryptFRMPayload(lorawan.AES128Key(m.appSKey))
	phyPayload.SetMIC(lorawan.AES128Key(m.nwkSKey))
	uplinkBytes, _ := phyPayload.MarshalBinary()
	return uplinkBytes
}

// Unmarshal a byte slice into a Message
func (m *Message) Unmarshal(bytes []byte) error {
	payload := &lorawan.PHYPayload{}
	payload.UnmarshalBinary(bytes)
	if micOK, _ := payload.ValidateMIC(lorawan.AES128Key(m.nwkSKey)); !micOK {
		return errors.New("Invalid MIC")
	}
	macPayload, ok := payload.MACPayload.(*lorawan.MACPayload)
	if !ok {
		return errors.New("No MACPayload")
	}
	m.FCnt = int(macPayload.FHDR.FCnt)
	m.FPort = -1
	if macPayload.FPort != nil {
		m.FPort = int(*macPayload.FPort)
	}
	m.Payload = []byte{}
	if len(macPayload.FRMPayload) > 0 {
		payload.DecryptFRMPayload(lorawan.AES128Key(m.appSKey))
		m.Payload = macPayload.FRMPayload[0].(*lorawan.DataPayload).Bytes
	}
	return nil
}

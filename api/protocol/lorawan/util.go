// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package lorawan

import (
	"github.com/TheThingsNetwork/ttn/core/types"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	"github.com/brocaar/lorawan"
)

func (m *Message) init() {
	m.Major = Major_LORAWAN_R1
	m.Mic = make([]byte, 4)
}

func (m *Message) initMACPayload() *MACPayload {
	m.init()
	macPayload := new(MACPayload)
	m.Payload = &Message_MacPayload{MacPayload: macPayload}
	return macPayload
}

// InitUplink initializes an unconfirmed uplink message
func (m *Message) InitUplink() *MACPayload {
	mac := m.initMACPayload()
	m.MType = MType_UNCONFIRMED_UP
	return mac
}

// InitDownlink initializes an unconfirmed downlink message
func (m *Message) InitDownlink() *MACPayload {
	mac := m.initMACPayload()
	m.MType = MType_UNCONFIRMED_DOWN
	return mac
}

// IsConfirmed returns wheter the message is a confirmed message
func (m *Message) IsConfirmed() bool {
	return m.MType == MType_CONFIRMED_UP || m.MType == MType_CONFIRMED_DOWN
}

// SetMIC sets the MIC of the message
func (m *Message) SetMIC(nwkSKey types.NwkSKey) error {
	phy := m.PHYPayload()
	err := phy.SetMIC(lorawan.AES128Key(nwkSKey))
	if err != nil {
		return err
	}
	m.Mic = phy.MIC[:]
	return nil
}

// ValidateMIC validates the MIC of the message
func (m *Message) ValidateMIC(nwkSKey types.NwkSKey) error {
	ok, err := m.PHYPayload().ValidateMIC(lorawan.AES128Key(nwkSKey))
	if err != nil {
		return err
	}
	if !ok {
		return errors.NewErrInvalidArgument("Uplink", "Invalid MIC")
	}
	return nil
}

func (m *Message) cryptFRMPayload(appSKey types.AppSKey) error {
	phy := m.PHYPayload()
	if err := phy.DecryptFRMPayload(lorawan.AES128Key(appSKey)); err != nil {
		return err
	}
	crypted := MessageFromPHYPayload(phy)
	if m.GetMacPayload() != nil && crypted.GetMacPayload() != nil {
		m.GetMacPayload().FrmPayload = crypted.GetMacPayload().FrmPayload
	}
	return nil
}

// EncryptFRMPayload encrypts the FRMPayload
func (m *Message) EncryptFRMPayload(appSKey types.AppSKey) error {
	return m.cryptFRMPayload(appSKey)
}

// DecryptFRMPayload decrypts the FRMPayload
func (m *Message) DecryptFRMPayload(appSKey types.AppSKey) error {
	return m.cryptFRMPayload(appSKey)
}

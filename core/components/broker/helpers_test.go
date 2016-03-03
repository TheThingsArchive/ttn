// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package broker

import (
	"testing"

	"github.com/TheThingsNetwork/ttn/core"
	"github.com/TheThingsNetwork/ttn/core/mocks"
	"github.com/brocaar/lorawan"
)

// ----- BUILD utilities
func newBPacket(rawDevAddr [4]byte, payload string, nwkSKey [16]byte, fcnt uint32) core.BPacket {
	var devAddr lorawan.DevAddr
	copy(devAddr[:], rawDevAddr[:])

	macPayload := lorawan.NewMACPayload(true)
	macPayload.FHDR = lorawan.FHDR{
		FCnt:    fcnt,
		DevAddr: devAddr,
	}
	macPayload.FRMPayload = []lorawan.Payload{&lorawan.DataPayload{Bytes: []byte(payload)}}
	macPayload.FPort = 1
	phyPayload := lorawan.NewPHYPayload(true)
	phyPayload.MACPayload = macPayload
	phyPayload.MHDR = lorawan.MHDR{
		MType: lorawan.UnconfirmedDataUp,
		Major: lorawan.LoRaWANR1,
	}

	var key lorawan.AES128Key
	copy(key[:], nwkSKey[:])
	if err := phyPayload.SetMIC(key); err != nil {
		panic(err)
	}

	packet, err := core.NewBPacket(phyPayload, core.Metadata{})
	if err != nil {
		panic(err)
	}
	return packet
}

// ----- CHECK utilities
func CheckDevEntries(t *testing.T, want []devEntry, got []devEntry) {
	mocks.Check(t, want, got, "DevEntries")
}

func CheckAppEntries(t *testing.T, want appEntry, got appEntry) {
	mocks.Check(t, want, got, "AppEntries")
}

func CheckRegistrations(t *testing.T, want core.Registration, got core.Registration) {
	mocks.Check(t, want, got, "Registrations")
}

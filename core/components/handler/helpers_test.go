// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package handler

import (
	"testing"

	. "github.com/TheThingsNetwork/ttn/core"
	. "github.com/TheThingsNetwork/ttn/core/mocks"
	"github.com/brocaar/lorawan"
)

// ----- BUILD utilities
func newHPacket(appEUI [8]byte, devEUI [8]byte, payload string, metadata Metadata, fcnt uint32, appSKey [16]byte) HPacket {
	macPayload := lorawan.NewMACPayload(true)
	macPayload.FHDR = lorawan.FHDR{
		FCnt: fcnt,
	}
	macPayload.FPort = 1
	macPayload.FRMPayload = []lorawan.Payload{&lorawan.DataPayload{Bytes: []byte(payload)}}

	var key lorawan.AES128Key
	copy(key[:], appSKey[:])
	if err := macPayload.EncryptFRMPayload(key); err != nil {
		panic(err)
	}

	phyPayload := lorawan.NewPHYPayload(true)
	phyPayload.MHDR = lorawan.MHDR{
		MType: lorawan.UnconfirmedDataUp,
		Major: lorawan.LoRaWANR1,
	}
	phyPayload.MACPayload = macPayload

	var appEUIp lorawan.EUI64
	var devEUIp lorawan.EUI64
	copy(appEUIp[:], appEUI[:])
	copy(devEUIp[:], devEUI[:])

	packet, err := NewHPacket(appEUIp, devEUIp, phyPayload, metadata)
	if err != nil {
		panic(err)
	}
	return packet
}

func newBPacket(rawDevAddr [4]byte, payload string, metadata Metadata, fcnt uint32, appSKey [16]byte) BPacket {
	var devAddr lorawan.DevAddr
	copy(devAddr[:], rawDevAddr[:])

	macPayload := lorawan.NewMACPayload(false)
	macPayload.FHDR = lorawan.FHDR{
		DevAddr: devAddr,
		FCnt:    fcnt,
	}
	macPayload.FPort = 1
	macPayload.FRMPayload = []lorawan.Payload{&lorawan.DataPayload{Bytes: []byte(payload)}}

	var key lorawan.AES128Key
	copy(key[:], appSKey[:])
	if err := macPayload.EncryptFRMPayload(key); err != nil {
		panic(err)
	}

	phyPayload := lorawan.NewPHYPayload(false)
	phyPayload.MHDR = lorawan.MHDR{
		MType: lorawan.UnconfirmedDataDown,
		Major: lorawan.LoRaWANR1,
	}
	phyPayload.MACPayload = macPayload

	packet, err := NewBPacket(phyPayload, metadata)
	if err != nil {
		panic(err)
	}
	return packet

}

// ----- CHECK utilities
func CheckPushed(t *testing.T, want APacket, got APacket) {
	Check(t, want, got, "Pushed")
}

func CheckPersonalized(t *testing.T, want HRegistration, got HRegistration) {
	Check(t, want, got, "Personalized")
}

func CheckPackets(t *testing.T, want APacket, got APacket) {
	Check(t, want, got, "Packets")
}

func CheckEntries(t *testing.T, want MockHRegistration, got devEntry) {
	// NOTE This only works in the case of Personalized devices
	var devAddr lorawan.DevAddr
	devEUI := want.DevEUI()
	copy(devAddr[:], devEUI[4:])

	wantEntry := devEntry{
		Recipient: want.RawRecipient(),
		DevAddr:   devAddr,
		NwkSKey:   want.NwkSKey(),
		AppSKey:   want.AppSKey(),
	}

	Check(t, wantEntry, got, "Entries")
}

func CheckSubscriptions(t *testing.T, want BRegistration, got Registration) {
	var mockGot BRegistration
	bgot, ok := got.(BRegistration)
	if got != nil && ok {
		r := NewMockBRegistration()
		r.OutRecipient = bgot.Recipient()
		r.OutDevEUI = bgot.DevEUI()
		r.OutAppEUI = bgot.AppEUI()
		r.OutNwkSKey = bgot.NwkSKey()
		mockGot = r
	}
	Check(t, want, mockGot, "Subscriptions")
}

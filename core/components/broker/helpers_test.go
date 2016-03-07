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

	phyPayload.MACPayload.(*lorawan.MACPayload).FHDR.FCnt = fcnt % 65536 // only 16-bits

	packet, err := core.NewBPacket(phyPayload, core.Metadata{})
	if err != nil {
		panic(err)
	}
	return packet
}

func newBPacketDown(fcnt uint32) core.BPacket {
	macPayload := lorawan.NewMACPayload(false)
	macPayload.FHDR = lorawan.FHDR{
		FCnt:    fcnt,
		DevAddr: lorawan.DevAddr([4]byte{1, 1, 1, 1}),
	}
	macPayload.FRMPayload = []lorawan.Payload{&lorawan.DataPayload{Bytes: []byte("downlink")}}
	macPayload.FPort = 1
	phyPayload := lorawan.NewPHYPayload(false)
	phyPayload.MACPayload = macPayload
	phyPayload.MHDR = lorawan.MHDR{
		MType: lorawan.UnconfirmedDataDown,
		Major: lorawan.LoRaWANR1,
	}

	if err := phyPayload.SetMIC(lorawan.AES128Key([16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6})); err != nil {
		panic(err)
	}

	phyPayload.MACPayload.(*lorawan.MACPayload).FHDR.FCnt = fcnt % 65536 // only 16-bits

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

func CheckCounters(t *testing.T, want uint32, got uint32) {
	mocks.Check(t, want, got, "Counters")
}

func CheckDirections(t *testing.T, want string, got string) {
	mocks.Check(t, want, got, "Directions")
}

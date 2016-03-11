// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package router

import (
	"testing"
	"time"

	. "github.com/TheThingsNetwork/ttn/core"
	. "github.com/TheThingsNetwork/ttn/core/mocks"
	. "github.com/TheThingsNetwork/ttn/utils/testing"
	"github.com/brocaar/lorawan"
)

// ----- BUILD utilities
func newRPacket(rawDevAddr [4]byte, payload string, gatewayID []byte, metadata Metadata) RPacket {
	var devAddr lorawan.DevAddr
	copy(devAddr[:], rawDevAddr[:])

	macPayload := lorawan.NewMACPayload(true)
	macPayload.FHDR.DevAddr = devAddr
	macPayload.FRMPayload = []lorawan.Payload{&lorawan.DataPayload{Bytes: []byte(payload)}}
	macPayload.FPort = 1
	phyPayload := lorawan.NewPHYPayload(true)
	phyPayload.MACPayload = macPayload
	phyPayload.MHDR = lorawan.MHDR{
		MType: lorawan.UnconfirmedDataUp,
		Major: lorawan.LoRaWANR1,
	}

	packet, err := NewRPacket(phyPayload, gatewayID, metadata)
	if err != nil {
		panic(err)
	}
	return packet
}

func newBPacket(rawDevAddr [4]byte, payload string, metadata Metadata) BPacket {
	var devAddr lorawan.DevAddr
	copy(devAddr[:], rawDevAddr[:])

	macPayload := lorawan.NewMACPayload(true)
	macPayload.FHDR.DevAddr = devAddr
	macPayload.FRMPayload = []lorawan.Payload{&lorawan.DataPayload{Bytes: []byte(payload)}}
	macPayload.FPort = 1
	phyPayload := lorawan.NewPHYPayload(true)
	phyPayload.MACPayload = macPayload
	phyPayload.MHDR = lorawan.MHDR{
		MType: lorawan.UnconfirmedDataUp,
		Major: lorawan.LoRaWANR1,
	}

	packet, err := NewBPacket(phyPayload, metadata)
	if err != nil {
		panic(err)
	}
	return packet
}

// ----- CHECK utilities
func CheckEntries(t *testing.T, want []entry, got []entry) {
	for i, w := range want {
		if i >= len(got) {
			Ko(t, "Didn't got enough entries: %v", got)
		}
		tmin := w.until.Add(-time.Second)
		tmax := w.until.Add(time.Second)
		if !tmin.Before(got[i].until) || !got[i].until.Before(tmax) {
			Ko(t, "Unexpected expiry time.\nWant: %s\nGot:  %s", w.until, got[i].until)
		}
		Check(t, w.Recipient, got[i].Recipient, "Recipients")
	}
}

func CheckRegistrations(t *testing.T, want Registration, got Registration) {
	Check(t, want, got, "Registrations")
}

func CheckIDs(t *testing.T, want []byte, got []byte) {
	Check(t, want, got, "IDs")
}

func CheckStats(t *testing.T, want SPacket, got SPacket) {
	Check(t, want, got, "Stats")
}

func CheckMetadata(t *testing.T, want Metadata, got Metadata) {
	Check(t, want, got, "Metadata")
}

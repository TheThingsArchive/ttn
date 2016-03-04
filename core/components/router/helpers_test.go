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
func newRPacket(rawDevAddr [4]byte, payload string, gatewayID []byte) RPacket {
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

	packet, err := NewRPacket(phyPayload, gatewayID, Metadata{})
	if err != nil {
		panic(err)
	}
	return packet
}

func newBPacket(rawDevAddr [4]byte, payload string) BPacket {
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

	packet, err := NewBPacket(phyPayload, Metadata{})
	if err != nil {
		panic(err)
	}
	return packet
}

// ----- CHECK utilities
func CheckEntries(t *testing.T, want entry, got entry) {
	tmin := want.until.Add(-time.Second)
	tmax := want.until.Add(time.Second)
	if !tmin.Before(got.until) || !got.until.Before(tmax) {
		Ko(t, "Unexpected expiry time.\nWant: %s\nGot:  %s", want.until, got.until)
	}
	Check(t, want.Recipient, got.Recipient, "Recipients")
}

func CheckRegistrations(t *testing.T, want Registration, got Registration) {
	Check(t, want, got, "Registrations")
}

func CheckSubBands(t *testing.T, want subBand, got subBand) {
	Check(t, want, got, "SubBands")
}

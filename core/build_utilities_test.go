// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package core

import (
	"time"

	"github.com/TheThingsNetwork/ttn/utils/pointer"
	"github.com/brocaar/lorawan"
)

// Generates a Metadata object with all field completed with relevant values
func genFullMetadata() Metadata {
	timeRef := time.Date(2016, 1, 13, 14, 11, 28, 207288421, time.UTC)
	return Metadata{
		Chan: pointer.Uint(2),
		Codr: pointer.String("4/6"),
		Datr: pointer.String("LORA"),
		Fdev: pointer.Uint(3),
		Freq: pointer.Float64(863.125),
		Imme: pointer.Bool(false),
		Ipol: pointer.Bool(false),
		Lsnr: pointer.Float64(5.2),
		Modu: pointer.String("LORA"),
		Ncrc: pointer.Bool(true),
		Powe: pointer.Uint(3),
		Prea: pointer.Uint(8),
		Rfch: pointer.Uint(2),
		Rssi: pointer.Int(-27),
		Size: pointer.Uint(14),
		Stat: pointer.Int(0),
		Time: pointer.Time(timeRef),
		Tmst: pointer.Uint(uint(timeRef.UnixNano())),
	}
}

// Generate a Physical payload representing an uplink or downlink message
func genPHYPayload(uplink bool) lorawan.PHYPayload {
	nwkSKey := [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}
	appSKey := [16]byte{16, 15, 14, 13, 12, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1}

	macPayload := lorawan.NewMACPayload(uplink)
	macPayload.FHDR = lorawan.FHDR{
		DevAddr: lorawan.DevAddr([4]byte{1, 2, 3, 4}),
		FCtrl: lorawan.FCtrl{
			ADR:       false,
			ADRACKReq: false,
			ACK:       false,
		},
		FCnt: 0,
	}
	macPayload.FPort = 10
	macPayload.FRMPayload = []lorawan.Payload{&lorawan.DataPayload{Bytes: []byte{1, 2, 3, 4}}}

	if err := macPayload.EncryptFRMPayload(appSKey); err != nil {
		panic(err)
	}

	payload := lorawan.NewPHYPayload(uplink)
	payload.MHDR = lorawan.MHDR{
		MType: lorawan.ConfirmedDataUp,
		Major: lorawan.LoRaWANR1,
	}
	payload.MACPayload = macPayload

	if err := payload.SetMIC(nwkSKey); err != nil {
		panic(err)
	}

	return payload
}

// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package components

import (
	"encoding/base64"
	"github.com/thethingsnetwork/core"
	"github.com/thethingsnetwork/core/lorawan"
	"github.com/thethingsnetwork/core/semtech"
	"github.com/thethingsnetwork/core/utils/pointer"
	. "github.com/thethingsnetwork/core/utils/testing"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestConvertRXPK(t *testing.T) {
	tests := []convertRXPKTest{
		genRXPKWithFullMetadata(),
		genRXPKWithPartialMetadata(),
		genRXPKWithNoData(),
	}

	for _, test := range tests {
		Desc(t, "Convert RXPK: %s", pointer.DumpPStruct(test.RXPK, false))
		packet, err := ConvertRXPK(test.RXPK)
		checkErrors(t, test.WantError, err)
		checkPackets(t, test.CorePacket, packet)
	}
}

// ----- Check Utilities

func checkPackets(t *testing.T, want core.Packet, got core.Packet) {
	if reflect.DeepEqual(want, got) {
		Ok(t, "Check packets")
		return
	}
	Ko(t, "Converted packet don't match expectations. \nWant: \n%s\nGot:  \n%s", want.String(), got.String())
}

// ---- Build utilities

type convertRXPKTest struct {
	CorePacket core.Packet
	RXPK       semtech.RXPK
	WantError  error
}

// Generates a test suite where the RXPK is fully complete
func genRXPKWithFullMetadata() convertRXPKTest {
	phyPayload := genPHYPayload(true)
	rxpk := genRXPK(phyPayload)
	metadata := genMetadata(rxpk)
	return convertRXPKTest{
		CorePacket: core.Packet{Metadata: &metadata, Payload: phyPayload},
		RXPK:       rxpk,
		WantError:  nil,
	}
}

// Generates a test suite where the RXPK contains partial metadata
func genRXPKWithPartialMetadata() convertRXPKTest {
	phyPayload := genPHYPayload(true)
	rxpk := genRXPK(phyPayload)
	rxpk.Codr = nil
	rxpk.Rfch = nil
	rxpk.Rssi = nil
	rxpk.Time = nil
	rxpk.Size = nil
	metadata := genMetadata(rxpk)
	return convertRXPKTest{
		CorePacket: core.Packet{Metadata: &metadata, Payload: phyPayload},
		RXPK:       rxpk,
		WantError:  nil,
	}
}

// Generates a test suite where the RXPK contains no data
func genRXPKWithNoData() convertRXPKTest {
	rxpk := genRXPK(genPHYPayload(true))
	rxpk.Data = nil
	return convertRXPKTest{
		CorePacket: core.Packet{},
		RXPK:       rxpk,
		WantError:  ErrImpossibleConversion,
	}
}

// Generate an RXPK packet using the given payload as Data
func genRXPK(phyPayload lorawan.PHYPayload) semtech.RXPK {
	raw, err := phyPayload.MarshalBinary()
	if err != nil {
		panic(err)
	}
	data := strings.Trim(base64.StdEncoding.EncodeToString(raw), "=")

	return semtech.RXPK{
		Chan: pointer.Uint(2),
		Codr: pointer.String("4/6"),
		Data: pointer.String(data),
		Freq: pointer.Float64(863.125),
		Lsnr: pointer.Float64(5.2),
		Modu: pointer.String("LORA"),
		Rfch: pointer.Uint(2),
		Rssi: pointer.Int(-27),
		Size: pointer.Uint(uint(len([]byte(data)))),
		Stat: pointer.Int(0),
		Time: pointer.Time(time.Now()),
		Tmst: pointer.Uint(uint(time.Now().UnixNano())),
	}
}

// Generate a Metadata object that matches RXPK metadata
func genMetadata(RXPK semtech.RXPK) Metadata {
	return Metadata{
		Chan: RXPK.Chan,
		Codr: RXPK.Codr,
		Data: RXPK.Data,
		Freq: RXPK.Freq,
		Lsnr: RXPK.Lsnr,
		Modu: RXPK.Modu,
		Rfch: RXPK.Rfch,
		Rssi: RXPK.Rssi,
		Size: RXPK.Size,
		Stat: RXPK.Stat,
		Time: RXPK.Time,
		Tmst: RXPK.Tmst,
	}
}

// Generate a Physical payload represting an uplink or downlink message
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

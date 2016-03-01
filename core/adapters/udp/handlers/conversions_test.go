// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package handlers

import (
	"reflect"
	"testing"

	"github.com/TheThingsNetwork/ttn/core"
	"github.com/TheThingsNetwork/ttn/semtech"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	. "github.com/TheThingsNetwork/ttn/utils/errors/checks"
	"github.com/TheThingsNetwork/ttn/utils/pointer"
	. "github.com/TheThingsNetwork/ttn/utils/testing"
)

type convertRXPKTest struct {
	CorePacket core.Packet
	RXPK       semtech.RXPK
	WantError  *string
}

func TestConvertRXPKPacket(t *testing.T) {
	tests := []convertRXPKTest{
		genRXPKWithFullMetadata(&convertRXPKTest{WantError: nil}),
		genRXPKWithPartialMetadata(&convertRXPKTest{WantError: nil}),
		genRXPKWithNoData(&convertRXPKTest{WantError: pointer.String(string(errors.Structural))}),
	}

	for _, test := range tests {
		Desc(t, "Convert RXPK: %s", pointer.DumpPStruct(test.RXPK, false))
		packet, err := rxpk2packet(test.RXPK, []byte{1, 2, 3, 4, 5, 6, 7, 8})
		CheckErrors(t, test.WantError, err)
		checkPackets(t, test.CorePacket, packet)
	}
}

type convertToTXPKTest struct {
	TXPK       semtech.TXPK
	CorePacket core.RPacket
	WantError  *string
}

func TestConvertTXPKPacket(t *testing.T) {
	tests := []convertToTXPKTest{
		genCoreFullMetadata(&convertToTXPKTest{WantError: nil}),
		genCorePartialMetadata(&convertToTXPKTest{WantError: nil}),
		genCoreExtraMetadata(&convertToTXPKTest{WantError: nil}),
		genCoreNoMetadata(&convertToTXPKTest{WantError: nil}),
	}

	for _, test := range tests {
		Desc(t, "Convert to TXPK: %s", test.CorePacket.String())
		txpk, err := packet2txpk(test.CorePacket)
		CheckErrors(t, test.WantError, err)
		checkTXPKs(t, test.TXPK, txpk)
	}
}

// ----- CHECK utilities

// Checks that obtained TXPK matches expeceted one
func checkTXPKs(t *testing.T, want semtech.TXPK, got semtech.TXPK) {
	if reflect.DeepEqual(want, got) {
		Ok(t, "check TXPKs")
		return
	}
	Ko(t, "Converted TXPK does not match expectations. \nWant: %s\nGot:  %s", pointer.DumpPStruct(want, false), pointer.DumpPStruct(got, false))
}

// Checks that two packets match
func checkPackets(t *testing.T, want core.Packet, got core.Packet) {
	if want == nil {
		if got == nil {
			Ok(t, "Check packets")
			return
		}
		Ko(t, "No packet was expected but got %s", got.String())
		return
	}

	if got == nil {
		Ko(t, "Was expecting %s but got nothing", want.String())
		return
	}

	if reflect.DeepEqual(want, got) {
		Ok(t, "Check packets")
		return
	}

	Ko(t, "Converted packet does not match expectations. \nWant: \n%s\nGot:  \n%s", want.String(), got.String())
}

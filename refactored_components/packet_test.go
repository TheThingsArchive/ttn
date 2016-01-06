// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package components

import (
	"github.com/thethingsnetwork/core"
	"github.com/thethingsnetwork/core/utils/pointer"
	. "github.com/thethingsnetwork/core/utils/testing"
	"testing"
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

// ---- Build utilities

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

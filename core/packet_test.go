// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package core

import (
	"encoding/json"
	"testing"

	"github.com/TheThingsNetwork/ttn/semtech"
	"github.com/TheThingsNetwork/ttn/utils/pointer"
	. "github.com/TheThingsNetwork/ttn/utils/testing"
	"github.com/brocaar/lorawan"
)

func TestConvertRXPKPacket(t *testing.T) {
	tests := []convertRXPKTest{
		genRXPKWithFullMetadata(&convertRXPKTest{WantError: nil}),
		genRXPKWithPartialMetadata(&convertRXPKTest{WantError: nil}),
		genRXPKWithNoData(&convertRXPKTest{WantError: pointer.String(ErrInvalidPacket)}),
	}

	for _, test := range tests {
		Desc(t, "Convert RXPK: %s", pointer.DumpPStruct(test.RXPK, false))
		packet, err := ConvertRXPK(test.RXPK)
		checkErrors(t, test.WantError, err)
		checkPackets(t, test.CorePacket, packet)
	}
}

func TestConvertTXPKPacket(t *testing.T) {
	tests := []convertToTXPKTest{
		genCoreFullMetadata(&convertToTXPKTest{WantError: nil}),
		genCorePartialMetadata(&convertToTXPKTest{WantError: nil}),
		genCoreExtraMetadata(&convertToTXPKTest{WantError: nil}),
		genCoreNoMetadata(&convertToTXPKTest{WantError: nil}),
		convertToTXPKTest{
			CorePacket: Packet{Metadata: genFullMetadata(), Payload: lorawan.PHYPayload{}},
			TXPK:       semtech.TXPK{},
			WantError:  pointer.String(ErrInvalidPacket),
		},
	}

	for _, test := range tests {
		Desc(t, "Convert to TXPK: %s", test.CorePacket.String())
		txpk, err := ConvertToTXPK(test.CorePacket)
		checkErrors(t, test.WantError, err)
		checkTXPKs(t, test.TXPK, txpk)
	}
}

func TestMarshalJSONPacket(t *testing.T) {
	tests := []marshalJSONTest{
		marshalJSONTest{ // Empty Payload
			Packet:     Packet{Metadata: genFullMetadata(), Payload: lorawan.PHYPayload{}},
			WantFields: []string{},
		},
		marshalJSONTest{ // Empty Metadata
			Packet:     Packet{Metadata: Metadata{}, Payload: genPHYPayload(true)},
			WantFields: []string{"payload", "metadata"},
		},
		marshalJSONTest{ // With Metadata and Payload
			Packet:     Packet{Metadata: genFullMetadata(), Payload: genPHYPayload(true)},
			WantFields: []string{"payload", "metadata"},
		},
	}

	for _, test := range tests {
		Desc(t, "Marshal packet to json: %s", test.Packet.String())
		raw, _ := json.Marshal(test.Packet)
		checkFields(t, test.WantFields, raw)
	}
}

func TestUnmarshalJSONPacket(t *testing.T) {
	tests := []unmarshalJSONTest{
		unmarshalJSONTest{
			JSON:       `{"payload":"gAQDAgEAAAAK4mTU97VqDnU=","metadata":{}}`,
			WantPacket: Packet{Metadata: Metadata{}, Payload: genPHYPayload(true)},
		},
		unmarshalJSONTest{
			JSON:       `{"payload":"gAQDAgEAAAAK4mTU97VqDnU=","metadata":{"chan":2,"codr":"4/6","fdev":3,"freq":863.125,"imme":false,"ipol":false,"lsnr":5.2,"modu":"LORA","ncrc":true,"powe":3,"prea":8,"rfch":2,"rssi":-27,"size":14,"stat":0,"tmst":1452694288207288421,"datr":"LORA","time":"2016-01-13T14:11:28.207288421Z"}}`,
			WantPacket: Packet{Metadata: genFullMetadata(), Payload: genPHYPayload(true)},
		},
		unmarshalJSONTest{
			JSON:       `invalid`,
			WantPacket: Packet{},
		},
		unmarshalJSONTest{
			JSON:       `{"metadata":{}}`,
			WantPacket: Packet{},
		},
	}

	for _, test := range tests {
		Desc(t, "Unmarshal json to packet: %s", test.JSON)
		var packet Packet
		json.Unmarshal([]byte(test.JSON), &packet)
		checkPackets(t, test.WantPacket, packet)
	}
}

// ---- Declaration
type convertRXPKTest struct {
	CorePacket Packet
	RXPK       semtech.RXPK
	WantError  *string
}

type convertToTXPKTest struct {
	TXPK       semtech.TXPK
	CorePacket Packet
	WantError  *string
}

type marshalJSONTest struct {
	Packet     Packet
	WantFields []string
}

type unmarshalJSONTest struct {
	JSON       string
	WantPacket Packet
}

// ---- Build utilities

// Generates a test suite where the RXPK is fully complete
func genRXPKWithFullMetadata(test *convertRXPKTest) convertRXPKTest {
	phyPayload := genPHYPayload(true)
	rxpk := genRXPK(phyPayload)
	metadata := genMetadata(rxpk)
	test.CorePacket = Packet{Metadata: metadata, Payload: phyPayload}
	test.RXPK = rxpk
	return *test
}

// Generates a test suite where the RXPK contains partial metadata
func genRXPKWithPartialMetadata(test *convertRXPKTest) convertRXPKTest {
	phyPayload := genPHYPayload(true)
	rxpk := genRXPK(phyPayload)
	rxpk.Codr = nil
	rxpk.Rfch = nil
	rxpk.Rssi = nil
	rxpk.Time = nil
	rxpk.Size = nil
	metadata := genMetadata(rxpk)
	test.CorePacket = Packet{Metadata: metadata, Payload: phyPayload}
	test.RXPK = rxpk
	return *test
}

// Generates a test suite where the RXPK contains no data
func genRXPKWithNoData(test *convertRXPKTest) convertRXPKTest {
	rxpk := genRXPK(genPHYPayload(true))
	rxpk.Data = nil
	test.RXPK = rxpk
	return *test
}

// Generates a test suite where the core packet has all txpk metadata
func genCoreFullMetadata(test *convertToTXPKTest) convertToTXPKTest {
	phyPayload := genPHYPayload(false)
	metadata := genFullMetadata()
	metadata.Chan = nil
	metadata.Lsnr = nil
	metadata.Rssi = nil
	metadata.Stat = nil
	test.TXPK = genTXPK(phyPayload, metadata)
	test.CorePacket = Packet{Metadata: metadata, Payload: phyPayload}
	return *test
}

// Generates a test suite where the core packet has no metadata
func genCoreNoMetadata(test *convertToTXPKTest) convertToTXPKTest {
	phyPayload := genPHYPayload(false)
	metadata := Metadata{}
	test.TXPK = genTXPK(phyPayload, metadata)
	test.CorePacket = Packet{Metadata: metadata, Payload: phyPayload}
	return *test
}

// Generates a test suite where the core packet has partial metadata but all supported
func genCorePartialMetadata(test *convertToTXPKTest) convertToTXPKTest {
	phyPayload := genPHYPayload(false)
	metadata := genFullMetadata()
	metadata.Chan = nil
	metadata.Lsnr = nil
	metadata.Rssi = nil
	metadata.Stat = nil
	metadata.Modu = nil
	metadata.Fdev = nil
	metadata.Time = nil
	test.TXPK = genTXPK(phyPayload, metadata)
	test.CorePacket = Packet{Metadata: metadata, Payload: phyPayload}
	return *test
}

// Generates a test suite where the core packet has extra metadata not supported by txpk
func genCoreExtraMetadata(test *convertToTXPKTest) convertToTXPKTest {
	phyPayload := genPHYPayload(false)
	metadata := genFullMetadata()
	test.TXPK = genTXPK(phyPayload, metadata)
	test.CorePacket = Packet{Metadata: metadata, Payload: phyPayload}
	return *test
}

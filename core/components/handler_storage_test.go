// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package components

import (
	"reflect"
	"testing"
	"time"

	"github.com/TheThingsNetwork/ttn/core"
	. "github.com/TheThingsNetwork/ttn/core/errors"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	"github.com/TheThingsNetwork/ttn/utils/pointer"
	. "github.com/TheThingsNetwork/ttn/utils/testing"
	"github.com/brocaar/lorawan"
)

func TestStoragePartition(t *testing.T) {
	// CONVENTION below -> first DevAddr byte will be used as falue for FPort
	setup := []handlerEntry{
		{ // App #1, Dev #1
			AppEUI:  lorawan.EUI64([8]byte{0, 0, 0, 0, 0, 0, 0, 1}),
			NwkSKey: lorawan.AES128Key([16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}),
			AppSKey: lorawan.AES128Key([16]byte{16, 15, 14, 13, 12, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1}),
			DevAddr: lorawan.DevAddr([4]byte{0, 0, 0, 1}),
		},
		{ // App #1, Dev #2
			AppEUI:  lorawan.EUI64([8]byte{0, 0, 0, 0, 0, 0, 0, 1}),
			NwkSKey: lorawan.AES128Key([16]byte{0, 0xa, 0xb, 1, 2, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}),
			AppSKey: lorawan.AES128Key([16]byte{14, 14, 14, 14, 12, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1}),
			DevAddr: lorawan.DevAddr([4]byte{10, 0, 0, 2}),
		},
		{ // App #1, Dev #3
			AppEUI:  lorawan.EUI64([8]byte{0, 0, 0, 0, 0, 0, 0, 1}),
			NwkSKey: lorawan.AES128Key([16]byte{12, 0xa, 0xb, 1, 2, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}),
			AppSKey: lorawan.AES128Key([16]byte{0xb, 15, 14, 14, 12, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1}),
			DevAddr: lorawan.DevAddr([4]byte{14, 0, 0, 3}),
		},
		{ // App #2, Dev #1
			AppEUI:  lorawan.EUI64([8]byte{0, 0, 0, 0, 0, 0, 0, 2}),
			NwkSKey: lorawan.AES128Key([16]byte{0, 0xa, 0xb, 5, 12, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}),
			AppSKey: lorawan.AES128Key([16]byte{14, 14, 14, 14, 1, 11, 10, 0xc, 8, 7, 6, 5, 4, 3, 2, 1}),
			DevAddr: lorawan.DevAddr([4]byte{0, 0, 0, 1}),
		},
		{ // App #2, Dev #2
			AppEUI:  lorawan.EUI64([8]byte{0, 0, 0, 0, 0, 0, 0, 2}),
			NwkSKey: lorawan.AES128Key([16]byte{0, 0xa, 0xb, 5, 12, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}),
			AppSKey: lorawan.AES128Key([16]byte{14, 14, 14, 14, 1, 11, 10, 0xc, 8, 7, 6, 5, 4, 3, 2, 1}),
			DevAddr: lorawan.DevAddr([4]byte{23, 0xaf, 0x14, 1}),
		},
	}

	unknown := handlerEntry{ // App #1, Dev #4
		AppEUI:  lorawan.EUI64([8]byte{0, 0, 0, 0, 0, 0, 0, 1}),
		NwkSKey: lorawan.AES128Key([16]byte{1, 2, 3, 4, 23, 6, 7, 8, 9, 0x19, 11, 12, 13, 14, 15, 16}),
		AppSKey: lorawan.AES128Key([16]byte{16, 0xba, 14, 13, 2, 11, 58, 9, 8, 7, 6, 5, 4, 3, 2, 1}),
		DevAddr: lorawan.DevAddr([4]byte{1, 0, 0, 4}),
	}

	tests := []struct {
		Desc           string
		PacketsShape   []handlerEntry
		WantPartitions []partitionShape
		WantError      *string
	}{
		{
			Desc:           "1 packet -> 1 partition | 1 packet",
			PacketsShape:   []handlerEntry{setup[0]},
			WantPartitions: []partitionShape{{setup[0], 1}},
			WantError:      nil,
		},
		{
			Desc:           "1 unknown packet -> error not found",
			PacketsShape:   []handlerEntry{unknown},
			WantPartitions: nil,
			WantError:      pointer.String(ErrNotFound),
		},
		{
			Desc:           "2 packets | diff DevAddr & diff AppEUI -> 2 partitions | 1 packet",
			PacketsShape:   []handlerEntry{setup[0], setup[4]},
			WantPartitions: []partitionShape{{setup[0], 1}, {setup[4], 1}},
			WantError:      nil,
		},
		{
			Desc:           "2 packets | same DevAddr & diff AppEUI -> 2 partitions | 1 packet",
			PacketsShape:   []handlerEntry{setup[0], setup[3]},
			WantPartitions: []partitionShape{{setup[0], 1}, {setup[3], 1}},
			WantError:      nil,
		},
		{
			Desc:           "3 packets | diff DevAddr & same AppEUI -> 3 partitions | 1 packet",
			PacketsShape:   []handlerEntry{setup[0], setup[1], setup[2]},
			WantPartitions: []partitionShape{{setup[0], 1}, {setup[1], 1}, {setup[2], 1}},
			WantError:      nil,
		},
		{
			Desc:           "3 packets | same DevAddr & same AppEUI -> 1 partitions | 3 packets",
			PacketsShape:   []handlerEntry{setup[0], setup[0], setup[0]},
			WantPartitions: []partitionShape{{setup[0], 3}},
			WantError:      nil,
		},
		{
			Desc:           "5 packets | same DevAddr & various AppEUI -> 2 partitions | 3 packets & 2 packets",
			PacketsShape:   []handlerEntry{setup[0], setup[0], setup[0], setup[3], setup[3]},
			WantPartitions: []partitionShape{{setup[0], 3}, {setup[3], 2}},
			WantError:      nil,
		},
	}

	for _, test := range tests {
		// Describe
		Desc(t, test.Desc)

		// Build
		db := genFilledHandlerStorage(setup)
		packets := genPacketsFromHandlerEntries(test.PacketsShape)

		// Operate
		partitions, err := db.Partition(packets...)

		// Check
		checkErrors(t, test.WantError, err)
		checkPartitions(t, test.WantPartitions, partitions)
		if err := db.Close(); err != nil {
			panic(err)
		}
	}
}

type partitionShape struct {
	handlerEntry
	PacketNb int
}

// ----- BUILD utilities

func genFilledHandlerStorage(setup []handlerEntry) HandlerStorage {
	db, err := NewHandlerStorage()
	if err != nil {
		panic(err)
	}

	if err := db.Reset(); err != nil {
		panic(err)
	}

	for _, entry := range setup {
		if err := db.Store(entry.DevAddr, entry); err != nil {
			panic(err)
		}
	}

	return db
}

func genPacketsFromHandlerEntries(shapes []handlerEntry) []core.Packet {
	var packets []core.Packet
	for _, entry := range shapes {

		// Build the macPayload
		macPayload := lorawan.NewMACPayload(true)
		macPayload.FHDR = lorawan.FHDR{DevAddr: entry.DevAddr}
		macPayload.FRMPayload = []lorawan.Payload{&lorawan.DataPayload{
			Bytes: []byte(time.Now().String()),
		}}
		macPayload.FPort = uint8(entry.DevAddr[0])
		key := entry.AppSKey
		if macPayload.FPort == 0 {
			key = entry.NwkSKey
		}
		if err := macPayload.EncryptFRMPayload(key); err != nil {
			panic(err)
		}

		// Build the physicalPayload
		phyPayload := lorawan.NewPHYPayload(true)
		phyPayload.MHDR = lorawan.MHDR{
			MType: lorawan.ConfirmedDataUp,
			Major: lorawan.LoRaWANR1,
		}
		phyPayload.MACPayload = macPayload
		if err := phyPayload.SetMIC(entry.NwkSKey); err != nil {
			panic(err)
		}

		// Finally build the packet
		packets = append(packets, core.Packet{
			Metadata: core.Metadata{
				Rssi: pointer.Int(-20),
				Datr: pointer.String("SF7BW125"),
				Modu: pointer.String("Lora"),
			},
			Payload: phyPayload,
		})
	}
	return packets
}

// ----- CHECK utilities
func checkErrors(t *testing.T, want *string, got error) {
	if want == nil && got == nil || got.(errors.Failure).Nature == *want {
		Ok(t, "Check errors")
		return
	}

	Ko(t, "Expected error to be %s but got %v", want, got)
}

func checkPartitions(t *testing.T, want []partitionShape, got []handlerPartition) {
	if len(want) != len(got) {
		Ko(t, "Expected %d partitions, but got %d", len(want), len(got))
		return
	}

browseGot:
	for _, gotPartition := range got {
		for _, wantPartition := range want { // Find right wanted partition
			if reflect.DeepEqual(wantPartition.handlerEntry, gotPartition.handlerEntry) {
				if len(gotPartition.Packets) == wantPartition.PacketNb {
					continue browseGot
				}
				Ko(t, "Partition don't match expectations.\nWant: %v\nGot:  %v", wantPartition, gotPartition)
				return
			}
		}
		Ko(t, "Got a partition that wasn't expected: %v", gotPartition)
		return
	}
	Ok(t, "Check partitions")
}

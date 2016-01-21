// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package components

import (
	"github.com/TheThingsNetwork/ttn/core"
	. "github.com/TheThingsNetwork/ttn/utils/testing"
	"github.com/brocaar/lorawan"
	"testing"
)

func TestStoragePartition(t *testing.T) {
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
			DevAddr: lorawan.DevAddr([4]byte{0, 0, 0, 2}),
		},
		{ // App #1, Dev #3
			AppEUI:  lorawan.EUI64([8]byte{0, 0, 0, 0, 0, 0, 0, 1}),
			NwkSKey: lorawan.AES128Key([16]byte{12, 0xa, 0xb, 1, 2, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}),
			AppSKey: lorawan.AES128Key([16]byte{0xb, 15, 14, 14, 12, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1}),
			DevAddr: lorawan.DevAddr([4]byte{0, 0, 0, 3}),
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
			DevAddr: lorawan.DevAddr([4]byte{0, 0xaf, 0x14, 1}),
		},
	}

	tests := []struct {
		Desc           string
		PacketsShape   []handlerEntry
		WantPartitions []partitionShape
		WantError      error
	}{
		{
			Desc:           "",
			PacketsShape:   []handlerEntry{setup[0]},
			WantPartitions: []partitionShape{{setup[0], 1}},
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
		partitions, err := db.partition(packets)

		// Check
		checkErrors(t, test.WantError, err)
		checkPartitions(t, test.WantPartitions, partitions)
	}
}

type partitionShape struct {
	Entry    handlerEntry
	PacketNb int
}

// ----- BUILD utilities

func genFilledHandlerStorage(setup []handlerEntry) (db handlerStorage) {
	db = newHandlerDB()

	for _, entry := range setup {
		if err := db.store(entry.DevAddr, entry); err != nil {
			panic(err)
		}
	}

	return db
}

func genPacketsFromHandlerEntries(shapes []handlerEntry) []core.Packet {
	return nil
}

// ----- CHECK utilities

func checkErrors(t *testing.T, want error, got error) {

}

func checkPartitions(t *testing.T, want []partitionShape, got []handlerPartition) {

}

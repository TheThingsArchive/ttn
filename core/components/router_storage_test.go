// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package components

import (
	"testing"
	"time"

	"github.com/TheThingsNetwork/ttn/core"
	. "github.com/TheThingsNetwork/ttn/utils/testing"
	"github.com/brocaar/lorawan"
)

type routerEntryShape struct {
	Until   time.Time
	DevAddr lorawan.DevAddr
	Address string
}

func TestStorageExpiration(t *testing.T) {
	tests := []struct {
		Desc            string
		ExistingEntries []routerEntryShape
		Store           *routerEntryShape
		Lookup          lorawan.DevAddr
		WantEntry       *routerEntryShape
		WantError       []error
	}{
		{
			Desc:            "No entry, Lookup address",
			ExistingEntries: nil,
			Store:           nil,
			Lookup:          lorawan.DevAddr([4]byte{0, 0, 0, 1}),
			WantEntry:       nil,
			WantError:       []error{ErrNotFound},
		},
	}

	for _, test := range tests {
		// Describe
		Desc(t, test.Desc)

		// Build
		db := genFilledRouterStorage(test.ExistingEntries)
		cherr := make(chan interface{}, 2)

		// Operate
		storeRouter(db, test.Store, cherr)
		got := lookupRouter(db, test.Lookup, cherr)

		// Check
		checkChErrors(t, test.WantError, cherr)
		checkRouterEntries(t, test.WantEntry, got)
	}
}

// ----- BUILD utilities
func genFilledRouterStorage(setup []routerEntryShape) RouterStorage {
	db, err := NewRouterStorage()
	if err != nil {
		panic(err)
	}

	if err := db.Reset(); err != nil {
		panic(err)
	}

	for i, shape := range setup {
		entry := routerEntry{
			until: shape.Until,
			Recipient: core.Recipient{
				Address: shape.Address,
				Id:      i,
			},
		}
		if err := db.Store(shape.DevAddr, entry); err != nil {
			panic(err)
		}
	}

	return db
}

// ----- OPERATE utilities
func storeRouter(db RouterStorage, entry *routerEntryShape, cherr chan interface{}) {

}

func lookupRouter(db RouterStorage, devAddr lorawan.DevAddr, cherr chan interface{}) routerEntry {
	return routerEntry{}
}

// ----- CHECK utilities
func checkRouterEntries(t *testing.T, want *routerEntryShape, got routerEntry) {
}

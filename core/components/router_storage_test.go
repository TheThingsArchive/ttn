// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package components

import (
	"reflect"
	"testing"
	"time"

	"github.com/TheThingsNetwork/ttn/core"
	. "github.com/TheThingsNetwork/ttn/utils/testing"
	"github.com/brocaar/lorawan"
)

type routerEntryShape struct {
	DevAddr lorawan.DevAddr
	Address string
}

func TestStorageExpiration(t *testing.T) {
	tests := []struct {
		Desc            string
		ExistingEntries []routerEntryShape
		ExpiryDelay     time.Duration
		Store           *routerEntryShape
		Lookup          lorawan.DevAddr
		WantEntry       *routerEntryShape
		WantError       []error
	}{
		{
			Desc:            "No entry, Lookup address",
			ExistingEntries: nil,
			ExpiryDelay:     time.Minute,
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
		go func() {
			time.After(time.Millisecond * 250)
			close(cherr)
		}()

		// Operate
		storeRouter(db, test.Store, cherr)
		got := lookupRouter(db, test.Lookup, cherr)

		// Check
		checkChErrors(t, test.WantError, cherr)
		checkRouterEntries(t, test.WantEntry, got)

		// Clean
		db.Close()
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
	if entry == nil {
		return
	}

	err := db.Store(entry.DevAddr, routerEntry{
		Recipient: core.Recipient{
			Address: entry.Address,
			Id:      "LikeICare",
		},
	})

	if err != nil {
		cherr <- err
	}
}

func lookupRouter(db RouterStorage, devAddr lorawan.DevAddr, cherr chan interface{}) routerEntry {
	entry, err := db.Lookup(devAddr)
	if err != nil {
		cherr <- err
	}
	return entry
}

// ----- CHECK utilities
func checkRouterEntries(t *testing.T, want *routerEntryShape, got routerEntry) {
	if want != nil {
		addr, ok := got.Recipient.Address.(string)

		if !ok {
			Ko(t, "Unexpected recipient address format: %+v", got.Recipient.Address)
			return
		}

		if addr != want.Address {
			Ko(t, "The retrieved address [%s] does not match expected [%s]", addr, want.Address)
			return
		}
	} else {
		if !reflect.DeepEqual(routerEntry{}, got) {
			Ko(t, "No entry was exected but got: %v", got)
			return
		}
	}
	Ok(t, "Check router entries")
}

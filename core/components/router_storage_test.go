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
	routerEntry
	DevAddr lorawan.DevAddr
}

func TestStorageExpiration(t *testing.T) {
	devices := []lorawan.DevAddr{
		lorawan.DevAddr([4]byte{0, 0, 0, 1}),
	}

	entries := []routerEntry{
		{Recipient: core.Recipient{Address: "MyAddress1", Id: ""}},
	}

	tests := []struct {
		Desc            string
		ExistingEntries []routerEntryShape
		ExpiryDelay     time.Duration
		Store           *routerEntryShape
		Lookup          lorawan.DevAddr
		WantEntry       *routerEntry
		WantError       []error
	}{
		{
			Desc:            "No entry, Lookup address",
			ExistingEntries: nil,
			ExpiryDelay:     time.Minute,
			Lookup:          devices[0],
			Store:           nil,
			WantEntry:       nil,
			WantError:       []error{ErrNotFound},
		},
		{
			Desc:            "No entry, Store and Lookup same",
			ExistingEntries: nil,
			ExpiryDelay:     time.Minute,
			Store:           &routerEntryShape{entries[0], devices[0]},
			Lookup:          devices[0],
			WantEntry:       &entries[0],
			WantError:       nil,
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

	for _, shape := range setup {
		if err := db.Store(shape.DevAddr, shape.routerEntry); err != nil {
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

	if err := db.Store(entry.DevAddr, entry.routerEntry); err != nil {
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
func checkRouterEntries(t *testing.T, want *routerEntry, got routerEntry) {
	if want != nil {
		addr, ok := got.Recipient.Address.(string)

		if !ok {
			Ko(t, "Unexpected recipient address format: %+v", got.Recipient.Address)
			return
		}

		if addr != want.Recipient.Address.(string) {
			Ko(t, "The retrieved address [%s] does not match expected [%s]", addr, want.Recipient.Address)
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

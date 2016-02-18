// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package components

import (
	"reflect"
	"testing"
	"time"

	"github.com/TheThingsNetwork/ttn/core"
	. "github.com/TheThingsNetwork/ttn/core/errors"
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
		lorawan.DevAddr([4]byte{14, 15, 8, 42}),
		lorawan.DevAddr([4]byte{14, 15, 8, 79}),
	}

	entries := []routerEntry{
		{Recipient: core.Recipient{Address: "MyAddress1", Id: ""}},
		{Recipient: core.Recipient{Address: "AnotherAddress", Id: ""}},
	}

	tests := []struct {
		Desc            string
		ExpiryDelay     time.Duration
		ExistingEntries []routerEntryShape
		WaitDelayS      time.Duration
		Store           *routerEntryShape
		WaitDelayL      time.Duration
		Lookup          lorawan.DevAddr
		WantEntry       *routerEntry
		WantError       []string
	}{
		{
			Desc:            "No entry, Lookup address",
			ExpiryDelay:     time.Minute,
			ExistingEntries: nil,
			Lookup:          devices[0],
			Store:           nil,
			WantEntry:       nil,
			WantError:       []string{ErrNotFound},
		},
		{
			Desc:            "No entry, Store and Lookup same",
			ExpiryDelay:     time.Minute,
			ExistingEntries: nil,
			Store:           &routerEntryShape{entries[0], devices[0]},
			Lookup:          devices[0],
			WantEntry:       &entries[0],
			WantError:       nil,
		},
		{
			Desc:            "No entry, store, wait expiry, and lookup same",
			ExpiryDelay:     time.Millisecond,
			ExistingEntries: nil,
			Store:           &routerEntryShape{entries[0], devices[0]},
			WaitDelayL:      time.Millisecond * 250,
			Lookup:          devices[0],
			WantEntry:       nil,
			WantError:       []string{ErrNotFound},
		},
		{
			Desc:        "One entry, store same, lookup same",
			ExpiryDelay: time.Minute,
			ExistingEntries: []routerEntryShape{
				{entries[0], devices[2]},
			},
			Store:     &routerEntryShape{entries[1], devices[2]},
			Lookup:    devices[2],
			WantEntry: &entries[0],
			WantError: []string{ErrFailedOperation},
		},
		{
			Desc:        "One entry, store different, lookup newly stored",
			ExpiryDelay: time.Minute,
			ExistingEntries: []routerEntryShape{
				{entries[0], devices[0]},
			},
			Store:     &routerEntryShape{entries[1], devices[1]},
			Lookup:    devices[1],
			WantEntry: &entries[1],
			WantError: nil,
		},
		{
			Desc:        "One entry, store different, lookup first one",
			ExpiryDelay: time.Minute,
			ExistingEntries: []routerEntryShape{
				{entries[0], devices[0]},
			},
			Store:     &routerEntryShape{entries[1], devices[1]},
			Lookup:    devices[0],
			WantEntry: &entries[0],
		},
		{
			Desc:        "One entry, store different, wait delay, lookup first one",
			ExpiryDelay: time.Millisecond,
			ExistingEntries: []routerEntryShape{
				{entries[0], devices[0]},
			},
			Store:      &routerEntryShape{entries[1], devices[1]},
			WaitDelayL: time.Millisecond,
			Lookup:     devices[0],
			WantEntry:  nil,
			WantError:  []string{ErrNotFound},
		},
		{
			Desc:        "One entry, wait delay, store same, lookup same",
			ExpiryDelay: time.Millisecond * 100,
			ExistingEntries: []routerEntryShape{
				{entries[0], devices[0]},
			},
			WaitDelayS: time.Millisecond * 200,
			Store:      &routerEntryShape{entries[1], devices[0]},
			Lookup:     devices[0],
			WantEntry:  &entries[1],
			WantError:  nil,
		},
	}

	for _, test := range tests {
		// Describe
		Desc(t, test.Desc)

		// Build
		db := genFilledRouterStorage(test.ExistingEntries, test.ExpiryDelay)
		cherr := make(chan interface{}, 2)
		// Operate
		storeRouter(db, test.WaitDelayS, test.Store, cherr)
		got := lookupRouter(db, test.WaitDelayL, test.Lookup, cherr)

		// Check
		go func() {
			time.After(time.Millisecond * 250)
			close(cherr)
		}()
		checkChErrors(t, test.WantError, cherr)
		checkRouterEntries(t, test.WantEntry, got)

		// Clean
		db.Close()
	}
}

// ----- BUILD utilities
func genFilledRouterStorage(setup []routerEntryShape, expiryDelay time.Duration) RouterStorage {
	db, err := NewRouterStorage(expiryDelay)
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
func storeRouter(db RouterStorage, delay time.Duration, entry *routerEntryShape, cherr chan interface{}) {
	if delay != 0 {
		<-time.After(delay)
	}
	if entry == nil {
		return
	}

	if err := db.Store(entry.DevAddr, entry.routerEntry); err != nil {
		cherr <- err
	}
}

func lookupRouter(db RouterStorage, delay time.Duration, devAddr lorawan.DevAddr, cherr chan interface{}) routerEntry {
	if delay != 0 {
		<-time.After(delay)
	}
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
			Ko(t, `The retrieved address "%s" does not match expected "%s"`, addr, want.Recipient.Address)
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

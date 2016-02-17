// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package components

import (
	"reflect"
	"testing"
	"time"

	. "github.com/TheThingsNetwork/ttn/utils/testing"
	"github.com/brocaar/lorawan"
)

type brokerEntryShape struct {
	brokerEntry
	DevAddr lorawan.DevAddr
}

func TestBrokerStorage(t *testing.T) {
	devices := []lorawan.DevAddr{
		lorawan.DevAddr([4]byte{0, 0, 0, 1}),
		lorawan.DevAddr([4]byte{14, 15, 8, 42}),
	}

	tests := []struct {
		Desc            string
		ExistingEntries []brokerEntryShape
		Store           *brokerEntryShape
		Lookup          lorawan.DevAddr
		WantEntries     []brokerEntry
		WantError       []error
	}{
		{
			Desc:            "Default",
			ExistingEntries: nil,
			Store:           nil,
			Lookup:          devices[0],
			WantEntries:     nil,
			WantError:       nil,
		},
	}

	for _, test := range tests {
		// Describe
		Desc(t, test.Desc)

		// Build
		db := genFilledBrokerStorage(test.ExistingEntries)
		cherr := make(chan interface{}, 2)

		// Operate
		storeBroker(db, test.Store, cherr)
		got := lookupBroker(db, test.Lookup, cherr)

		// Check
		go func() {
			time.After(time.Millisecond * 250)
			close(cherr)
		}()
		checkChErrors(t, test.WantError, cherr)
		checkBrokerEntries(t, test.WantEntries, got)

		// Clean
		db.Close()
	}
}

// ----- BUILD utilities
func genFilledBrokerStorage(setup []brokerEntryShape) BrokerStorage {
	db, err := NewBrokerStorage()
	if err != nil {
		panic(err)
	}

	if err := db.Reset(); err != nil {
		panic(err)
	}

	for _, shape := range setup {
		if err := db.Store(shape.DevAddr, shape.brokerEntry); err != nil {
			panic(err)
		}
	}

	return db
}

// ----- OPERATE utilities
func storeBroker(db BrokerStorage, entry *brokerEntryShape, cherr chan interface{}) {
	if entry == nil {
		return
	}

	if err := db.Store(entry.DevAddr, entry.brokerEntry); err != nil {
		cherr <- err
	}
}

func lookupBroker(db BrokerStorage, devAddr lorawan.DevAddr, cherr chan interface{}) []brokerEntry {
	entry, err := db.Lookup(devAddr)
	if err != nil {
		cherr <- err
	}
	return entry
}

// ----- CHECK utilities
func checkBrokerEntries(t *testing.T, want []brokerEntry, got []brokerEntry) {
	if len(want) != len(got) {
		Ko(t, "Expecting %d entries but got %d.", len(want), len(got))
		return
	}

outer:
	for _, gentry := range got {
		for _, wentry := range want {
			if reflect.DeepEqual(gentry, wentry) {
				continue outer
			}
		}
		Ko(t, "Got an unexpected entry: %+v\nExpected only: %+v", gentry, want)
		return
	}

	Ok(t, "Check broker entries")
}

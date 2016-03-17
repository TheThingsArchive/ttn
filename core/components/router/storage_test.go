// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package router

import (
	"os"
	"path"
	"testing"
	"time"

	"github.com/TheThingsNetwork/ttn/core"
	. "github.com/TheThingsNetwork/ttn/utils/testing"
)

const storageDB = "TestRouterStorage.db"

func CheckEntries(t *testing.T, want []entry, got []entry) {
	for i, w := range want {
		if i >= len(got) {
			Ko(t, "Didn't got enough entries: %v", got)
		}
		tmin := w.until.Add(-time.Second)
		tmax := w.until.Add(time.Second)
		if !tmin.Before(got[i].until) || !got[i].until.Before(tmax) {
			Ko(t, "Unexpected expiry time.\nWant: %s\nGot:  %s", w.until, got[i].until)
		}
		Check(t, w.BrokerIndex, got[i].BrokerIndex, "Brokers")
	}
}

func TestStoreAndLookup(t *testing.T) {
	storageDB := path.Join(os.TempDir(), storageDB)

	defer func() {
		os.Remove(storageDB)
	}()

	// ------------------

	{
		Desc(t, "Create a new storage")
		db, err := NewStorage(storageDB, time.Hour)
		CheckErrors(t, nil, err)
		err = db.Close()
		CheckErrors(t, nil, err)
	}

	// ------------------

	{
		Desc(t, "Store then lookup a device")

		// Build
		db, _ := NewStorage(storageDB, time.Hour)
		devAddr := []byte{0, 0, 0, 1}

		// Operate
		err := db.Store(devAddr, 1)
		FatalUnless(t, err)
		gotEntry, err := db.Lookup(devAddr)

		// Expectations
		wantEntry := []entry{
			{
				BrokerIndex: 1,
				until:       time.Now().Add(time.Hour),
			},
		}

		// Check
		CheckErrors(t, nil, err)
		CheckEntries(t, wantEntry, gotEntry)
		_ = db.Close()
	}

	// ------------------

	{
		Desc(t, "Lookup non-existing entry")

		// Build
		db, _ := NewStorage(storageDB, time.Hour)
		devAddr := []byte{0, 0, 0, 2}

		// Operate
		gotEntry, err := db.Lookup(devAddr)

		// Checks
		CheckErrors(t, ErrNotFound, err)
		CheckEntries(t, nil, gotEntry)
		_ = db.Close()
	}

	// ------------------

	{
		Desc(t, "Lookup an expired entry")

		// Build
		db, _ := NewStorage(storageDB, time.Millisecond*100)
		devAddr := []byte{0, 0, 0, 3}

		// Operate
		_ = db.Store(devAddr, 1)
		<-time.After(time.Millisecond * 200)
		gotEntry, err := db.Lookup(devAddr)

		// Checks
		CheckErrors(t, ErrNotFound, err)
		CheckEntries(t, nil, gotEntry)
		_ = db.Close()
	}

	// ------------------

	{
		Desc(t, "Store above an expired entry")

		// Build
		db, _ := NewStorage(storageDB, time.Millisecond*100)
		devAddr := []byte{0, 0, 0, 4}

		// Operate
		_ = db.Store(devAddr, 1)
		<-time.After(time.Millisecond * 200)
		err := db.Store(devAddr, 2)
		FatalUnless(t, err)
		gotEntry, err := db.Lookup(devAddr)

		// Expectations
		wantEntry := []entry{
			{
				BrokerIndex: 2,
				until:       time.Now().Add(time.Millisecond * 100),
			},
		}

		// Checks
		CheckErrors(t, nil, err)
		CheckEntries(t, wantEntry, gotEntry)
		_ = db.Close()
	}

	// ------------------

	{
		Desc(t, "Store on a closed database")

		// Build
		db, _ := NewStorage(storageDB, time.Hour)
		_ = db.Close()
		devAddr := []byte{0, 0, 0, 5}

		// Operate
		err := db.Store(devAddr, 1)

		// Checks
		CheckErrors(t, ErrOperational, err)
	}

	// ------------------

	{
		Desc(t, "Lookup on a closed database")

		// Build
		db, _ := NewStorage(storageDB, time.Hour)
		_ = db.Close()
		devAddr := []byte{0, 0, 0, 1}

		// Operate
		gotEntry, err := db.Lookup(devAddr)

		// Checks
		CheckErrors(t, ErrOperational, err)
		CheckEntries(t, nil, gotEntry)
	}

	// ------------------

	{
		Desc(t, "Store two entries in a row")

		// Build
		db, _ := NewStorage(storageDB, time.Hour)
		devAddr := []byte{0, 0, 0, 6}

		// Operate
		err := db.Store(devAddr, 1)
		FatalUnless(t, err)
		err = db.Store(devAddr, 2)
		FatalUnless(t, err)
		gotEntries, err := db.Lookup(devAddr)
		FatalUnless(t, err)

		// Expectations
		wantEntries := []entry{
			{
				BrokerIndex: 1,
				until:       time.Now().Add(time.Hour),
			},
			{
				BrokerIndex: 2,
				until:       time.Now().Add(time.Hour),
			},
		}

		// Check
		CheckEntries(t, wantEntries, gotEntries)
		_ = db.Close()
	}
}

func TestUpdateAndLookup(t *testing.T) {
	storageDB := path.Join(os.TempDir(), storageDB)

	defer func() {
		os.Remove(storageDB)
	}()

	// ------------------

	{
		Desc(t, "Store then lookup stats")

		// Build
		db, _ := NewStorage(storageDB, time.Hour)
		stats := core.StatsMetadata{
			Altitude:  35,
			Longitude: -3.4546,
			Latitude:  35.212,
		}
		gid := []byte{0, 0, 0, 0, 0, 0, 0, 1}

		// Operate
		errUpdate := db.UpdateStats(gid, stats)
		got, errLookup := db.LookupStats(gid)

		// Check
		CheckErrors(t, nil, errUpdate)
		CheckErrors(t, nil, errLookup)
		Check(t, stats, got, "Metadata")
		_ = db.Close()
	}

	// ------------------

	{
		Desc(t, "Lookup stats from unknown gateway")

		// Build
		db, _ := NewStorage(storageDB, time.Hour)
		gid := []byte{0, 0, 0, 0, 0, 0, 0, 2}

		// Operate
		got, errLookup := db.LookupStats(gid)

		// Check
		CheckErrors(t, ErrNotFound, errLookup)
		Check(t, core.StatsMetadata{}, got, "Metadata")
		_ = db.Close()
	}
}

// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package router

import (
	"os"
	"path"
	"testing"
	"time"

	. "github.com/TheThingsNetwork/ttn/utils/testing"
)

const storageDB = "TestBrkStorage.db"

func CheckEntries(t *testing.T, want []brkEntry, got []brkEntry) {
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

func TestCreateAndRead(t *testing.T) {
	storageDB := path.Join(os.TempDir(), storageDB)

	defer func() {
		os.Remove(storageDB)
	}()

	// ------------------

	{
		Desc(t, "Create a new storage")
		db, err := NewBrkStorage(storageDB, time.Hour)
		CheckErrors(t, nil, err)
		err = db.done()
		CheckErrors(t, nil, err)
	}

	// ------------------

	{
		Desc(t, "create then read a device")

		// Build
		db, _ := NewBrkStorage(storageDB, time.Hour)
		entry := brkEntry{
			DevAddr:     []byte{0, 0, 0, 1},
			BrokerIndex: 1,
		}

		// Operate
		err := db.create(entry)
		FatalUnless(t, err)
		gotbrkEntry, err := db.read(entry.DevAddr)

		// Expectations
		wantbrkEntry := []brkEntry{
			{
				DevAddr:     entry.DevAddr,
				BrokerIndex: entry.BrokerIndex,
				until:       time.Now().Add(time.Hour),
			},
		}

		// Check
		CheckErrors(t, nil, err)
		CheckEntries(t, wantbrkEntry, gotbrkEntry)
		_ = db.done()
	}

	// ------------------

	{
		Desc(t, "read non-existing brkEntry")

		// Build
		db, _ := NewBrkStorage(storageDB, time.Hour)
		entry := brkEntry{
			DevAddr:     []byte{0, 0, 0, 2},
			BrokerIndex: 1,
		}

		// Operate
		gotbrkEntry, err := db.read(entry.DevAddr)

		// Checks
		CheckErrors(t, ErrNotFound, err)
		CheckEntries(t, nil, gotbrkEntry)
		_ = db.done()
	}

	// ------------------

	{
		Desc(t, "read an expired brkEntry")

		// Build
		db, _ := NewBrkStorage(storageDB, time.Millisecond*100)
		entry := brkEntry{
			DevAddr:     []byte{0, 0, 0, 3},
			BrokerIndex: 1,
		}

		// Operate
		_ = db.create(entry)
		<-time.After(time.Millisecond * 200)
		gotbrkEntry, err := db.read(entry.DevAddr)

		// Checks
		CheckErrors(t, ErrNotFound, err)
		CheckEntries(t, nil, gotbrkEntry)
		_ = db.done()
	}

	// ------------------

	{
		Desc(t, "create above an expired brkEntry")

		// Build
		db, _ := NewBrkStorage(storageDB, time.Millisecond*100)
		entry := brkEntry{
			DevAddr:     []byte{0, 0, 0, 4},
			BrokerIndex: 1,
		}
		entry2 := brkEntry{
			DevAddr:     []byte{0, 0, 0, 4},
			BrokerIndex: 12,
		}

		// Operate
		_ = db.create(entry)
		<-time.After(time.Millisecond * 200)
		err := db.create(entry2)
		FatalUnless(t, err)
		gotbrkEntry, err := db.read(entry.DevAddr)

		// Expectations
		wantbrkEntry := []brkEntry{
			{
				DevAddr:     entry2.DevAddr,
				BrokerIndex: entry2.BrokerIndex,
				until:       time.Now().Add(time.Millisecond * 100),
			},
		}

		// Checks
		CheckErrors(t, nil, err)
		CheckEntries(t, wantbrkEntry, gotbrkEntry)
		_ = db.done()
	}

	// ------------------

	{
		Desc(t, "create on a closed database")

		// Build
		db, _ := NewBrkStorage(storageDB, time.Hour)
		_ = db.done()
		entry := brkEntry{
			DevAddr: []byte{0, 0, 0, 5},
		}

		// Operate
		err := db.create(entry)

		// Checks
		CheckErrors(t, ErrOperational, err)
	}

	// ------------------

	{
		Desc(t, "read on a closed database")

		// Build
		db, _ := NewBrkStorage(storageDB, time.Hour)
		_ = db.done()
		devAddr := []byte{0, 0, 0, 1}

		// Operate
		gotbrkEntry, err := db.read(devAddr)

		// Checks
		CheckErrors(t, ErrOperational, err)
		CheckEntries(t, nil, gotbrkEntry)
	}

	// ------------------

	{
		Desc(t, "create two entries in a row")

		// Build
		db, _ := NewBrkStorage(storageDB, time.Hour)
		entry1 := brkEntry{
			DevAddr:     []byte{0, 0, 0, 6},
			BrokerIndex: 14,
		}

		entry2 := brkEntry{
			DevAddr:     []byte{0, 0, 0, 6},
			BrokerIndex: 20,
		}

		// Operate
		err := db.create(entry1)
		FatalUnless(t, err)
		err = db.create(entry2)
		FatalUnless(t, err)
		gotEntries, err := db.read(entry1.DevAddr)
		FatalUnless(t, err)

		// Expectations
		wantEntries := []brkEntry{
			{
				DevAddr:     entry1.DevAddr,
				BrokerIndex: entry1.BrokerIndex,
				until:       time.Now().Add(time.Hour),
			},
			{
				DevAddr:     entry2.DevAddr,
				BrokerIndex: entry2.BrokerIndex,
				until:       time.Now().Add(time.Hour),
			},
		}

		// Check
		CheckEntries(t, wantEntries, gotEntries)
		_ = db.done()
	}
}

// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package router

import (
	"os"
	"path"
	"testing"

	"github.com/TheThingsNetwork/ttn/core"
	. "github.com/TheThingsNetwork/ttn/utils/testing"
)

const gatewaysDB = "TestGtwStorage.db"

func TestUpsertAndRead(t *testing.T) {
	gatewaysDB := path.Join(os.TempDir(), gatewaysDB)

	defer func() {
		os.Remove(gatewaysDB)
	}()

	// ------------------

	{
		Desc(t, "Createa new storage")
		db, err := NewGtwStorage(gatewaysDB)
		CheckErrors(t, nil, err)
		err = db.done()
		CheckErrors(t, nil, err)
	}

	// ------------------

	{
		Desc(t, "upsert then read a device")

		// Build
		db, _ := NewGtwStorage(gatewaysDB)
		entry := gtwEntry{
			GatewayID: []byte{0, 0, 0, 1},
			Metadata: core.StatsMetadata{
				Altitude:  14,
				Longitude: -14,
			},
		}

		// Operate
		err := db.upsert(entry)
		FatalUnless(t, err)
		gotGtwEntry, err := db.read(entry.GatewayID)

		// Expectations
		wantGtwEntry := gtwEntry{
			GatewayID: entry.GatewayID,
			Metadata:  entry.Metadata,
		}

		// Check
		CheckErrors(t, nil, err)
		Check(t, wantGtwEntry, gotGtwEntry, "Gateway Entries")
		_ = db.done()
	}

	// ------------------

	{
		Desc(t, "read non-existing gtwEntry")

		// Build
		db, _ := NewGtwStorage(gatewaysDB)
		entry := gtwEntry{
			GatewayID: []byte{0, 0, 0, 2},
			Metadata: core.StatsMetadata{
				Altitude:  14,
				Longitude: -14,
			},
		}

		// Operate
		gotGtwEntry, err := db.read(entry.GatewayID)

		// Checks
		CheckErrors(t, ErrNotFound, err)
		Check(t, gtwEntry{}, gotGtwEntry, "Gateway Entries")
		_ = db.done()
	}

	// ------------------

	{
		Desc(t, "upsert on a closed database")

		// Build
		db, _ := NewGtwStorage(gatewaysDB)
		_ = db.done()
		entry := gtwEntry{
			GatewayID: []byte{0, 0, 0, 5},
			Metadata: core.StatsMetadata{
				Altitude:  14,
				Longitude: -14,
			},
		}

		// Operate
		err := db.upsert(entry)

		// Checks
		CheckErrors(t, ErrOperational, err)
	}

	// ------------------

	{
		Desc(t, "read on a closed database")

		// Build
		db, _ := NewGtwStorage(gatewaysDB)
		_ = db.done()
		devAddr := []byte{0, 0, 0, 1}

		// Operate
		gotGtwEntry, err := db.read(devAddr)

		// Checks
		CheckErrors(t, ErrOperational, err)
		Check(t, gtwEntry{}, gotGtwEntry, "Gateway Entries")
	}

	// ------------------

	{
		Desc(t, "upsert two entries in a row")

		// Build
		db, _ := NewGtwStorage(gatewaysDB)
		entry1 := gtwEntry{
			GatewayID: []byte{0, 0, 0, 6},
			Metadata: core.StatsMetadata{
				Altitude:  14,
				Longitude: -14,
			},
		}

		entry2 := gtwEntry{
			GatewayID: []byte{0, 0, 0, 6},
			Metadata: core.StatsMetadata{
				Altitude:  42,
				Longitude: -42,
			},
		}

		// Operate
		err := db.upsert(entry1)
		FatalUnless(t, err)
		err = db.upsert(entry2)
		FatalUnless(t, err)
		gotEntries, err := db.read(entry1.GatewayID)
		FatalUnless(t, err)

		// Expectations
		wantEntries := gtwEntry{
			GatewayID: entry1.GatewayID,
			Metadata:  entry2.Metadata,
		}

		// Check
		Check(t, wantEntries, gotEntries, "Gateway Entries")
		_ = db.done()
	}
}

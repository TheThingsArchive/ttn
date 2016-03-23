// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package broker

import (
	"os"
	"path"
	"testing"

	. "github.com/TheThingsNetwork/ttn/utils/testing"
)

const devDB = "TestDevStorage.db"

func TestReadStore(t *testing.T) {
	var db AppStorage
	defer func() {
		os.Remove(path.Join(os.TempDir(), devDB))
	}()

	// ------------------

	{
		Desc(t, "Create a new storage")
		var err error
		db, err = NewAppStorage(path.Join(os.TempDir(), devDB))
		CheckErrors(t, nil, err)
	}

	// ------------------

	{
		Desc(t, "Store and read a registration")

		// Build
		entry := appEntry{
			Dialer: NewDialer([]byte("dialer")),
			AppEUI: []byte{0, 2},
		}

		// Operate
		err := db.upsert(entry)
		FatalUnless(t, err)
		got, err := db.read(entry.AppEUI)

		// Check
		CheckErrors(t, nil, err)
		Check(t, entry, got, "Device Entries")
	}

	// ------------------

	{
		Desc(t, "read a non-existing registration")

		// Build
		appEUI := []byte{0, 0, 0, 0, 0, 0, 0, 2}

		// Operate
		_, err := db.read(appEUI)

		// Check
		CheckErrors(t, ErrNotFound, err)
	}

	// ------------------

	{
		Desc(t, "Store twice the same registration")

		// Build
		entry := appEntry{
			Dialer: NewDialer([]byte("dialer")),
			AppEUI: []byte{0, 1},
		}

		// Operate
		err := db.upsert(entry)
		FatalUnless(t, err)
		err = db.upsert(entry)
		FatalUnless(t, err)
		got, err := db.read(entry.AppEUI)

		// Check
		CheckErrors(t, nil, err)
		Check(t, entry, got, "Device Entries")
	}

	// ------------------

	{
		Desc(t, "Update an entry")

		// Build
		entry := appEntry{
			Dialer: NewDialer([]byte("dialer")),
			AppEUI: []byte{0, 4},
		}
		update := appEntry{
			Dialer: NewDialer([]byte("dialer")),
			AppEUI: []byte{0, 4},
		}

		// Operate
		err := db.upsert(entry)
		FatalUnless(t, err)
		err = db.upsert(update)
		got, errRead := db.read(entry.AppEUI)
		FatalUnless(t, errRead)

		// Check
		CheckErrors(t, nil, err)
		Check(t, update, got, "Device Entries")
	}

	// ------------------

	{
		Desc(t, "Close the storage")
		err := db.done()
		CheckErrors(t, nil, err)
	}
}

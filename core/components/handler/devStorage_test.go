// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package handler

import (
	"os"
	"path"
	"testing"

	. "github.com/TheThingsNetwork/ttn/utils/testing"
)

const devDB = "TestDevStorage.db"

//UpdateFCnt(appEUI []byte, devEUI []byte, fcnt uint32) error
//Lookup(appEUI []byte, devEUI []byte) (devEntry, error)
//StorePersonalized(appEUI []byte, devAddr [4]byte, appSKey, nwkSKey [16]byte) error
//Close() error

func TestLookupStore(t *testing.T) {
	var db DevStorage
	defer func() {
		os.Remove(path.Join(os.TempDir(), devDB))
	}()

	// ------------------

	{
		Desc(t, "Create a new storage")
		var err error
		db, err = NewDevStorage(path.Join(os.TempDir(), devDB))
		CheckErrors(t, nil, err)
	}

	// ------------------

	{
		Desc(t, "Store and Lookup a registration")

		// Build
		appEUI := []byte{1, 2, 3, 4, 5, 6, 7, 8}
		devEUI := []byte{0, 0, 0, 0, 1, 2, 3, 4}
		devAddr := [4]byte{1, 2, 3, 4}
		appSKey := [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6}
		nwkSKey := [16]byte{6, 5, 4, 3, 2, 1, 0, 9, 8, 7, 6, 5, 4, 3, 2, 1}

		// Expect
		var want = devEntry{
			DevAddr:  devAddr,
			AppSKey:  appSKey,
			NwkSKey:  nwkSKey,
			FCntDown: 0,
		}

		// Operate
		err := db.StorePersonalized(appEUI, devAddr, appSKey, nwkSKey)
		FatalUnless(t, err)
		entry, err := db.Lookup(appEUI, devEUI)

		// Check
		CheckErrors(t, nil, err)
		Check(t, want, entry, "Device Entries")
	}

	// ------------------

	{
		Desc(t, "Lookup a non-existing registration")

		// Build
		appEUI := []byte{0, 0, 0, 0, 0, 0, 0, 1}
		devEUI := []byte{0, 0, 0, 0, 1, 2, 3, 4}

		// Operate
		_, err := db.Lookup(appEUI, devEUI)

		// Check
		CheckErrors(t, ErrNotFound, err)
	}

	// ------------------

	{
		Desc(t, "Store twice the same registration")

		// Build
		appEUI := []byte{1, 2, 3, 4, 5, 6, 7, 9}
		devEUI := []byte{0, 0, 0, 0, 1, 2, 3, 4}
		devAddr := [4]byte{1, 2, 3, 4}
		appSKey := [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6}
		nwkSKey := [16]byte{6, 5, 4, 3, 2, 1, 0, 9, 8, 7, 6, 5, 4, 3, 2, 1}

		// Expect
		var want = devEntry{
			DevAddr:  devAddr,
			AppSKey:  appSKey,
			NwkSKey:  nwkSKey,
			FCntDown: 0,
		}

		// Operate
		err := db.StorePersonalized(appEUI, devAddr, appSKey, nwkSKey)
		FatalUnless(t, err)
		err = db.StorePersonalized(appEUI, devAddr, appSKey, nwkSKey)
		FatalUnless(t, err)
		entry, err := db.Lookup(appEUI, devEUI)

		// Check
		CheckErrors(t, nil, err)
		Check(t, want, entry, "Device Entries")
	}

	// ------------------

	{
		Desc(t, "Update FCnt")

		// Build
		appEUI := []byte{1, 2, 3, 4, 5, 6, 7, 14}
		devEUI := []byte{0, 0, 0, 0, 1, 2, 3, 4}
		devAddr := [4]byte{1, 2, 3, 4}
		appSKey := [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6}
		nwkSKey := [16]byte{6, 5, 4, 3, 2, 1, 0, 9, 8, 7, 6, 5, 4, 3, 2, 1}
		fcnt := uint32(2)

		// Expect
		var want = devEntry{
			DevAddr:  devAddr,
			AppSKey:  appSKey,
			NwkSKey:  nwkSKey,
			FCntDown: fcnt,
		}

		// Operate
		err := db.StorePersonalized(appEUI, devAddr, appSKey, nwkSKey)
		FatalUnless(t, err)
		err = db.UpdateFCnt(appEUI, devEUI, fcnt)
		entry, errLookup := db.Lookup(appEUI, devEUI)
		FatalUnless(t, errLookup)

		// Check
		CheckErrors(t, nil, err)
		Check(t, want, entry, "Device Entries")
	}

	// ------------------

	{
		Desc(t, "Update FCnt, device not found")

		// Build
		appEUI := []byte{1, 2, 3, 4, 5, 6, 7, 15}
		devEUI := []byte{0, 0, 0, 0, 1, 2, 3, 4}
		fcnt := uint32(2)

		// Operate
		err := db.UpdateFCnt(appEUI, devEUI, fcnt)

		// Check
		CheckErrors(t, ErrNotFound, err)
	}

	// ------------------

	{
		Desc(t, "Close the storage")
		err := db.Close()
		CheckErrors(t, nil, err)
	}
}

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

func TestStore(t *testing.T) {
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
		Desc(t, "Store and read a registration")

		// Build
		entry := devEntry{
			AppEUI:  []byte{1, 2, 3, 4, 5, 6, 7, 8},
			DevEUI:  []byte{0, 0, 0, 0, 1, 2, 3, 4},
			DevAddr: []byte{1, 2, 3, 4},
			AppSKey: [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6},
			NwkSKey: [16]byte{6, 5, 4, 3, 2, 1, 0, 9, 8, 7, 6, 5, 4, 3, 2, 1},
		}

		// Operate
		err := db.upsert(entry)
		FatalUnless(t, err)
		got, err := db.read(entry.AppEUI, entry.DevEUI)

		// Check
		CheckErrors(t, nil, err)
		Check(t, entry, got, "Device Entries")
	}

	// ------------------

	{
		Desc(t, "Read a non-existing registration")

		// Build
		appEUI := []byte{0, 0, 0, 0, 0, 0, 0, 1}
		devEUI := []byte{0, 0, 0, 0, 1, 2, 3, 4}

		// Operate
		_, err := db.read(appEUI, devEUI)

		// Check
		CheckErrors(t, ErrNotFound, err)
	}

	// ------------------

	{
		Desc(t, "Store twice the same registration")

		// Build
		entry := devEntry{
			AppEUI:  []byte{1, 2, 3, 4, 5, 6, 7, 9},
			DevEUI:  []byte{0, 0, 0, 0, 1, 2, 3, 4},
			DevAddr: []byte{1, 2, 3, 4},
			AppSKey: [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6},
			NwkSKey: [16]byte{6, 5, 4, 3, 2, 1, 0, 9, 8, 7, 6, 5, 4, 3, 2, 1},
		}

		// Operate
		err := db.upsert(entry)
		FatalUnless(t, err)
		err = db.upsert(entry)
		FatalUnless(t, err)
		got, err := db.read(entry.AppEUI, entry.DevEUI)

		// Check
		CheckErrors(t, nil, err)
		Check(t, entry, got, "Device Entries")
	}

	// ------------------

	{
		Desc(t, "Update FCnt")

		// Build
		entry := devEntry{
			AppEUI:   []byte{1, 2, 3, 4, 5, 6, 7, 14},
			DevEUI:   []byte{0, 0, 0, 0, 1, 2, 3, 4},
			DevAddr:  []byte{1, 2, 3, 4},
			AppSKey:  [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6},
			NwkSKey:  [16]byte{6, 5, 4, 3, 2, 1, 0, 9, 8, 7, 6, 5, 4, 3, 2, 1},
			FCntDown: 2,
		}
		update := devEntry{
			AppEUI:   entry.AppEUI,
			DevEUI:   entry.DevEUI,
			DevAddr:  entry.DevAddr,
			AppSKey:  entry.AppSKey,
			NwkSKey:  entry.NwkSKey,
			FCntDown: 14,
		}

		// Operate
		err := db.upsert(entry)
		FatalUnless(t, err)
		err = db.upsert(update)
		got, errRead := db.read(entry.AppEUI, entry.DevEUI)
		FatalUnless(t, errRead)

		// Check
		CheckErrors(t, nil, err)
		Check(t, update, got, "Device Entries")
	}

	// ------------------

	{
		Desc(t, "Store several, then readAll")

		// Build
		entry1 := devEntry{
			AppEUI:   []byte{1, 2, 3, 44, 54, 6, 7, 14},
			DevEUI:   []byte{0, 0, 0, 0, 1, 2, 3, 4},
			DevAddr:  []byte{1, 2, 3, 4},
			AppSKey:  [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6},
			NwkSKey:  [16]byte{6, 5, 4, 3, 2, 1, 0, 9, 8, 7, 6, 5, 4, 3, 2, 1},
			FCntDown: 2,
		}
		entry2 := devEntry{
			AppEUI:   []byte{1, 2, 3, 44, 54, 6, 7, 14},
			DevEUI:   []byte{0, 0, 0, 0, 1, 2, 3, 5},
			DevAddr:  []byte{2, 2, 3, 4},
			AppSKey:  [16]byte{2, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6},
			NwkSKey:  [16]byte{7, 5, 4, 3, 2, 1, 0, 9, 8, 7, 6, 5, 4, 3, 2, 1},
			FCntDown: 3,
		}
		entry3 := devEntry{
			AppEUI:   []byte{1, 8, 9, 44, 54, 6, 7, 14},
			DevEUI:   []byte{0, 0, 0, 0, 1, 2, 3, 5},
			DevAddr:  []byte{2, 2, 3, 4},
			AppSKey:  [16]byte{2, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6},
			NwkSKey:  [16]byte{7, 5, 4, 3, 2, 1, 0, 9, 8, 7, 6, 5, 4, 3, 2, 1},
			FCntDown: 3,
		}

		// Operate
		err := db.upsert(entry1)
		FatalUnless(t, err)
		err = db.upsert(entry2)
		FatalUnless(t, err)
		err = db.upsert(entry3)
		FatalUnless(t, err)
		entries, err := db.readAll(entry1.AppEUI)

		// Check
		CheckErrors(t, nil, err)
		Check(t, []devEntry{entry1, entry2}, entries, "Devices Entries")
	}

	// ------------------

	{
		Desc(t, "Read a non-existing default device entry")

		// Build
		appEUI := []byte{0, 0, 0, 0, 0, 0, 0, 1}

		// Operate
		entry, err := db.getDefault(appEUI)

		// Expect
		var want *devDefaultEntry

		// Check
		CheckErrors(t, nil, err)
		Check(t, want, entry, "Default Entry")
	}

	// ------------------

	{
		Desc(t, "Set and get default device entry")

		// Build
		appEUI := []byte{1, 2, 3, 4, 5, 6, 7, 8}
		entry := devDefaultEntry{
			AppKey: [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 1, 2, 3, 4, 5, 6, 7, 8},
		}

		// Operate
		err := db.setDefault(appEUI, &entry)
		FatalUnless(t, err)
		want, err := db.getDefault(appEUI)

		// Check
		FatalUnless(t, err)
		Check(t, *want, entry, "Default Entry")
	}

	// ------------------

	{
		Desc(t, "Close the storage")
		err := db.done()
		CheckErrors(t, nil, err)
	}
}

func TestMarshalUnmarshalDevEntries(t *testing.T) {
	{
		Desc(t, "Complete Entry")
		entry := devEntry{
			AppEUI:   []byte{1, 2, 3, 4, 5, 6, 7, 8},
			AppKey:   &[16]byte{1, 2, 1, 2, 1, 2, 1, 2, 1, 2, 1, 2, 1, 2, 1, 2},
			AppSKey:  [16]byte{0, 9, 8, 7, 6, 5, 4, 3, 2, 1, 6, 5, 4, 3, 2, 1},
			DevAddr:  []byte{4, 4, 4, 4},
			DevEUI:   []byte{14, 14, 14, 14, 14, 14, 14, 14},
			FCntDown: 42,
			NwkSKey:  [16]byte{28, 27, 26, 25, 24, 23, 22, 21, 20, 19, 18, 17, 16, 15, 14, 13},
		}

		data, err := entry.MarshalBinary()
		CheckErrors(t, nil, err)
		unmarshaled := new(devEntry)
		err = unmarshaled.UnmarshalBinary(data)
		CheckErrors(t, nil, err)
		Check(t, entry, *unmarshaled, "Entries")
	}

	// --------------------

	{
		Desc(t, "Partial Entry")
		entry := devEntry{
			AppEUI:   []byte{1, 2, 3, 4, 5, 6, 7, 8},
			AppKey:   &[16]byte{1, 2, 1, 2, 1, 2, 1, 2, 1, 2, 1, 2, 1, 2, 1, 2},
			DevEUI:   []byte{14, 14, 14, 14, 14, 14, 14, 14},
			FCntDown: 0,
		}
		want := devEntry{
			AppEUI:   []byte{1, 2, 3, 4, 5, 6, 7, 8},
			AppKey:   &[16]byte{1, 2, 1, 2, 1, 2, 1, 2, 1, 2, 1, 2, 1, 2, 1, 2},
			AppSKey:  [16]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			DevAddr:  make([]byte, 0, 0),
			DevEUI:   []byte{14, 14, 14, 14, 14, 14, 14, 14},
			FCntDown: 0,
			NwkSKey:  [16]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		}

		data, err := entry.MarshalBinary()
		CheckErrors(t, nil, err)
		unmarshaled := new(devEntry)
		err = unmarshaled.UnmarshalBinary(data)
		CheckErrors(t, nil, err)
		Check(t, want, *unmarshaled, "Entries")
	}

	// --------------------

	{
		Desc(t, "Partial Entry bis")
		entry := devEntry{
			AppEUI:   []byte{1, 2, 3, 4, 5, 6, 7, 8},
			AppKey:   &[16]byte{1, 2, 1, 2, 1, 2, 1, 2, 1, 2, 1, 2, 1, 2, 1, 2},
			DevEUI:   []byte{14, 14, 14, 14, 14, 14, 14, 14},
			DevAddr:  []byte{},
			FCntDown: 0,
		}
		want := devEntry{
			AppEUI:   []byte{1, 2, 3, 4, 5, 6, 7, 8},
			AppKey:   &[16]byte{1, 2, 1, 2, 1, 2, 1, 2, 1, 2, 1, 2, 1, 2, 1, 2},
			AppSKey:  [16]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
			DevAddr:  make([]byte, 0, 0),
			DevEUI:   []byte{14, 14, 14, 14, 14, 14, 14, 14},
			FCntDown: 0,
			NwkSKey:  [16]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		}

		data, err := entry.MarshalBinary()
		CheckErrors(t, nil, err)
		unmarshaled := new(devEntry)
		err = unmarshaled.UnmarshalBinary(data)
		CheckErrors(t, nil, err)
		Check(t, want, *unmarshaled, "Entries")
	}
}

func TestMarshalUnmarshalDevDefaultEntries(t *testing.T) {
	{
		Desc(t, "Entry")
		entry := devDefaultEntry{
			AppKey: [16]byte{1, 2, 1, 2, 1, 2, 1, 2, 1, 2, 1, 2, 1, 2, 1, 2},
		}

		data, err := entry.MarshalBinary()
		CheckErrors(t, nil, err)
		unmarshaled := new(devDefaultEntry)
		err = unmarshaled.UnmarshalBinary(data)
		CheckErrors(t, nil, err)
		Check(t, entry, *unmarshaled, "Entries")
	}
}

// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package broker

import (
	"os"
	"path"
	"testing"

	. "github.com/TheThingsNetwork/ttn/utils/testing"
)

const NetworkControllerDB = "TestBrokerNetworkController.db"

func TestNetworkControllerDevice(t *testing.T) {
	NetworkControllerDB := path.Join(os.TempDir(), NetworkControllerDB)
	defer func() {
		os.Remove(NetworkControllerDB)
	}()

	// -------------------

	{
		Desc(t, "Create a new NetworkController")
		db, err := NewNetworkController(NetworkControllerDB)
		CheckErrors(t, nil, err)
		err = db.done()
		CheckErrors(t, nil, err)
	}

	// -------------------

	{
		Desc(t, "Store then lookup a registration")

		// Build
		db, _ := NewNetworkController(NetworkControllerDB)
		entry := devEntry{
			DevAddr: []byte{1, 2, 3, 4},
			Dialer:  NewDialer([]byte("url")),
			AppEUI:  []byte{1, 2, 3, 4, 5, 6, 7, 8},
			DevEUI:  []byte{0, 0, 0, 0, 1, 2, 3, 4},
			NwkSKey: [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6},
			FCntUp:  14,
		}

		// Operate
		err := db.upsert(entry)
		FatalUnless(t, err)
		entries, err := db.read(entry.DevAddr)

		// Expect
		want := []devEntry{entry}

		// Check
		CheckErrors(t, nil, err)
		Check(t, want, entries, "DevEntries")
		_ = db.done()
	}

	// -------------------

	{
		Desc(t, "Store entries with same DevAddr")

		// Build
		db, _ := NewNetworkController(NetworkControllerDB)
		entry1 := devEntry{
			DevAddr: []byte{1, 2, 3, 5},
			Dialer:  NewDialer([]byte("url")),
			AppEUI:  []byte{1, 2, 3, 4, 5, 6, 7, 8},
			DevEUI:  []byte{0, 0, 0, 0, 1, 2, 3, 5},
			NwkSKey: [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6},
			FCntUp:  14,
		}
		entry2 := devEntry{
			DevAddr: []byte{1, 2, 3, 5},
			Dialer:  NewDialer([]byte("url")),
			AppEUI:  []byte{8, 7, 6, 5, 4, 3, 2, 1},
			DevEUI:  []byte{0, 0, 0, 0, 1, 2, 3, 5},
			NwkSKey: [16]byte{6, 5, 4, 3, 2, 1, 0, 9, 8, 7, 6, 5, 4, 3, 2, 1},
			FCntUp:  42,
		}

		// Operate
		err := db.upsert(entry1)
		FatalUnless(t, err)
		err = db.upsert(entry2)
		FatalUnless(t, err)
		entries, err := db.read(entry1.DevAddr)

		// Expectations
		want := []devEntry{entry1, entry2}

		// Check
		CheckErrors(t, nil, err)
		Check(t, want, entries, "DevEntries")
		_ = db.done()
	}

	// -------------------

	{
		Desc(t, "Lookup non-existing entry")

		// Build
		db, _ := NewNetworkController(NetworkControllerDB)
		devAddr := []byte{0, 0, 0, 1}

		// Operate
		entries, err := db.read(devAddr)

		// Expect
		var want []devEntry

		// Checks
		CheckErrors(t, ErrNotFound, err)
		Check(t, want, entries, "DevEntries")
		_ = db.done()
	}

	// -------------------

	{
		Desc(t, "Store on a closed database")

		// Build
		db, _ := NewNetworkController(NetworkControllerDB)
		_ = db.done()
		entry := devEntry{
			DevAddr: []byte{1, 0, 0, 2},
			Dialer:  NewDialer([]byte("url")),
			AppEUI:  []byte{1, 2, 3, 4, 5, 6, 7, 8},
			DevEUI:  []byte{0, 0, 0, 0, 1, 2, 3, 4},
			NwkSKey: [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6},
			FCntUp:  14,
		}

		// Operate
		err := db.upsert(entry)

		// Checks
		CheckErrors(t, ErrOperational, err)
	}

	// -------------------

	{
		Desc(t, "Lookup on a closed database")

		// Build
		db, _ := NewNetworkController(NetworkControllerDB)
		_ = db.done()
		devAddr := []byte{1, 2, 3, 4}

		// Operate
		entries, err := db.read(devAddr)

		// Expect
		var want []devEntry

		// Checks
		CheckErrors(t, ErrOperational, err)
		Check(t, want, entries, "DevEntries")
	}

	// -------------------

	{
		Desc(t, "Update counter up of an entry -> one device")

		// Build
		db, _ := NewNetworkController(NetworkControllerDB)
		entry := devEntry{
			DevAddr: []byte{1, 0, 0, 4},
			Dialer:  NewDialer([]byte("url")),
			AppEUI:  []byte{1, 2, 3, 4, 5, 6, 7, 8},
			DevEUI:  []byte{0, 0, 0, 0, 1, 2, 3, 4},
			NwkSKey: [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6},
			FCntUp:  14,
		}

		// Operate
		err := db.upsert(entry)
		FatalUnless(t, err)
		entry.FCntUp = 42
		err = db.upsert(entry)
		FatalUnless(t, err)
		entries, err := db.read(entry.DevAddr)
		FatalUnless(t, err)

		// Expectation
		want := []devEntry{entry}

		// Check
		Check(t, want, entries, "DevEntries")
		_ = db.done()
	}

	// -------------------

	{
		Desc(t, "Update counter several entries")

		// Build
		db, _ := NewNetworkController(NetworkControllerDB)
		entry1 := devEntry{
			DevAddr: []byte{8, 8, 8, 8},
			Dialer:  NewDialer([]byte("url")),
			AppEUI:  []byte{1, 2, 3, 4, 5, 6, 7, 8},
			DevEUI:  []byte{0, 0, 0, 0, 1, 2, 3, 5},
			NwkSKey: [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6},
			FCntUp:  14,
		}
		entry2 := devEntry{
			DevAddr: []byte{8, 8, 8, 8},
			Dialer:  NewDialer([]byte("url")),
			AppEUI:  []byte{8, 7, 6, 5, 4, 3, 2, 1},
			DevEUI:  []byte{0, 0, 0, 0, 1, 2, 3, 5},
			NwkSKey: [16]byte{6, 5, 4, 3, 2, 1, 0, 9, 8, 7, 6, 5, 4, 3, 2, 1},
			FCntUp:  42,
		}

		// Operate
		err := db.upsert(entry1)
		FatalUnless(t, err)
		err = db.upsert(entry2)
		FatalUnless(t, err)
		entry2.FCntUp = 8
		err = db.upsert(entry2)
		FatalUnless(t, err)
		entries, err := db.read(entry1.DevAddr)
		FatalUnless(t, err)

		// Expectations
		want := []devEntry{entry1, entry2}

		// Check
		Check(t, want, entries, "DevEntries")
		_ = db.done()
	}

	// --------------------

	{
		Desc(t, "Test counters, both < 65535")

		// Build
		db, _ := NewNetworkController(NetworkControllerDB)
		wholeCnt := uint32(13)
		cnt16 := wholeCnt + 1

		// Operate
		cnt32, err := db.wholeCounter(cnt16, wholeCnt)

		// Check
		CheckErrors(t, nil, err)
		Check(t, wholeCnt+1, cnt32, "Counters")

		_ = db.done()
	}

	// --------------------

	{
		Desc(t, "Test counters, devCnt < wholeCnt, | | > max_gap")

		// Build
		db, _ := NewNetworkController(NetworkControllerDB)
		wholeCnt := uint32(14)
		cnt16 := wholeCnt - 1

		// Operate
		_, err := db.wholeCounter(cnt16, wholeCnt)

		// Check
		CheckErrors(t, ErrStructural, err)

		_ = db.done()
	}

	// --------------------

	{
		Desc(t, "Test counters, whole > 65535, | | < max_gap")

		// Build
		db, _ := NewNetworkController(NetworkControllerDB)
		wholeCnt := uint32(3824624235)
		cnt16 := uint32(wholeCnt%65536 + 2)

		// Operate
		cnt32, err := db.wholeCounter(cnt16, wholeCnt)

		// Check
		CheckErrors(t, nil, err)
		Check(t, wholeCnt+2, cnt32, "Counters")

		_ = db.done()
	}

	// --------------------

	{
		Desc(t, "Test counters, whole > 65535, | | > max_gap")

		// Build
		db, _ := NewNetworkController(NetworkControllerDB)
		wholeCnt := uint32(3824624235)
		cnt16 := uint32(wholeCnt%65536 + 45000)

		// Operate
		_, err := db.wholeCounter(cnt16, wholeCnt)

		// Check
		CheckErrors(t, ErrStructural, err)

		_ = db.done()
	}

	// --------------------

	{
		Desc(t, "Test counters, whole > 65535, | | < max_gap, via inf")

		// Build
		db, _ := NewNetworkController(NetworkControllerDB)
		wholeCnt := uint32(65535)
		cnt16 := uint32(2)

		// Operate
		cnt32, err := db.wholeCounter(cnt16, wholeCnt)

		// Check
		CheckErrors(t, nil, err)
		Check(t, wholeCnt+3, cnt32, "Counters")

		_ = db.done()
	}
}

func TestNonces(t *testing.T) {
	NetworkControllerDB := path.Join(os.TempDir(), NetworkControllerDB)
	defer func() {
		os.Remove(NetworkControllerDB)
	}()

	// -------------------

	{
		Desc(t, "Store then lookup a noncesEntry")

		// Build
		db, _ := NewNetworkController(NetworkControllerDB)
		entry := noncesEntry{
			AppEUI:    []byte{1, 2, 3},
			DevEUI:    []byte{4, 5, 6},
			DevNonces: [][]byte{[]byte{14, 42}},
		}

		// Operate
		err := db.upsertNonces(entry)
		FatalUnless(t, err)
		got, err := db.readNonces(entry.AppEUI, entry.DevEUI)
		FatalUnless(t, err)

		// Check
		Check(t, entry, got, "Nonces Entries")
		_ = db.done()
	}

	// -------------------

	{
		Desc(t, "Update an existing nonce")

		// Build
		db, _ := NewNetworkController(NetworkControllerDB)
		entry := noncesEntry{
			AppEUI:    []byte{26, 2, 3},
			DevEUI:    []byte{4, 26, 26},
			DevNonces: [][]byte{[]byte{14, 42}},
		}
		update := noncesEntry{
			AppEUI:    entry.AppEUI,
			DevEUI:    entry.DevEUI,
			DevNonces: [][]byte{[]byte{58, 27}, []byte{12, 11}},
		}

		// Operate
		err := db.upsertNonces(entry)
		FatalUnless(t, err)
		err = db.upsertNonces(update)
		FatalUnless(t, err)
		got, err := db.readNonces(entry.AppEUI, entry.DevEUI)
		FatalUnless(t, err)

		// Check
		Check(t, update, got, "Nonces Entries")
		_ = db.done()
	}

	// -------------------

	{

		Desc(t, "Lookup an non-existing nonces entry")

		// Build
		db, _ := NewNetworkController(NetworkControllerDB)
		appEUI := []byte{4, 5, 3, 4}
		devEUI := []byte{1, 2, 2, 2}

		// Operate
		_, err := db.readNonces(appEUI, devEUI)

		// Check
		CheckErrors(t, ErrNotFound, err)
		_ = db.done()
	}
}

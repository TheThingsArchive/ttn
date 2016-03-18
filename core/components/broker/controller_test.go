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
		err = db.Close()
		CheckErrors(t, nil, err)
	}

	// -------------------

	{
		Desc(t, "Store then lookup a registration")

		// Build
		db, _ := NewNetworkController(NetworkControllerDB)
		devAddr := []byte{1, 2, 3, 4}
		entry := devEntry{
			Dialer:  NewDialer([]byte("url")),
			AppEUI:  []byte{1, 2, 3, 4, 5, 6, 7, 8},
			DevEUI:  []byte{0, 0, 0, 0, 1, 2, 3, 4},
			NwkSKey: [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6},
			FCntUp:  14,
		}

		// Operate
		err := db.StoreDevice(devAddr, entry)
		FatalUnless(t, err)
		entries, err := db.LookupDevices(devAddr)

		// Expect
		want := []devEntry{entry}

		// Check
		CheckErrors(t, nil, err)
		Check(t, want, entries, "DevEntries")
		_ = db.Close()
	}

	// -------------------

	{
		Desc(t, "Store entries with same DevAddr")

		// Build
		db, _ := NewNetworkController(NetworkControllerDB)
		devAddr := []byte{1, 2, 3, 5}
		entry1 := devEntry{
			Dialer:  NewDialer([]byte("url")),
			AppEUI:  []byte{1, 2, 3, 4, 5, 6, 7, 8},
			DevEUI:  []byte{0, 0, 0, 0, 1, 2, 3, 5},
			NwkSKey: [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6},
			FCntUp:  14,
		}
		entry2 := devEntry{
			Dialer:  NewDialer([]byte("url")),
			AppEUI:  []byte{8, 7, 6, 5, 4, 3, 2, 1},
			DevEUI:  []byte{0, 0, 0, 0, 1, 2, 3, 5},
			NwkSKey: [16]byte{6, 5, 4, 3, 2, 1, 0, 9, 8, 7, 6, 5, 4, 3, 2, 1},
			FCntUp:  42,
		}

		// Operate
		err := db.StoreDevice(devAddr, entry1)
		FatalUnless(t, err)
		err = db.StoreDevice(devAddr, entry2)
		FatalUnless(t, err)
		entries, err := db.LookupDevices(devAddr)

		// Expectations
		want := []devEntry{entry1, entry2}

		// Check
		CheckErrors(t, nil, err)
		Check(t, want, entries, "DevEntries")
		_ = db.Close()
	}

	// -------------------

	{
		Desc(t, "Lookup non-existing entry")

		// Build
		db, _ := NewNetworkController(NetworkControllerDB)
		devAddr := []byte{0, 0, 0, 1}

		// Operate
		entries, err := db.LookupDevices(devAddr)

		// Expect
		var want []devEntry

		// Checks
		CheckErrors(t, ErrNotFound, err)
		Check(t, want, entries, "DevEntries")
		_ = db.Close()
	}

	// -------------------

	{
		Desc(t, "Store on a closed database")

		// Build
		db, _ := NewNetworkController(NetworkControllerDB)
		_ = db.Close()
		devAddr := []byte{1, 0, 0, 2}
		entry := devEntry{
			Dialer:  NewDialer([]byte("url")),
			AppEUI:  []byte{1, 2, 3, 4, 5, 6, 7, 8},
			DevEUI:  []byte{0, 0, 0, 0, 1, 2, 3, 4},
			NwkSKey: [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6},
			FCntUp:  14,
		}

		// Operate
		err := db.StoreDevice(devAddr, entry)

		// Checks
		CheckErrors(t, ErrOperational, err)
	}

	// -------------------

	{
		Desc(t, "Lookup on a closed database")

		// Build
		db, _ := NewNetworkController(NetworkControllerDB)
		_ = db.Close()
		devAddr := []byte{1, 2, 3, 4}

		// Operate
		entries, err := db.LookupDevices(devAddr)

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
		devAddr := []byte{1, 0, 0, 4}
		entry := devEntry{
			Dialer:  NewDialer([]byte("url")),
			AppEUI:  []byte{1, 2, 3, 4, 5, 6, 7, 8},
			DevEUI:  []byte{0, 0, 0, 0, 1, 2, 3, 4},
			NwkSKey: [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6},
			FCntUp:  14,
		}

		// Operate
		err := db.StoreDevice(devAddr, entry)
		FatalUnless(t, err)
		err1 := db.UpdateFCnt(entry.AppEUI, entry.DevEUI, devAddr, 42)
		entries, err2 := db.LookupDevices(devAddr)

		// Expectations
		want := []devEntry{entry}
		want[0].FCntUp = 42

		// Check
		CheckErrors(t, nil, err1)
		CheckErrors(t, nil, err2)
		Check(t, want, entries, "DevEntries")
		_ = db.Close()
	}

	// -------------------

	{
		Desc(t, "Update counter -> fail to lookup")

		// Build
		db, _ := NewNetworkController(NetworkControllerDB)
		devAddr := []byte{14, 14, 14, 14}
		appEUI := []byte{1, 2, 3, 4, 5, 6, 7, 8}
		devEUI := []byte{1, 2, 3, 4, 5, 6, 7, 8}

		// Operate
		err := db.UpdateFCnt(appEUI, devEUI, devAddr, 14)

		// Checks
		CheckErrors(t, ErrNotFound, err)
		_ = db.Close()
	}

	// -------------------

	{
		Desc(t, "Update counter several entries")

		// Build
		db, _ := NewNetworkController(NetworkControllerDB)
		devAddr := []byte{8, 8, 8, 8}
		entry1 := devEntry{
			Dialer:  NewDialer([]byte("url")),
			AppEUI:  []byte{1, 2, 3, 4, 5, 6, 7, 8},
			DevEUI:  []byte{0, 0, 0, 0, 1, 2, 3, 5},
			NwkSKey: [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6},
			FCntUp:  14,
		}
		entry2 := devEntry{
			Dialer:  NewDialer([]byte("url")),
			AppEUI:  []byte{8, 7, 6, 5, 4, 3, 2, 1},
			DevEUI:  []byte{0, 0, 0, 0, 1, 2, 3, 5},
			NwkSKey: [16]byte{6, 5, 4, 3, 2, 1, 0, 9, 8, 7, 6, 5, 4, 3, 2, 1},
			FCntUp:  42,
		}

		// Operate
		err := db.StoreDevice(devAddr, entry1)
		FatalUnless(t, err)
		err = db.StoreDevice(devAddr, entry2)
		FatalUnless(t, err)
		err1 := db.UpdateFCnt(entry2.AppEUI, entry2.DevEUI, devAddr, 8)
		entries, err2 := db.LookupDevices(devAddr)

		// Expectations
		want := []devEntry{entry1, entry2}
		want[1].FCntUp = 8

		// Check
		CheckErrors(t, nil, err1)
		CheckErrors(t, nil, err2)
		Check(t, want, entries, "DevEntries")
		_ = db.Close()
	}

	// --------------------

	{
		Desc(t, "Test counters, both < 65535")

		// Build
		db, _ := NewNetworkController(NetworkControllerDB)
		wholeCnt := uint32(13)
		cnt16 := wholeCnt + 1

		// Operate
		cnt32, err := db.WholeCounter(cnt16, wholeCnt)

		// Check
		CheckErrors(t, nil, err)
		Check(t, wholeCnt+1, cnt32, "Counters")

		_ = db.Close()
	}

	// --------------------

	{
		Desc(t, "Test counters, devCnt < wholeCnt, | | > max_gap")

		// Build
		db, _ := NewNetworkController(NetworkControllerDB)
		wholeCnt := uint32(14)
		cnt16 := wholeCnt - 1

		// Operate
		_, err := db.WholeCounter(cnt16, wholeCnt)

		// Check
		CheckErrors(t, ErrStructural, err)

		_ = db.Close()
	}

	// --------------------

	{
		Desc(t, "Test counters, whole > 65535, | | < max_gap")

		// Build
		db, _ := NewNetworkController(NetworkControllerDB)
		wholeCnt := uint32(3824624235)
		cnt16 := uint32(wholeCnt%65536 + 2)

		// Operate
		cnt32, err := db.WholeCounter(cnt16, wholeCnt)

		// Check
		CheckErrors(t, nil, err)
		Check(t, wholeCnt+2, cnt32, "Counters")

		_ = db.Close()
	}

	// --------------------

	{
		Desc(t, "Test counters, whole > 65535, | | > max_gap")

		// Build
		db, _ := NewNetworkController(NetworkControllerDB)
		wholeCnt := uint32(3824624235)
		cnt16 := uint32(wholeCnt%65536 + 45000)

		// Operate
		_, err := db.WholeCounter(cnt16, wholeCnt)

		// Check
		CheckErrors(t, ErrStructural, err)

		_ = db.Close()
	}

	// --------------------

	{
		Desc(t, "Test counters, whole > 65535, | | < max_gap, via inf")

		// Build
		db, _ := NewNetworkController(NetworkControllerDB)
		wholeCnt := uint32(65535)
		cnt16 := uint32(2)

		// Operate
		cnt32, err := db.WholeCounter(cnt16, wholeCnt)

		// Check
		CheckErrors(t, nil, err)
		Check(t, wholeCnt+3, cnt32, "Counters")

		_ = db.Close()
	}
}

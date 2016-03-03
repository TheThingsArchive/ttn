// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package broker

import (
	"os"
	"path"
	"testing"

	. "github.com/TheThingsNetwork/ttn/core"
	. "github.com/TheThingsNetwork/ttn/core/mocks"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	. "github.com/TheThingsNetwork/ttn/utils/errors/checks"
	"github.com/TheThingsNetwork/ttn/utils/pointer"
	. "github.com/TheThingsNetwork/ttn/utils/testing"
)

const storageDB = "TestBrokerStorage.db"

func TestStorageDevice(t *testing.T) {
	storageDB := path.Join(os.TempDir(), storageDB)
	defer func() {
		os.Remove(storageDB)
	}()

	// -------------------

	{
		Desc(t, "Create a new storage")
		db, err := NewStorage(storageDB)
		CheckErrors(t, nil, err)
		err = db.Close()
		CheckErrors(t, nil, err)
	}

	// -------------------

	{
		Desc(t, "Store then lookup a registration")

		// Build
		db, _ := NewStorage(storageDB)
		r := NewMockBRegistration()

		// Operate
		err := db.StoreDevice(r)
		CheckErrors(t, nil, err)
		entries, err := db.LookupDevices(r.DevEUI())

		// Expectations
		want := []devEntry{
			{
				AppEUI:    r.AppEUI(),
				DevEUI:    r.DevEUI(),
				NwkSKey:   r.NwkSKey(),
				Recipient: r.RawRecipient(),
			},
		}

		// Check
		CheckErrors(t, nil, err)
		CheckDevEntries(t, want, entries)
		_ = db.Close()
	}

	// -------------------

	{
		Desc(t, "Store entries with same DevEUI")

		// Build
		db, _ := NewStorage(storageDB)
		r := NewMockBRegistration()
		r.OutDevEUI[0] = 34

		// Operate
		err := db.StoreDevice(r)
		CheckErrors(t, nil, err)
		err = db.StoreDevice(r)
		CheckErrors(t, nil, err)
		entries, err := db.LookupDevices(r.DevEUI())

		// Expectations
		want := []devEntry{
			{
				AppEUI:    r.AppEUI(),
				DevEUI:    r.DevEUI(),
				NwkSKey:   r.NwkSKey(),
				Recipient: r.RawRecipient(),
			},
			{
				AppEUI:    r.AppEUI(),
				DevEUI:    r.DevEUI(),
				NwkSKey:   r.NwkSKey(),
				Recipient: r.RawRecipient(),
			},
		}

		// Check
		CheckErrors(t, nil, err)
		CheckDevEntries(t, want, entries)
		_ = db.Close()
	}

	// -------------------

	{
		Desc(t, "Lookup non-existing entry")

		// Build
		db, _ := NewStorage(storageDB)
		devEUI := NewMockBRegistration().DevEUI()
		devEUI[1] = 98

		// Operate
		entries, err := db.LookupDevices(devEUI)

		// Checks
		CheckErrors(t, pointer.String(string(errors.Behavioural)), err)
		CheckDevEntries(t, nil, entries)
		_ = db.Close()
	}

	// -------------------

	{
		Desc(t, "Store on a closed database")

		// Build
		db, _ := NewStorage(storageDB)
		_ = db.Close()
		r := NewMockBRegistration()
		r.OutDevEUI[5] = 9

		// Operate
		err := db.StoreDevice(r)

		// Checks
		CheckErrors(t, pointer.String(string(errors.Operational)), err)
	}

	// -------------------

	{
		Desc(t, "Lookup on a closed database")

		// Build
		db, _ := NewStorage(storageDB)
		_ = db.Close()
		devEUI := NewMockBRegistration().DevEUI()

		// Operate
		entries, err := db.LookupDevices(devEUI)

		// Checks
		CheckErrors(t, pointer.String(string(errors.Operational)), err)
		CheckDevEntries(t, nil, entries)
	}

	// -------------------

	{
		Desc(t, "Store an invalid recipient")

		// Build
		db, _ := NewStorage(storageDB)
		r := NewMockBRegistration()
		r.OutDevEUI[7] = 99
		r.OutRecipient.(*MockRecipient).Failures["MarshalBinary"] = errors.New(errors.Structural, "Mock Error: MarshalBinary")

		// Operate & Check
		err := db.StoreDevice(r)
		CheckErrors(t, pointer.String(string(errors.Structural)), err)
		entries, err := db.LookupDevices(r.DevEUI())
		CheckErrors(t, pointer.String(string(errors.Behavioural)), err)
		CheckDevEntries(t, nil, entries)

		_ = db.Close()
	}
}

func TestStorageApplication(t *testing.T) {
	defer func() {
		os.Remove(storageDB)
	}()

	// -------------------

	{
		Desc(t, "Create a new storage")
		db, err := NewStorage(storageDB)
		CheckErrors(t, nil, err)
		err = db.Close()
		CheckErrors(t, nil, err)
	}

	// -------------------

	{
		Desc(t, "Store then lookup a registration")

		// Build
		db, _ := NewStorage(storageDB)
		r := NewMockARegistration()

		// Operate
		err := db.StoreApplication(r)
		CheckErrors(t, nil, err)
		entry, err := db.LookupApplication(r.AppEUI())

		// Expectations
		want := appEntry{
			AppEUI:    r.AppEUI(),
			Recipient: r.RawRecipient(),
		}

		// Check
		CheckErrors(t, nil, err)
		CheckAppEntries(t, want, entry)
		_ = db.Close()
	}

	// -------------------

	{
		Desc(t, "Store entries with same AppEUI")

		// Build
		db, _ := NewStorage(storageDB)
		r1 := NewMockARegistration()
		r1.OutAppEUI[0] = 34
		r2 := NewMockARegistration()
		r2.OutAppEUI[0] = 34
		r2.OutRecipient = NewMockRecipient()
		r2.OutRecipient.(*MockRecipient).OutMarshalBinary = []byte{88, 99, 77, 66}

		// Operate
		err := db.StoreApplication(r1)
		CheckErrors(t, nil, err)
		err = db.StoreApplication(r2)
		CheckErrors(t, nil, err)
		entry, err := db.LookupApplication(r1.AppEUI())

		// Expectations
		want := appEntry{
			AppEUI:    r2.AppEUI(),
			Recipient: r2.RawRecipient(),
		}

		// Check
		CheckErrors(t, nil, err)
		CheckAppEntries(t, want, entry)
		_ = db.Close()
	}

	// -------------------

	{
		Desc(t, "Lookup non-existing entry")

		// Build
		db, _ := NewStorage(storageDB)
		appEUI := NewMockARegistration().AppEUI()
		appEUI[1] = 98

		// Operate
		entry, err := db.LookupApplication(appEUI)

		// Checks
		CheckErrors(t, pointer.String(string(errors.Behavioural)), err)
		CheckAppEntries(t, appEntry{}, entry)
		_ = db.Close()
	}

	// -------------------

	{
		Desc(t, "Store on a closed database")

		// Build
		db, _ := NewStorage(storageDB)
		_ = db.Close()
		r := NewMockARegistration()
		r.OutAppEUI[5] = 9

		// Operate
		err := db.StoreApplication(r)

		// Checks
		CheckErrors(t, pointer.String(string(errors.Operational)), err)
	}

	// -------------------

	{
		Desc(t, "Lookup on a closed database")

		// Build
		db, _ := NewStorage(storageDB)
		_ = db.Close()
		appEUI := NewMockARegistration().AppEUI()

		// Operate
		entry, err := db.LookupApplication(appEUI)

		// Checks
		CheckErrors(t, pointer.String(string(errors.Operational)), err)
		CheckAppEntries(t, appEntry{}, entry)
	}

	// -------------------

	{
		Desc(t, "Store an invalid recipient")

		// Build
		db, _ := NewStorage(storageDB)
		r := NewMockARegistration()
		r.OutAppEUI[7] = 99
		r.OutRecipient.(*MockRecipient).Failures["MarshalBinary"] = errors.New(errors.Structural, "Mock Error: MarshalBinary")

		// Operate & Check
		err := db.StoreApplication(r)
		CheckErrors(t, pointer.String(string(errors.Structural)), err)
		entry, err := db.LookupApplication(r.AppEUI())
		CheckErrors(t, pointer.String(string(errors.Behavioural)), err)
		CheckAppEntries(t, appEntry{}, entry)

		_ = db.Close()
	}
}

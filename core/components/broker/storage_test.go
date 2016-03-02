// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package broker

import (
	"os"
	"testing"

	. "github.com/TheThingsNetwork/ttn/core"
	. "github.com/TheThingsNetwork/ttn/core/mocks"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	. "github.com/TheThingsNetwork/ttn/utils/errors/checks"
	"github.com/TheThingsNetwork/ttn/utils/pointer"
	. "github.com/TheThingsNetwork/ttn/utils/testing"
)

const storageDB = "TestBrokerStorage.db"

func TestStorage(t *testing.T) {
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
		r := NewMockRegistration()

		// Operate
		err := db.Store(r)
		CheckErrors(t, nil, err)
		entries, err := db.Lookup(r.DevEUI())

		// Expectations
		want := []entry{
			{
				AppEUI:    r.AppEUI(),
				DevEUI:    r.DevEUI(),
				NwkSKey:   r.NwkSKey(),
				Recipient: r.RawRecipient(),
			},
		}

		// Check
		CheckErrors(t, nil, err)
		CheckEntries(t, want, entries)
		_ = db.Close()
	}

	// -------------------

	{
		Desc(t, "Store entries with same DevEUI")

		// Build
		db, _ := NewStorage(storageDB)
		r := NewMockRegistration()
		r.OutDevEUI[0] = 34

		// Operate
		err := db.Store(r)
		CheckErrors(t, nil, err)
		err = db.Store(r)
		CheckErrors(t, nil, err)
		entries, err := db.Lookup(r.DevEUI())

		// Expectations
		want := []entry{
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
		CheckEntries(t, want, entries)
		_ = db.Close()
	}

	// -------------------

	{
		Desc(t, "Lookup non-existing entry")

		// Build
		db, _ := NewStorage(storageDB)
		devEUI := NewMockRegistration().DevEUI()
		devEUI[1] = 98

		// Operate
		entries, err := db.Lookup(devEUI)

		// Checks
		CheckErrors(t, pointer.String(string(errors.Behavioural)), err)
		CheckEntries(t, nil, entries)
		_ = db.Close()
	}

	// -------------------

	{
		Desc(t, "Store on a closed database")

		// Build
		db, _ := NewStorage(storageDB)
		_ = db.Close()
		r := NewMockRegistration()
		r.OutDevEUI[5] = 9

		// Operate
		err := db.Store(r)

		// Checks
		CheckErrors(t, pointer.String(string(errors.Operational)), err)
	}

	// -------------------

	{
		Desc(t, "Lookup on a closed database")

		// Build
		db, _ := NewStorage(storageDB)
		_ = db.Close()
		devEUI := NewMockRegistration().DevEUI()

		// Operate
		entries, err := db.Lookup(devEUI)

		// Checks
		CheckErrors(t, pointer.String(string(errors.Operational)), err)
		CheckEntries(t, nil, entries)
	}

	// -------------------

	{
		Desc(t, "Store an invalid recipient")

		// Build
		db, _ := NewStorage(storageDB)
		r := NewMockRegistration()
		r.OutDevEUI[7] = 99
		r.OutRecipient.(*MockRecipient).Failures["MarshalBinary"] = errors.New(errors.Structural, "Mock Error: MarshalBinary")

		// Operate & Check
		err := db.Store(r)
		CheckErrors(t, pointer.String(string(errors.Structural)), err)
		entries, err := db.Lookup(r.DevEUI())
		CheckErrors(t, pointer.String(string(errors.Behavioural)), err)
		CheckEntries(t, nil, entries)

		_ = db.Close()
	}
}

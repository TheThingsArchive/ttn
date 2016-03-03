// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package router

import (
	"os"
	"path"
	"testing"
	"time"

	. "github.com/TheThingsNetwork/ttn/core/mocks"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	. "github.com/TheThingsNetwork/ttn/utils/errors/checks"
	"github.com/TheThingsNetwork/ttn/utils/pointer"
	. "github.com/TheThingsNetwork/ttn/utils/testing"
)

const storageDB = "TestRouterStorage.db"

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
		Desc(t, "Store then lookup a registration")

		// Build
		db, _ := NewStorage(storageDB, time.Hour)
		r := NewMockRegistration()

		// Operate
		err := db.Store(r)
		CheckErrors(t, nil, err)
		gotEntry, err := db.Lookup(r.DevEUI())

		// Expectations
		wantEntry := entry{
			Recipient: r.RawRecipient(),
			until:     time.Now().Add(time.Hour),
		}

		// Check
		CheckErrors(t, nil, err)
		CheckEntries(t, wantEntry, gotEntry)
		_ = db.Close()
	}

	// ------------------

	{
		Desc(t, "Store an existing entry")

		// Build
		db, _ := NewStorage(storageDB, time.Hour)
		r := NewMockRegistration()

		// Operate
		err := db.Store(r)
		gotEntry, _ := db.Lookup(r.DevEUI())

		// Expectations
		wantEntry := entry{
			Recipient: r.RawRecipient(),
			until:     time.Now().Add(time.Hour),
		}

		// Checks
		CheckErrors(t, pointer.String(string(errors.Structural)), err)
		CheckEntries(t, wantEntry, gotEntry)
		_ = db.Close()
	}

	// ------------------

	{
		Desc(t, "Lookup non-existing entry")

		// Build
		db, _ := NewStorage(storageDB, time.Hour)
		devEUI := NewMockRegistration().DevEUI()
		devEUI[1] = 14

		// Operate
		gotEntry, err := db.Lookup(devEUI)

		// Checks
		CheckErrors(t, pointer.String(string(errors.Behavioural)), err)
		CheckEntries(t, entry{}, gotEntry)
		_ = db.Close()
	}

	// ------------------

	{
		Desc(t, "Lookup an expired entry")

		// Build
		db, _ := NewStorage(storageDB, time.Millisecond*100)
		r := NewMockRegistration()
		r.OutDevEUI[0] = 12

		// Operate
		_ = db.Store(r)
		<-time.After(time.Millisecond * 200)
		gotEntry, err := db.Lookup(r.DevEUI())

		// Checks
		CheckErrors(t, pointer.String(string(errors.Behavioural)), err)
		CheckEntries(t, entry{}, gotEntry)
		_ = db.Close()
	}

	// ------------------

	{
		Desc(t, "Store above an expired entry")

		// Build
		db, _ := NewStorage(storageDB, time.Millisecond*100)
		r := NewMockRegistration()
		r.OutDevEUI[4] = 27

		// Operate
		_ = db.Store(r)
		<-time.After(time.Millisecond * 200)
		err := db.Store(r)
		CheckErrors(t, nil, err)
		gotEntry, err := db.Lookup(r.DevEUI())

		// Expectations
		wantEntry := entry{
			Recipient: r.RawRecipient(),
			until:     time.Now().Add(time.Millisecond * 200),
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
		r := NewMockRegistration()
		r.OutDevEUI[5] = 9

		// Operate
		err := db.Store(r)

		// Checks
		CheckErrors(t, pointer.String(string(errors.Operational)), err)
	}

	// ------------------

	{
		Desc(t, "Lookup on a closed database")

		// Build
		db, _ := NewStorage(storageDB, time.Hour)
		_ = db.Close()
		devEUI := NewMockRegistration().DevEUI()

		// Operate
		gotEntry, err := db.Lookup(devEUI)

		// Checks
		CheckErrors(t, pointer.String(string(errors.Operational)), err)
		CheckEntries(t, entry{}, gotEntry)
	}

	// ------------------

	{
		Desc(t, "Store an invalid recipient")

		// Build
		db, _ := NewStorage(storageDB, time.Hour)
		r := NewMockRegistration()
		r.OutDevEUI[7] = 99
		r.OutRecipient.(*MockRecipient).Failures["MarshalBinary"] = errors.New(errors.Structural, "Mock Error: MarshalBinary")

		// Operate & Check
		err := db.Store(r)
		CheckErrors(t, pointer.String(string(errors.Structural)), err)
		gotEntry, err := db.Lookup(r.DevEUI())
		CheckErrors(t, pointer.String(string(errors.Behavioural)), err)
		CheckEntries(t, entry{}, gotEntry)

		_ = db.Close()
	}
}

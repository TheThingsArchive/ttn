// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package handler

import (
	"os"
	"path"
	"testing"

	. "github.com/TheThingsNetwork/ttn/core/mocks"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	. "github.com/TheThingsNetwork/ttn/utils/errors/checks"
	"github.com/TheThingsNetwork/ttn/utils/pointer"
	. "github.com/TheThingsNetwork/ttn/utils/testing"
	"github.com/brocaar/lorawan"
)

const devDB = "TestDevStorage.db"

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
		r := NewMockHRegistration()

		// Operate
		err := db.StorePersonalized(r)
		CheckErrors(t, nil, err)
		entry, err := db.Lookup(r.AppEUI(), r.DevEUI())

		// Check
		CheckErrors(t, nil, err)
		CheckEntries(t, r, entry)
	}

	// ------------------

	{
		Desc(t, "Lookup a non-existing registration")

		// Build
		r := NewMockHRegistration()
		r.OutAppEUI = lorawan.EUI64([8]byte{1, 2, 1, 2, 1, 2, 1, 2})

		// Operate
		_, err := db.Lookup(r.AppEUI(), r.DevEUI())

		// Check
		CheckErrors(t, pointer.String(string(errors.NotFound)), err)
	}

	// ------------------

	{
		Desc(t, "Store twice the same registration")

		// Build
		r := NewMockHRegistration()
		r.OutAppEUI = lorawan.EUI64([8]byte{1, 4, 1, 4, 1, 4, 1, 4})

		// Operate
		_ = db.StorePersonalized(r)
		err := db.StorePersonalized(r)
		CheckErrors(t, nil, err)
		entry, err := db.Lookup(r.AppEUI(), r.DevEUI())

		// Check
		CheckErrors(t, nil, err)
		CheckEntries(t, r, entry)
	}

	// ------------------

	{
		Desc(t, "Store Activated")

		// Build
		r := NewMockHRegistration()
		r.OutAppEUI = lorawan.EUI64([8]byte{6, 6, 6, 7, 8, 6, 7, 6})

		// Operate
		err := db.StoreActivated(r)

		// Check
		CheckErrors(t, pointer.String(string(errors.Implementation)), err)
	}

	// ------------------

	{
		Desc(t, "Close the storage")
		err := db.Close()
		CheckErrors(t, nil, err)
	}

}

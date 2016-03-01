// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package handler

import (
	"os"
	"reflect"
	"testing"
	"time"

	. "github.com/TheThingsNetwork/ttn/core"
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
		os.Remove(devDB)
	}()

	// ------------------

	{
		Desc(t, "Create a new storage")
		var err error
		db, err = NewDevStorage(devDB)
		CheckErrors(t, nil, err)
	}

	// ------------------

	{
		Desc(t, "Store and Lookup a registration")
		r := newTestRegistration(
			[8]byte{1, 1, 1, 1, 1, 1, 1, 1},
			[8]byte{2, 2, 2, 2, 2, 2, 2, 2},
		)

		err := db.StorePersonalized(r)
		CheckErrors(t, nil, err)
		entry, err := db.Lookup(r.AppEUI(), r.DevEUI())
		CheckErrors(t, nil, err)
		CheckEntries(t, r, entry)
	}

	// ------------------

	{
		Desc(t, "Lookup a non-existing registration")
		r := newTestRegistration(
			[8]byte{1, 1, 1, 1, 1, 1, 1, 2},
			[8]byte{2, 2, 2, 2, 2, 2, 2, 3},
		)
		_, err := db.Lookup(r.AppEUI(), r.DevEUI())
		CheckErrors(t, pointer.String(string(errors.Behavioural)), err)
	}

	// ------------------

	{
		Desc(t, "Store twice the same registration")
		r := newTestRegistration(
			[8]byte{3, 1, 1, 1, 1, 1, 1, 1},
			[8]byte{4, 2, 2, 2, 2, 2, 2, 2},
		)
		err := db.StorePersonalized(r)
		CheckErrors(t, nil, err)
		err = db.StorePersonalized(r)
		CheckErrors(t, nil, err)
		entry, err := db.Lookup(r.AppEUI(), r.DevEUI())
		CheckErrors(t, nil, err)
		CheckEntries(t, r, entry)
	}

	// ------------------

	{
		Desc(t, "Store Activated")
		r := newTestRegistration(
			[8]byte{3, 1, 1, 1, 1, 1, 1, 1},
			[8]byte{4, 2, 2, 2, 2, 2, 2, 2},
		)
		err := db.StoreActivated(r)
		CheckErrors(t, pointer.String(string(errors.Implementation)), err)
	}

	// ------------------

	{
		Desc(t, "Close the storage")
		err := db.Close()
		CheckErrors(t, nil, err)
	}

}

// ----- TYPE utilities
type testRegistration struct {
	appEUI lorawan.EUI64
	devEUI lorawan.EUI64
}

func newTestRegistration(appEUI [8]byte, devEUI [8]byte) testRegistration {
	r := testRegistration{}
	copy(r.appEUI[:], appEUI[:])
	copy(r.devEUI[:], devEUI[:])
	return r
}

func (r testRegistration) Recipient() Recipient {
	return time.Date(2016, 3, 1, 11, 13, 6, 0, time.UTC)
}

func (r testRegistration) RawRecipient() []byte {
	data, _ := r.Recipient().MarshalBinary()
	return data
}

func (r testRegistration) AppEUI() lorawan.EUI64 {
	return r.appEUI
}

func (r testRegistration) DevEUI() lorawan.EUI64 {
	return r.devEUI
}

func (r testRegistration) DevAddr() lorawan.DevAddr {
	devAddr := lorawan.DevAddr{}
	copy(devAddr[:], r.devEUI[4:])
	return devAddr
}

func (r testRegistration) NwkSKey() lorawan.AES128Key {
	return lorawan.AES128Key([16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6})
}

func (r testRegistration) AppSKey() lorawan.AES128Key {
	return lorawan.AES128Key([16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5, 6})
}

// ----- CHECK utilities
func CheckEntries(t *testing.T, want testRegistration, got devEntry) {
	wantEntry := devEntry{
		Recipient: want.RawRecipient(),
		DevAddr:   want.DevAddr(),
		NwkSKey:   want.NwkSKey(),
		AppSKey:   want.AppSKey(),
	}

	if reflect.DeepEqual(wantEntry, got) {
		Ok(t, "Check Entries")
		return
	}
	Ko(t, "Check Entries")
}

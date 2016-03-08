// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package storage

import (
	"os"
	"path"
	"reflect"
	"testing"

	"github.com/TheThingsNetwork/ttn/utils/errors"
	. "github.com/TheThingsNetwork/ttn/utils/errors/checks"
	"github.com/TheThingsNetwork/ttn/utils/pointer"
	. "github.com/TheThingsNetwork/ttn/utils/testing"
)

const database = "TestStoreAndLookup.db"

func TestStoreAndLookup(t *testing.T) {
	var itf Interface
	defer func() {
		if itf == nil {
			return
		}
		if err := itf.Close(); err != nil {
			panic(err)
		}
		os.Remove(path.Join(os.TempDir(), database))
	}()

	// --------------------

	{
		Desc(t, "Open database in tmp folder")
		var err error
		itf, err = New(path.Join(os.TempDir(), database))
		CheckErrors(t, nil, err)
	}

	// --------------------

	{
		Desc(t, "Open database in a forbidden place")
		_, err := New("/usr/bin")
		CheckErrors(t, pointer.String(string(errors.Operational)), err)
	}

	// --------------------

	{
		Desc(t, "Store then lookup in a 1-level bucket")
		err := itf.Store("bucket", []byte{1, 2, 3}, []Entry{&testEntry{Data: "TTN"}})
		CheckErrors(t, nil, err)

		entries, err := itf.Lookup("bucket", []byte{1, 2, 3}, &testEntry{})
		CheckErrors(t, nil, err)
		CheckEntries(t, []testEntry{testEntry{Data: "TTN"}}, entries)
	}

	// -------------------

	{
		Desc(t, "Store then lookup in a nested bucket")
		err := itf.Store("nested.bucket", []byte{14, 42}, []Entry{&testEntry{Data: "IoT"}})
		CheckErrors(t, nil, err)

		entries, err := itf.Lookup("nested.bucket", []byte{14, 42}, &testEntry{})
		CheckErrors(t, nil, err)
		CheckEntries(t, []testEntry{testEntry{Data: "IoT"}}, entries)
	}

	// -------------------

	{
		Desc(t, "Lookup in non-existing bucket")
		entries, err := itf.Lookup("DoesntExist", []byte{1, 2, 3}, &testEntry{})
		CheckErrors(t, pointer.String(string(errors.NotFound)), err)
		CheckEntries(t, nil, entries)
	}

	// -------------------

	{
		Desc(t, "Lookup a non-existing key")
		entries, err := itf.Lookup("bucket", []byte{9, 9, 9, 9, 9}, &testEntry{})
		CheckErrors(t, pointer.String(string(errors.NotFound)), err)
		CheckEntries(t, nil, entries)
	}

	// -------------------

	{
		Desc(t, "Flush an 1-level bucket entry")
		itf.Store("bucket", []byte{1, 1, 1}, []Entry{&testEntry{Data: "TTN"}})
		err := itf.Flush("bucket", []byte{1, 1, 1})
		CheckErrors(t, nil, err)
		entries, err := itf.Lookup("bucket", []byte{1, 1, 1}, &testEntry{})
		CheckErrors(t, pointer.String(string(errors.NotFound)), err)
		CheckEntries(t, nil, entries)
	}

	// -------------------

	{
		Desc(t, "Flush a nested bucket entry")
		itf.Store("nested.bucket", []byte{2, 2, 2}, []Entry{&testEntry{Data: "TTN"}})
		err := itf.Flush("nested.bucket", []byte{2, 2, 2})
		CheckErrors(t, nil, err)
		entries, err := itf.Lookup("nested.bucket", []byte{2, 2, 2}, &testEntry{})
		CheckErrors(t, pointer.String(string(errors.NotFound)), err)
		CheckEntries(t, nil, entries)
	}

	// -------------------

	{
		Desc(t, "Reset a 1-level bucket")
		itf.Store("mybucket", []byte{1, 1, 1}, []Entry{&testEntry{Data: "TTN"}})
		err := itf.Reset("mybucket")
		CheckErrors(t, nil, err)
		entries, err := itf.Lookup("mybucket", []byte{1, 1, 1}, &testEntry{})
		CheckErrors(t, pointer.String(string(errors.NotFound)), err)
		CheckEntries(t, nil, entries)
	}

	// -------------------

	{
		Desc(t, "Reset a nested bucket")
		itf.Store("mybucket.nested", []byte{2, 2, 2}, []Entry{&testEntry{Data: "TTN"}})
		err := itf.Reset("mybucket.nested")
		CheckErrors(t, nil, err)
		entries, err := itf.Lookup("mybucket.nested", []byte{2, 2, 2}, &testEntry{})
		CheckErrors(t, pointer.String(string(errors.NotFound)), err)
		CheckEntries(t, nil, entries)
	}

	// -------------------

	{
		Desc(t, "Reset a nested bucket parent")
		itf.Store("mybucket.nested", []byte{2, 2, 2}, []Entry{&testEntry{Data: "TTN"}})
		err := itf.Reset("mybucket")
		CheckErrors(t, nil, err)
		entries, err := itf.Lookup("mybucket.nested", []byte{2, 2, 2}, &testEntry{})
		CheckErrors(t, pointer.String(string(errors.NotFound)), err)
		CheckEntries(t, nil, entries)
	}

	// -------------------

	{
		Desc(t, "Replace an existing entry in a bucket")
		itf.Store("anotherbucket", []byte{14, 14, 14}, []Entry{&testEntry{Data: "I don't like IoT"}})
		err := itf.Replace("anotherbucket", []byte{14, 14, 14}, []Entry{&testEntry{Data: "IoT is Awesome"}})
		CheckErrors(t, nil, err)
		entries, err := itf.Lookup("anotherbucket", []byte{14, 14, 14}, &testEntry{})
		CheckErrors(t, nil, err)
		CheckEntries(t, []testEntry{{Data: "IoT is Awesome"}}, entries)
	}

	// -------------------

	{
		Desc(t, "Store several entries under the same key")
		itf.Store("several", []byte{1, 1, 1}, []Entry{&testEntry{Data: "FirstEntry"}})
		itf.Store("several", []byte{1, 1, 1}, []Entry{&testEntry{Data: "SecondEntry"}, &testEntry{Data: "ThirdEntry"}})
		entries, err := itf.Lookup("several", []byte{1, 1, 1}, &testEntry{})
		CheckErrors(t, nil, err)
		CheckEntries(t, []testEntry{{Data: "FirstEntry"}, {Data: "SecondEntry"}, {Data: "ThirdEntry"}}, entries)
	}
}

// ----- Type Utilities

type testEntry struct {
	Data string
}

func (e testEntry) MarshalBinary() ([]byte, error) {
	return []byte(e.Data), nil
}

func (e *testEntry) UnmarshalBinary(data []byte) error {
	e.Data = string(data)
	return nil
}

// ----- Check Utilities

func CheckEntries(t *testing.T, want []testEntry, got interface{}) {
	if want == nil && got == nil {
		Ok(t, "Check Entries")
		return
	}

	entries, ok := got.([]testEntry)
	if !ok {
		Ko(t, "Expected []testEntry but got %+v", got)
		return
	}

	if !reflect.DeepEqual(want, entries) {
		Ko(t, "Retrieved entries don't match expectations.\nWant: %v\nGot:  %v", want, entries)
		return
	}
	Ok(t, "Check Entries")
}

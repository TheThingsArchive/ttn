// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package storage

import (
	"encoding"
	"os"
	"path"
	"testing"

	. "github.com/TheThingsNetwork/ttn/utils/testing"
)

const database = "TestStoreAndLookup.db"

func TestStoreAndRead(t *testing.T) {
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
		CheckErrors(t, ErrOperational, err)
	}

	// --------------------

	{
		Desc(t, "Store then lookup in a 1-level bucket")
		err := itf.Append([]byte{1, 2, 3}, []encoding.BinaryMarshaler{testEntry{Data: "TTN"}}, []byte("bucket"))
		FatalUnless(t, err)
		entries, err := itf.Read([]byte{1, 2, 3}, &testEntry{}, []byte("bucket"))

		CheckErrors(t, nil, err)
		Check(t, []testEntry{testEntry{Data: "TTN"}}, entries.([]testEntry), "Entries")
	}

	// -------------------

	{
		Desc(t, "Store then lookup in a nested bucket")
		err := itf.Append([]byte{14, 42}, []encoding.BinaryMarshaler{&testEntry{Data: "IoT"}}, []byte("nested"), []byte("bucket"))
		FatalUnless(t, err)
		entries, err := itf.Read([]byte{14, 42}, &testEntry{}, []byte("nested"), []byte("bucket"))

		CheckErrors(t, nil, err)
		Check(t, []testEntry{testEntry{Data: "IoT"}}, entries.([]testEntry), "Entries")
	}

	// -------------------

	{
		Desc(t, "Store then lookup in a nested bucket with default key")
		err := itf.Append(nil, []encoding.BinaryMarshaler{&testEntry{Data: "IoT"}}, []byte("nested"), []byte("buckettt"))
		FatalUnless(t, err)
		entries, err := itf.Read(nil, &testEntry{}, []byte("nested"), []byte("buckettt"))

		CheckErrors(t, nil, err)
		Check(t, []testEntry{testEntry{Data: "IoT"}}, entries.([]testEntry), "Entries")
	}

	// -------------------

	{
		Desc(t, "Lookup in non-existing bucket")
		entries, err := itf.Read([]byte{1, 2, 3}, &testEntry{}, []byte("DoesnExists"))

		CheckErrors(t, ErrNotFound, err)
		Check(t, nil, entries, "Entries")
	}

	// -------------------

	{
		Desc(t, "Lookup a non-existing key")
		entries, err := itf.Read([]byte{9, 9, 9, 9, 9}, &testEntry{}, []byte("bucket"))

		CheckErrors(t, ErrNotFound, err)
		Check(t, nil, entries, "Entries")
	}

	// -------------------

	{
		Desc(t, "Flush an 1-level bucket entry")
		itf.Append([]byte{1, 1, 1}, []encoding.BinaryMarshaler{&testEntry{Data: "TTN"}}, []byte("bucket"))
		err := itf.Delete([]byte{1, 1, 1}, []byte("bucket"))
		FatalUnless(t, err)
		entries, err := itf.Read([]byte{1, 1, 1}, &testEntry{}, []byte("bucket"))

		CheckErrors(t, ErrNotFound, err)
		Check(t, nil, entries, "Entries")
	}

	// -------------------

	{
		Desc(t, "Flush a nested bucket entry")
		itf.Append([]byte{2, 2, 2}, []encoding.BinaryMarshaler{&testEntry{Data: "TTN"}}, []byte("nested"), []byte("bucket"))
		err := itf.Delete([]byte{2, 2, 2}, []byte("nested"), []byte("bucket"))
		FatalUnless(t, err)
		entries, err := itf.Read([]byte{2, 2, 2}, &testEntry{}, []byte("nested"), []byte("bucket"))

		CheckErrors(t, ErrNotFound, err)
		Check(t, nil, entries, "Entries")
	}

	// -------------------

	{
		Desc(t, "Reset a 1-level bucket")
		itf.Append([]byte{1, 1, 1}, []encoding.BinaryMarshaler{testEntry{Data: "TTN"}}, []byte("mybucket"))
		err := itf.Reset([]byte("mybucket"))
		FatalUnless(t, err)
		entries, err := itf.Read([]byte{1, 1, 1}, &testEntry{}, []byte("mybucket"))

		CheckErrors(t, ErrNotFound, err)
		Check(t, nil, entries, "Entries")
	}

	// -------------------

	{
		Desc(t, "Reset a nested bucket")
		itf.Append([]byte{2, 2, 2}, []encoding.BinaryMarshaler{&testEntry{Data: "TTN"}}, []byte("mybucket"), []byte("nested"))
		err := itf.Reset([]byte("mybucket"), []byte("nested"))
		FatalUnless(t, err)
		entries, err := itf.Read([]byte{2, 2, 2}, &testEntry{}, []byte("mybucket"), []byte("nested"))

		CheckErrors(t, ErrNotFound, err)
		Check(t, nil, entries, "Entries")
	}

	// -------------------

	{
		Desc(t, "Reset a nested bucket parent")
		itf.Append([]byte{2, 2, 2}, []encoding.BinaryMarshaler{&testEntry{Data: "TTN"}}, []byte("mybucket"), []byte("nested"))
		err := itf.Reset([]byte("mybucket"))
		FatalUnless(t, err)
		entries, err := itf.Read([]byte{2, 2, 2}, &testEntry{}, []byte("mybucket"), []byte("nested"))

		CheckErrors(t, ErrNotFound, err)
		Check(t, nil, entries, "Entries")
	}

	// -------------------

	{
		Desc(t, "Replace an existing entry in a bucket")
		itf.Append([]byte{14, 14, 14}, []encoding.BinaryMarshaler{&testEntry{Data: "I don't like IoT"}}, []byte("anotherbucket"))
		err := itf.Update([]byte{14, 14, 14}, []encoding.BinaryMarshaler{testEntry{Data: "IoT is Awesome"}}, []byte("anotherbucket"))
		FatalUnless(t, err)
		entries, err := itf.Read([]byte{14, 14, 14}, &testEntry{}, []byte("anotherbucket"))

		CheckErrors(t, nil, err)
		Check(t, []testEntry{{Data: "IoT is Awesome"}}, entries.([]testEntry), "Entries")
	}

	// -------------------

	{
		Desc(t, "Store several entries under the same key")
		itf.Append([]byte{1, 1, 1}, []encoding.BinaryMarshaler{&testEntry{Data: "FirstEntry"}}, []byte("several"))
		itf.Append([]byte{1, 1, 1}, []encoding.BinaryMarshaler{&testEntry{Data: "SecondEntry"}, &testEntry{Data: "ThirdEntry"}}, []byte("several"))
		entries, err := itf.Read([]byte{1, 1, 1}, &testEntry{}, []byte("several"))

		CheckErrors(t, nil, err)
		Check(t, []testEntry{{Data: "FirstEntry"}, {Data: "SecondEntry"}, {Data: "ThirdEntry"}}, entries.([]testEntry), "Entries")
	}

	// --------------------

	{
		Desc(t, "Store, Read, Update and Delete with default key")
		err := itf.Append(nil, []encoding.BinaryMarshaler{&testEntry{Data: "Patate"}}, []byte("defaultkeybucket"))
		CheckErrors(t, nil, err)
		_, err = itf.Read(nil, &testEntry{}, []byte("defaultkeybucket"))
		CheckErrors(t, nil, err)
		err = itf.Update(nil, []encoding.BinaryMarshaler{&testEntry{Data: "TTN"}}, []byte("defaultkeybucket"))
		CheckErrors(t, nil, err)
		err = itf.Delete(nil, []byte("defaultkeybucket"))
		CheckErrors(t, nil, err)
	}

	// --------------------

	{
		Desc(t, "Store, Read, Update, Delete & Reset with no bucket names")
		err := itf.Append([]byte{1, 2, 3}, []encoding.BinaryMarshaler{&testEntry{Data: "Patate"}})
		CheckErrors(t, ErrStructural, err)
		_, err = itf.Read([]byte{1, 2, 3}, &testEntry{})
		CheckErrors(t, ErrStructural, err)
		err = itf.Update([]byte{1, 2, 3}, []encoding.BinaryMarshaler{&testEntry{Data: "TTN"}})
		CheckErrors(t, ErrStructural, err)
		err = itf.Delete([]byte{1, 2, 3})
		CheckErrors(t, ErrStructural, err)
		err = itf.Reset()
		CheckErrors(t, ErrStructural, err)
	}

	// ---------------------

	{
		Desc(t, "Read All entry of a bucket")
		err := itf.Update([]byte{0, 0, 1}, []encoding.BinaryMarshaler{&testEntry{Data: "The"}}, []byte("level1"))
		FatalUnless(t, err)
		err = itf.Update([]byte{0, 0, 2}, []encoding.BinaryMarshaler{&testEntry{Data: "Things"}}, []byte("level1"))
		FatalUnless(t, err)
		err = itf.Update([]byte{0, 0, 3}, []encoding.BinaryMarshaler{&testEntry{Data: "Network"}}, []byte("level1"))
		FatalUnless(t, err)
		entries, err := itf.ReadAll(&testEntry{}, []byte("level1"))
		want := []testEntry{{Data: "The"}, {Data: "Things"}, {Data: "Network"}}
		CheckErrors(t, nil, err)
		Check(t, want, entries, "Entries")
	}

	// ---------------------

	{
		Desc(t, "Store, Read, Update, Delete & Reset on closed storage")
		_ = itf.Close()
		err := itf.Append([]byte{1, 2, 3}, []encoding.BinaryMarshaler{&testEntry{Data: "Patate"}}, []byte("closeddb"))
		CheckErrors(t, ErrOperational, err)
		_, err = itf.Read([]byte{1, 2, 3}, &testEntry{}, []byte("closeddb"))
		CheckErrors(t, ErrOperational, err)
		_, err = itf.ReadAll(&testEntry{}, []byte("closeddb"))
		CheckErrors(t, ErrOperational, err)
		err = itf.Update([]byte{1, 2, 3}, []encoding.BinaryMarshaler{&testEntry{Data: "TTN"}}, []byte("closeddb"))
		CheckErrors(t, ErrOperational, err)
		err = itf.Delete([]byte{1, 2, 3}, []byte("closeddb"))
		CheckErrors(t, ErrOperational, err)
		err = itf.Reset([]byte("closeddb"))
		CheckErrors(t, ErrOperational, err)
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

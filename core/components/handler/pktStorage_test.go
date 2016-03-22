// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package handler

import (
	"os"
	"path"
	"testing"
	"time"

	. "github.com/TheThingsNetwork/ttn/utils/testing"
)

const pktDB = "TestPktStorage.db"

func TestPullSize1(t *testing.T) {
	var db PktStorage
	defer func() {
		os.Remove(path.Join(os.TempDir(), pktDB))
	}()

	// ------------------

	{
		Desc(t, "Create a new storage")
		var err error
		db, err = NewPktStorage(path.Join(os.TempDir(), pktDB), 1)
		CheckErrors(t, nil, err)
	}

	// ------------------

	{
		Desc(t, "Dequeue an empty area")
		got, err := db.dequeue([]byte{1, 2}, []byte{3, 4})
		CheckErrors(t, ErrNotFound, err)
		Check(t, pktEntry{}, got, "Packet entries")
	}

	// ------------------

	{

		Desc(t, "Peek an empty area")
		got, err := db.peek([]byte{1, 2}, []byte{3, 4})
		CheckErrors(t, ErrNotFound, err)
		Check(t, pktEntry{}, got, "Packet entries")
	}

	// ------------------

	{
		Desc(t, "Queue / Dequeue an entry")
		entry := pktEntry{AppEUI: []byte{1, 2}, DevEUI: []byte{3, 4}, TTL: time.Now().Add(time.Hour), Payload: []byte{14, 42}}
		err := db.enqueue(entry)
		FatalUnless(t, err)
		got, err := db.dequeue(entry.AppEUI, entry.DevEUI)
		FatalUnless(t, err)
		Check(t, entry, got, "Packet entries")
		_, err = db.dequeue(entry.AppEUI, entry.DevEUI)
		CheckErrors(t, ErrNotFound, err)
	}

	// ------------------

	{
		Desc(t, "Queue / Peek an entry")
		entry := pktEntry{AppEUI: []byte{1, 2}, DevEUI: []byte{3, 4}, TTL: time.Now().Add(time.Hour), Payload: []byte{14, 42}}
		err := db.enqueue(entry)
		FatalUnless(t, err)
		got, err := db.peek(entry.AppEUI, entry.DevEUI)
		FatalUnless(t, err)
		Check(t, entry, got, "Packet entries")
		got, err = db.peek(entry.AppEUI, entry.DevEUI)
		FatalUnless(t, err)
		Check(t, entry, got, "Packet entries")
	}

	// ------------------

	{
		Desc(t, "Queue on an existing entry")
		entry1 := pktEntry{AppEUI: []byte{1, 2}, DevEUI: []byte{3, 4}, TTL: time.Now().Add(time.Hour), Payload: []byte{14, 42}}
		entry2 := pktEntry{AppEUI: []byte{1, 2}, DevEUI: []byte{3, 4}, TTL: time.Now().Add(time.Hour), Payload: []byte{1, 2, 3, 4}}
		err := db.enqueue(entry1)
		FatalUnless(t, err)
		err = db.enqueue(entry2)
		FatalUnless(t, err)
		got, err := db.dequeue(entry1.AppEUI, entry1.DevEUI)
		FatalUnless(t, err)
		Check(t, entry2, got, "Packet entries")
	}

	// ------------------

	{
		Desc(t, "Queue / Wait expiry / Dequeue")
		entry := pktEntry{AppEUI: []byte{1, 2}, DevEUI: []byte{3, 4}, TTL: time.Now(), Payload: []byte{14, 42}}
		err := db.enqueue(entry)
		FatalUnless(t, err)
		<-time.After(time.Millisecond * 10)
		_, err = db.dequeue(entry.AppEUI, entry.DevEUI)
		CheckErrors(t, ErrNotFound, err)
	}

	// ------------------

	{
		Desc(t, "Queue / Wait expiry / Peek")
		entry := pktEntry{AppEUI: []byte{1, 2}, DevEUI: []byte{3, 4}, TTL: time.Now(), Payload: []byte{14, 42}}
		err := db.enqueue(entry)
		FatalUnless(t, err)
		<-time.After(time.Millisecond * 10)
		_, err = db.peek(entry.AppEUI, entry.DevEUI)
		CheckErrors(t, ErrNotFound, err)
	}

	// ------------------

	{
		Desc(t, "Close")
		err := db.done()
		CheckErrors(t, nil, err)
	}
}

func TestPullSize2(t *testing.T) {
	var db PktStorage
	defer func() {
		os.Remove(path.Join(os.TempDir(), pktDB))
	}()

	// ------------------

	{
		Desc(t, "Create a new storage")
		var err error
		db, err = NewPktStorage(path.Join(os.TempDir(), pktDB), 2)
		CheckErrors(t, nil, err)
	}

	// ------------------

	{
		Desc(t, "Queue two, then Dequeue two")
		entry1 := pktEntry{AppEUI: []byte{1, 2}, DevEUI: []byte{3, 4}, TTL: time.Now().Add(time.Hour), Payload: []byte{42}}
		entry2 := pktEntry{AppEUI: []byte{1, 2}, DevEUI: []byte{3, 4}, TTL: time.Now().Add(time.Hour), Payload: []byte{14}}
		err := db.enqueue(entry1)
		FatalUnless(t, err)
		err = db.enqueue(entry2)
		FatalUnless(t, err)
		got1, err := db.dequeue(entry1.AppEUI, entry1.DevEUI)
		FatalUnless(t, err)
		got2, err := db.dequeue(entry2.AppEUI, entry2.DevEUI)
		FatalUnless(t, err)
		Check(t, entry1, got1, "Packet Entries")
		Check(t, entry2, got2, "Packet Entries")
	}

	// ------------------

	{
		Desc(t, "Queue two, wait first expires, then peek")
		entry1 := pktEntry{AppEUI: []byte{1, 2}, DevEUI: []byte{3, 4}, TTL: time.Now(), Payload: []byte{42}}
		entry2 := pktEntry{AppEUI: []byte{1, 2}, DevEUI: []byte{3, 4}, TTL: time.Now().Add(time.Hour), Payload: []byte{14}}
		err := db.enqueue(entry1)
		FatalUnless(t, err)
		err = db.enqueue(entry2)
		FatalUnless(t, err)
		<-time.After(time.Millisecond * 10)
		got, err := db.peek(entry2.AppEUI, entry2.DevEUI)
		CheckErrors(t, nil, err)
		Check(t, entry2, got, "Packet Entries")
	}

	// ------------------

	{
		Desc(t, "Queue three, dequeue then peek")
		entry1 := pktEntry{AppEUI: []byte{1, 2}, DevEUI: []byte{3, 4}, TTL: time.Now().Add(time.Hour), Payload: []byte{42}}
		entry2 := pktEntry{AppEUI: []byte{1, 2}, DevEUI: []byte{3, 4}, TTL: time.Now().Add(time.Hour), Payload: []byte{14}}
		entry3 := pktEntry{AppEUI: []byte{1, 2}, DevEUI: []byte{3, 4}, TTL: time.Now().Add(time.Hour), Payload: []byte{6}}
		err := db.enqueue(entry1)
		FatalUnless(t, err)
		err = db.enqueue(entry2)
		FatalUnless(t, err)
		err = db.enqueue(entry3)
		FatalUnless(t, err)
		got1, err := db.dequeue(entry1.AppEUI, entry1.DevEUI)
		FatalUnless(t, err)
		got2, err := db.peek(entry2.AppEUI, entry2.DevEUI)
		FatalUnless(t, err)
		Check(t, entry2, got1, "Packet Entries")
		Check(t, entry3, got2, "Packet Entries")
	}

	// ------------------

	{
		Desc(t, "Close")
		err := db.done()
		CheckErrors(t, nil, err)
	}
}

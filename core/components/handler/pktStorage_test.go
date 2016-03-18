// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package handler

import (
	"os"
	"path"
	"testing"

	. "github.com/TheThingsNetwork/ttn/utils/testing"
)

const pktDB = "TestPktStorage.db"

func TestPushPullNormal(t *testing.T) {
	var db PktStorage
	defer func() {
		os.Remove(path.Join(os.TempDir(), pktDB))
	}()

	// ------------------

	{
		Desc(t, "Create a new storage")
		var err error
		db, err = NewPktStorage(path.Join(os.TempDir(), pktDB))
		CheckErrors(t, nil, err)
	}

	// ------------------

	{
		Desc(t, "Push and Pull a valid Payload")

		// Build
		appEUI := []byte{1, 1, 1, 1, 1, 1, 1, 1}
		devEUI := []byte{2, 2, 2, 2, 2, 2, 2, 2}
		payload := pktEntry{[]byte("TheThingsNetwork")}

		// Expects
		var want = payload

		// Operate
		err := db.Push(appEUI, devEUI, payload)
		FatalUnless(t, err)
		p, err := db.Pull(appEUI, devEUI)

		// Check
		CheckErrors(t, nil, err)
		Check(t, want, p, "Payloads")
	}

	// ------------------

	{
		Desc(t, "Push two payloads -> same device")

		// Build
		appEUI := []byte{1, 1, 1, 1, 1, 1, 1, 2}
		devEUI := []byte{2, 2, 2, 2, 2, 2, 2, 3}
		payload1 := pktEntry{[]byte("TheThingsNetwork1")}
		payload2 := pktEntry{[]byte("TheThingsNetwork2")}

		// Expects
		var want1 = payload1
		var want2 = payload2

		// Operate
		err := db.Push(appEUI, devEUI, payload1)
		FatalUnless(t, err)
		err = db.Push(appEUI, devEUI, payload2)
		FatalUnless(t, err)
		p1, err := db.Pull(appEUI, devEUI)
		FatalUnless(t, err)
		p2, err := db.Pull(appEUI, devEUI)
		FatalUnless(t, err)

		// Check
		Check(t, want1, p1, "Payloads")
		Check(t, want2, p2, "Payloads")
	}

	// ------------------

	{
		Desc(t, "Pull a non existing entry")

		// Build
		appEUI := []byte{1, 1, 1, 1, 1, 1, 1, 3}
		devEUI := []byte{2, 2, 2, 2, 2, 2, 2, 3}

		// Operate
		_, err := db.Pull(appEUI, devEUI)

		// Check
		CheckErrors(t, ErrNotFound, err)
	}

	// ------------------

	{
		Desc(t, "Close the storage")
		err := db.Close()
		CheckErrors(t, nil, err)
	}

	// ------------------

	{
		Desc(t, "Push after close")

		// Build
		appEUI := []byte{1, 1, 1, 1, 1, 1, 1, 5}
		devEUI := []byte{2, 2, 2, 2, 2, 2, 2, 6}
		payload := pktEntry{[]byte("TheThingsNetwork")}

		// Operate
		err := db.Push(appEUI, devEUI, payload)

		// Check
		CheckErrors(t, ErrOperational, err)
	}

	// ------------------

	{
		Desc(t, "Pull after close")

		// Build
		appEUI := []byte{1, 1, 1, 1, 1, 1, 1, 1}
		devEUI := []byte{2, 2, 2, 2, 2, 2, 2, 2}

		// Operate
		_, err := db.Pull(appEUI, devEUI)

		// Check
		CheckErrors(t, ErrOperational, err)
	}
}

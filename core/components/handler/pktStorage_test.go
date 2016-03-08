// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package handler

import (
	"os"
	"path"
	"testing"

	. "github.com/TheThingsNetwork/ttn/core"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	. "github.com/TheThingsNetwork/ttn/utils/errors/checks"
	"github.com/TheThingsNetwork/ttn/utils/pointer"
	. "github.com/TheThingsNetwork/ttn/utils/testing"
	"github.com/brocaar/lorawan"
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
		Desc(t, "Push and Pull a valid APacket")

		// Build
		p, _ := NewAPacket(
			lorawan.EUI64([8]byte{1, 1, 1, 1, 1, 1, 1, 1}),
			lorawan.EUI64([8]byte{2, 2, 2, 2, 2, 2, 2, 2}),
			[]byte("TheThingsNetwork"),
			[]Metadata{},
		)

		// Operate
		err := db.Push(p)
		CheckErrors(t, nil, err)
		a, err := db.Pull(p.AppEUI(), p.DevEUI())

		// Check
		CheckErrors(t, nil, err)
		CheckPackets(t, p, a)
	}

	// ------------------

	{
		Desc(t, "Push two packets")

		// Build
		p1, _ := NewAPacket(
			lorawan.EUI64([8]byte{1, 1, 1, 1, 1, 1, 1, 1}),
			lorawan.EUI64([8]byte{2, 2, 2, 2, 2, 2, 2, 2}),
			[]byte("TheThingsNetwork1"),
			[]Metadata{},
		)
		p2, _ := NewAPacket(
			lorawan.EUI64([8]byte{1, 1, 1, 1, 1, 1, 1, 1}),
			lorawan.EUI64([8]byte{2, 2, 2, 2, 2, 2, 2, 2}),
			[]byte("TheThingsNetwork2"),
			[]Metadata{},
		)

		// Operate & Check
		err := db.Push(p1)
		CheckErrors(t, nil, err)
		err = db.Push(p2)
		CheckErrors(t, nil, err)

		a, err := db.Pull(p1.AppEUI(), p1.DevEUI())
		CheckErrors(t, nil, err)
		CheckPackets(t, p1, a)

		a, err = db.Pull(p1.AppEUI(), p1.DevEUI())
		CheckErrors(t, nil, err)
		CheckPackets(t, p2, a)
	}

	// ------------------

	{
		Desc(t, "Pull a non existing entry")

		// Build
		appEUI := lorawan.EUI64([8]byte{1, 2, 1, 2, 1, 2, 1, 2})
		devEUI := lorawan.EUI64([8]byte{2, 3, 4, 2, 3, 4, 2, 3})

		// Operate
		p, err := db.Pull(appEUI, devEUI)

		// Check
		CheckErrors(t, pointer.String(string(errors.NotFound)), err)
		CheckPackets(t, nil, p)
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
		p, _ := NewAPacket(
			lorawan.EUI64([8]byte{1, 1, 1, 1, 1, 1, 1, 1}),
			lorawan.EUI64([8]byte{2, 2, 2, 2, 2, 2, 2, 2}),
			[]byte("TheThingsNetwork"),
			[]Metadata{},
		)

		// Operate
		err := db.Push(p)

		// Check
		CheckErrors(t, pointer.String(string(errors.Operational)), err)
	}

	// ------------------

	{
		Desc(t, "Pull after close")

		// Build
		appEUI := lorawan.EUI64([8]byte{1, 2, 1, 2, 1, 2, 1, 2})
		devEUI := lorawan.EUI64([8]byte{2, 3, 4, 2, 3, 4, 2, 3})

		// Operate
		_, err := db.Pull(appEUI, devEUI)

		// Check
		CheckErrors(t, pointer.String(string(errors.Operational)), err)
	}
}

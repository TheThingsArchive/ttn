// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package handler

import (
	"os"
	"reflect"
	"testing"

	. "github.com/TheThingsNetwork/ttn/core"
	//"github.com/TheThingsNetwork/ttn/utils/errors"
	. "github.com/TheThingsNetwork/ttn/utils/errors/checks"
	. "github.com/TheThingsNetwork/ttn/utils/testing"
	"github.com/brocaar/lorawan"
)

const pktDB = "TestPktStorage.db"

func TestPushPullNormal(t *testing.T) {
	var db PktStorage
	defer func() {
		os.Remove(pktDB)
	}()

	// ------------------

	{
		Desc(t, "Create a new storage")
		var err error
		db, err = NewPktStorage(pktDB)
		CheckErrors(t, nil, err)
	}

	// ------------------

	{
		Desc(t, "Push and Pull a valid APacket")
		p, _ := NewAPacket(
			lorawan.EUI64([8]byte{1, 1, 1, 1, 1, 1, 1, 1}),
			lorawan.EUI64([8]byte{2, 2, 2, 2, 2, 2, 2, 2}),
			[]byte("TheThingsNetwork"),
			[]Metadata{},
		)
		err := db.Push(p)
		CheckErrors(t, nil, err)

		a, err := db.Pull(p.AppEUI(), p.DevEUI())
		CheckErrors(t, nil, err)
		CheckPackets(t, p, a)
	}

	// ------------------
	// ------------------
	// ------------------
	// ------------------
	// ------------------
	// ------------------
	// ------------------
	// ------------------
	// ------------------
	// ------------------
	// ------------------
}

func TestPushPullErrors(t *testing.T) {
	Ko(t, "TODO")
}

//type mockStorage struct {
//	Store(name string, key []byte, entries []Entry) error
//	Replace(name string, key []byte, entries []Entry) error
//	Lookup(name string, key []byte, shape Entry) (interface{}, error)
//	Flush(name string, key []byte) error
//	Reset(name string) error
//	Close() error
//}

// ----- CHECK utilities
func CheckPackets(t *testing.T, want APacket, got APacket) {
	if reflect.DeepEqual(want, got) {
		Ok(t, "Check Packets")
		return
	}
	Ko(t, "Obtained packet doesn't match expectations.\nWant: %s\nGot:  %s", want, got)
}

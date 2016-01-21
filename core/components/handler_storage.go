// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package components

import (
	"sync"

	"github.com/TheThingsNetwork/ttn/core"
	"github.com/brocaar/lorawan"
)

type handlerStorage interface {
	store(lorawan.DevAddr, handlerEntry) error
	partition([]core.Packet) ([]handlerPartition, error)
}

type handlerPartition struct {
	handlerEntry
	Packets []core.Packet
}

type handlerEntry struct {
	AppEUI  lorawan.EUI64
	NwkSKey lorawan.AES128Key
	AppSKey lorawan.AES128Key
	DevAddr lorawan.DevAddr
}

type handlerDB struct {
	sync.RWMutex // Guards entries
	entries      map[lorawan.DevAddr][]handlerEntry
}

// newHandlerDB construct a new local handlerStorage
func newHandlerDB() handlerStorage {
	return &handlerDB{entries: make(map[lorawan.DevAddr][]handlerEntry)}
}

// store implements the handlerStorage interface
func (db *handlerDB) store(devAddr lorawan.DevAddr, entry handlerEntry) error {
	db.Lock()
	db.entries[devAddr] = append(db.entries[devAddr], entry)
	db.Unlock()
	return nil
}

// partition implements the handlerStorage interface
func (db *handlerDB) partition(packets []core.Packet) ([]handlerPartition, error) {
	// Create a map in order to do the partition
	partitions := make(map[[20]byte]handlerPartition)

	db.RLock() // We require lock on the whole block because we don't want the entries to change while building the partition.
	for _, packet := range packets {
		// First, determine devAddr and get the macPayload. Those are mandatory.
		devAddr, err := packet.DevAddr()
		if err != nil {
			return nil, ErrInvalidPacket
		}

		// Now, get all tuples associated to that device address, and choose the right one
		for _, entry := range db.entries[devAddr] {
			// Compute MIC check to find the right keys
			ok, err := packet.Payload.ValidateMIC(entry.NwkSKey)
			if err != nil || !ok {
				continue // These aren't the droid you're looking for
			}

			// #Easy
			var id [20]byte
			copy(id[:16], entry.AppEUI[:])
			copy(id[16:], entry.DevAddr[:])
			partitions[id] = handlerPartition{
				handlerEntry: entry,
				Packets:      append(partitions[id].Packets, packet),
			}
			break // We shouldn't look for other entries, we've found the right one
		}
	}
	db.RUnlock()

	// Transform the map in a slice
	res := make([]handlerPartition, 0, len(partitions))
	for _, p := range partitions {
		res = append(res, p)
	}

	if len(res) == 0 {
		return nil, ErrNotFound
	}
	return res, nil
}

// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package components

import (
	"github.com/TheThingsNetwork/ttn/core"
	"github.com/boltdb/bolt"
	"github.com/brocaar/lorawan"
)

type handlerStorage interface {
	Lookup(devAddr lorawan.DevAddr) ([]handlerEntry, error)
	Store(devAddr lorawan.DevAddr, entry handlerEntry) error
	Partition(packet ...core.Packet) ([]handlerPartition, error)
}

type handlerBoltStorage struct {
	*bolt.DB
}

type handlerEntry struct {
	AppEUI  lorawan.EUI64
	AppSKey lorawan.AES128Key
	DevAddr lorawan.DevAddr
	NwkSKey lorawan.AES128Key
}

type handlerPartition struct {
	handlerEntry
	Id      partitionId
	Packets []core.Packet
}

type partitionId [20]byte

func (s handlerBoltStorage) Lookup(devAddr lorawan.DevAddr) ([]handlerEntry, error) {
	entries, err := lookup(s.DB, []byte("applications"), devAddr, &handlerEntry{})
	if err != nil {
		return nil, err
	}
	return entries.([]handlerEntry), nil
}

func (s handlerBoltStorage) Store(devAddr lorawan.DevAddr, entry handlerEntry) error {
	return store(s.DB, []byte("applications"), devAddr, &entry)
}

func (s handlerBoltStorage) Partition(packets ...core.Packet) ([]handlerPartition, error) {
	// Create a map in order to do the partition
	partitions := make(map[partitionId]handlerPartition)

	for _, packet := range packets {
		// First, determine devAddr, mandatory
		devAddr, err := packet.DevAddr()
		if err != nil {
			return nil, ErrInvalidPacket
		}

		entries, err := s.Lookup(devAddr)
		if err != nil {
			return nil, err
		}

		// Now get all tuples associated to that device address, and choose the right one
		for _, entry := range entries {
			// Compute MIC check to find the right keys
			ok, err := packet.Payload.ValidateMIC(entry.NwkSKey)
			if err != nil || !ok {
				continue // These aren't the droids you're looking for
			}

			// #Easy
			var id partitionId
			copy(id[:16], entry.AppEUI[:])
			copy(id[16:], entry.DevAddr[:])
			partitions[id] = handlerPartition{
				handlerEntry: entry,
				Id:           id,
				Packets:      append(partitions[id].Packets, packet),
			}
			break // We shouldn't look for other entries, we've found the right one
		}
	}

	// Transform the map to a slice
	res := make([]handlerPartition, 0, len(partitions))
	for _, p := range partitions {
		res = append(res, p)
	}

	if len(res) == 0 {
		return nil, ErrNotFound
	}

	return res, nil
}

func (entry handlerEntry) MarshalBinary() ([]byte, error) {
	w := NewEntryReadWriter(nil)
	w.Write(entry.AppEUI)
	w.Write(entry.AppSKey)
	w.Write(entry.DevAddr)
	w.Write(entry.NwkSKey)
	return w.Bytes()
}

func (entry *handlerEntry) UnmarshalBinary(data []byte) error {
	if entry == nil || len(data) < 4 {
		return ErrNotUnmarshable
	}
	r := NewEntryReadWriter(data)
	r.Read(func(data []byte) { copy(entry.AppEUI[:], data) })
	r.Read(func(data []byte) { copy(entry.AppSKey[:], data) })
	r.Read(func(data []byte) { copy(entry.DevAddr[:], data) })
	r.Read(func(data []byte) { copy(entry.NwkSKey[:], data) })
	return r.Err()
}

// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package components

import (
	"bytes"
	"encoding/binary"
	"io"
	"time"

	"github.com/TheThingsNetwork/ttn/core"
	"github.com/boltdb/bolt"
	"github.com/brocaar/lorawan"
)

const handlerBucket = "applications"

type boltHandlerStorage struct {
	db *bolt.DB
}

func NewHandlerBolt() (handlerStorage, error) {
	db, err := bolt.Open("handler_storage.db", 0600, &bolt.Options{Timeout: time.Second})
	if err != nil {
		return nil, err
	}
	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(handlerBucket))
		return err
	})
	if err != nil {
		return nil, err
	}

	return boltHandlerStorage{
		db: db,
	}, nil
}

func (s boltHandlerStorage) partition(packets []core.Packet) ([]handlerPartition, error) {
	// Create a map in order to do the partition
	partitions := make(map[partitionId]handlerPartition)

	for _, packet := range packets {
		// First, determine devAddr, mandatory
		devAddr, err := packet.DevAddr()
		if err != nil {
			return nil, ErrInvalidPacket
		}

		entries, err := s.lookup(devAddr)
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
				id:           id,
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

func (s boltHandlerStorage) lookup(devAddr lorawan.DevAddr) ([]handlerEntry, error) {
	var rawEntry []byte
	err := s.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(handlerBucket))
		if bucket == nil {
			return ErrStorageUnreachable
		}
		rawEntry = bucket.Get(devAddr[:])
		if rawEntry == nil {
			return ErrNotFound
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	buf := bytes.NewBuffer(rawEntry)
	var entries []handlerEntry
	for {
		lenEntry := new(uint16)
		if err := binary.Read(buf, binary.BigEndian, lenEntry); err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		entry := new(handlerEntry)
		if err := binary.Read(buf, binary.BigEndian, entry); err != nil {
			return nil, err
		}
		entries = append(entries, *entry)
	}

	return entries, nil
}

func (s boltHandlerStorage) store(devAddr lorawan.DevAddr, entry handlerEntry) error {
	marshalled, err := entry.MarshalBinary()
	if err != nil {
		return err
	}

	err = s.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(brokerBucket))
		if bucket == nil {
			return ErrStorageUnreachable
		}
		buf := bytes.NewBuffer(bucket.Get(devAddr[:]))
		binary.Write(buf, binary.BigEndian, len(marshalled))
		binary.Write(buf, binary.BigEndian, marshalled)
		return bucket.Put(devAddr[:], buf.Bytes())
	})

	return err
}

func (entry handlerEntry) MarshalBinary() ([]byte, error) {
	data := new(bytes.Buffer)
	var err error
	writeToData := func(content interface{}) {
		if err != nil {
			return
		}
		err = binary.Write(data, binary.BigEndian, content)
	}

	writeToData(uint16(len(entry.AppEUI)))
	writeToData(entry.AppEUI)
	writeToData(uint16(len(entry.NwkSKey)))
	writeToData(entry.NwkSKey)
	writeToData(uint16(len(entry.AppSKey)))
	writeToData(entry.AppSKey)
	writeToData(uint16(len(entry.DevAddr)))
	writeToData(entry.DevAddr)
	if err != nil {
		return nil, err
	}
	return data.Bytes(), nil
}

func (entry *handlerEntry) UnmarshalBinary(data []byte) error {
	if entry == nil || len(data) < 4 {
		return ErrNotUnmarshable
	}

	buf := bytes.NewBuffer(data)
	var err error
	readFromData := func(to func(data []byte)) {
		if err != nil {
			return
		}
		lenTo := new(uint16)
		if err = binary.Read(buf, binary.BigEndian, lenTo); err != nil {
			return
		}
		to(buf.Next(int(*lenTo)))
	}

	readFromData(func(data []byte) { copy(entry.AppEUI[:], data) })
	readFromData(func(data []byte) { copy(entry.NwkSKey[:], data) })
	readFromData(func(data []byte) { copy(entry.AppSKey[:], data) })
	readFromData(func(data []byte) { copy(entry.DevAddr[:], data) })

	return err
}

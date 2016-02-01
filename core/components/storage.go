// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package components

import (
	"bytes"
	"encoding/binary"
	"io"
	"reflect"

	"github.com/TheThingsNetwork/ttn/core"
	"github.com/boltdb/bolt"
	"github.com/brocaar/lorawan"
)

type handlerStorage interface {
	Lookup(devAddr lorawan.DevAddr) ([]handlerEntry, error)
	Store(devAddr lorawan.DevAddr, entry handlerEntry) error
	Partition(packet ...core.Packet) ([]handlerPartition, error)
}

type storageEntry interface {
	MarshalBinary() ([]byte, error)
	UnmarshalBinary(data []byte) error
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

func store(db *bolt.DB, bucketName []byte, devAddr lorawan.DevAddr, entry storageEntry) error {
	marshalled, err := entry.MarshalBinary()
	if err != nil {
		return err
	}

	err = db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(bucketName)
		if bucket == nil {
			return ErrStorageUnreachable
		}
		buf := bytes.NewBuffer(bucket.Get(devAddr[:]))
		binary.Write(buf, binary.BigEndian, uint16(len(marshalled)))
		binary.Write(buf, binary.BigEndian, marshalled)
		return bucket.Put(devAddr[:], buf.Bytes())
	})

	return err
}

func lookup(db *bolt.DB, bucketName []byte, devAddr lorawan.DevAddr, shape storageEntry) (interface{}, error) {
	var rawEntry []byte
	err := db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(bucketName)
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
	entryType := reflect.TypeOf(shape).Elem()
	entries := reflect.MakeSlice(reflect.SliceOf(entryType), 0, 0)
	for {
		lenEntry := new(uint16)
		if err := binary.Read(buf, binary.BigEndian, lenEntry); err != nil {
			if err == io.EOF {
				break
			}
			panic(err)
		}
		entry := reflect.New(entryType).Interface()
		entry.(storageEntry).UnmarshalBinary(buf.Next(int(*lenEntry)))
		entries = reflect.Append(entries, reflect.ValueOf(entry).Elem())
	}
	return entries.Interface(), nil
}

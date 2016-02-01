// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package components

import (
	"io"
	"reflect"

	"github.com/boltdb/bolt"
	"github.com/brocaar/lorawan"
)

type storageEntry interface {
	MarshalBinary() ([]byte, error)
	UnmarshalBinary(data []byte) error
}

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
		w := NewEntryReadWriter(bucket.Get(devAddr[:]))
		w.Write(uint16(len(marshalled)))
		w.Write(marshalled)
		data, err := w.Bytes()
		if err != nil {
			return err
		}
		return bucket.Put(devAddr[:], data)
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

	r := NewEntryReadWriter(rawEntry)
	entryType := reflect.TypeOf(shape).Elem()
	entries := reflect.MakeSlice(reflect.SliceOf(entryType), 0, 0)
	for {
		r.Read(func(data []byte) {
			entry := reflect.New(entryType).Interface()
			entry.(storageEntry).UnmarshalBinary(data)
			entries = reflect.Append(entries, reflect.ValueOf(entry).Elem())
		})
		if err = r.Err(); err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
	}
	return entries.Interface(), nil
}

// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package components

import (
	"fmt"
	"io"
	"reflect"

	. "github.com/TheThingsNetwork/ttn/core/errors"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	"github.com/boltdb/bolt"
	"github.com/brocaar/lorawan"
)

// storageEntry offers a friendly interface on which the storage will operate.
// Basically, a storageEntry is nothing more than a binary marshaller/unmarshaller.
type storageEntry interface {
	MarshalBinary() ([]byte, error)    // implements binary.Marshaller interface
	UnmarshalBinary(data []byte) error // implements binary.Unmarshaller interface
}

// initDB initializes the given bolt database by creating (if not already exists) an empty bucket
func initDB(db *bolt.DB, bucketName string) error {
	return db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(bucketName))
		return err
	})
}

// store put a new entry in the given bolt database. It adds the entry to an existing set or create
// a new set containing one element.
func store(db *bolt.DB, bucketName string, devAddr lorawan.DevAddr, entry storageEntry) error {
	marshalled, err := entry.MarshalBinary()
	if err != nil {
		return errors.New(ErrInvalidStructure, err)
	}

	err = db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(bucketName))
		if bucket == nil {
			return errors.New(ErrFailedOperation, "storage unreachable")
		}
		w := newEntryReadWriter(bucket.Get(devAddr[:]))
		w.Write(marshalled)
		data, err := w.Bytes()
		if err != nil {
			return errors.New(ErrInvalidStructure, err)
		}
		if err := bucket.Put(devAddr[:], data); err != nil {
			return errors.New(ErrFailedOperation, err)
		}
		return nil
	})

	return err
}

// lookup retrieve a set of entry from a given bolt database.
//
// The shape is used as a template for retrieving and creating the data. All entries extracted from
// the database will be interpreted as instance of shape and the return result will be a slice of
// the same type of shape.
func lookup(db *bolt.DB, bucketName string, devAddr lorawan.DevAddr, shape storageEntry) (interface{}, error) {
	// First, lookup the raw entries
	var rawEntry []byte
	err := db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(bucketName))
		if bucket == nil {
			return errors.New(ErrFailedOperation, "storage unreachable")
		}
		rawEntry = bucket.Get(devAddr[:])
		if rawEntry == nil {
			return errors.New(ErrNotFound, fmt.Sprintf("%+v", devAddr))
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	// Then, interpret them as instance of 'shape'
	r := newEntryReadWriter(rawEntry)
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
			return nil, errors.New(ErrFailedOperation, err)
		}
	}
	return entries.Interface(), nil
}

// flush empties each entry of a bucket associated to a given device
func flush(db *bolt.DB, bucketName string, devAddr lorawan.DevAddr) error {
	return db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(bucketName))
		if bucket == nil {
			return errors.New(ErrFailedOperation, "storage unreachable")
		}
		if err := bucket.Delete(devAddr[:]); err != nil {
			return errors.New(ErrFailedOperation, err)
		}
		return nil
	})
}

// resetDB resets a given bucket from a given bolt database
func resetDB(db *bolt.DB, bucketName string) error {
	return db.Update(func(tx *bolt.Tx) error {
		if err := tx.DeleteBucket([]byte(bucketName)); err != nil {
			return errors.New(ErrFailedOperation, err)
		}
		if _, err := tx.CreateBucketIfNotExists([]byte(bucketName)); err != nil {
			return errors.New(ErrFailedOperation, err)
		}
		return nil
	})
}

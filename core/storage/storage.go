// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package storage

import (
	"encoding"
	"fmt"
	"reflect"
	"time"

	"github.com/TheThingsNetwork/ttn/utils/errors"
	"github.com/TheThingsNetwork/ttn/utils/readwriter"
	"github.com/boltdb/bolt"
)

// The storage Interface provides an abstraction on top of the bolt database to store and access
// data in a local in-memory database.
// All methods uses a slice of buckets as bucket selector, each subsequent bucket being nested in
// the previous one. Each time also, the key can be omitted in which case it will fall back to a
// default key.
type Interface interface {
	// Read retrieves a slice of entries from the storage. The output is a slice of the same type as
	// of `shape`, possibly of length 0.
	// The provided type has to implement a binary.Unmarshaler interface
	Read(key []byte, shape encoding.BinaryUnmarshaler, buckets ...[]byte) (interface{}, error)
	// ReadAll goes through each key of a bucket and return a slice of all values. This implies
	// therefore that all values in each key are of the same type (the shape provided)
	ReadAll(shape encoding.BinaryUnmarshaler, buckets ...[]byte) (interface{}, error)
	// Update replaces a set of entries by the new given set
	Update(key []byte, entries []encoding.BinaryMarshaler, buckets ...[]byte) error
	// Append adds at the end of an existing set the new given set
	Append(key []byte, entries []encoding.BinaryMarshaler, buckets ...[]byte) error
	// Delete complete removes all targeted entries
	Delete(key []byte, buckets ...[]byte) error
	// Reset empties a bucket
	Reset(buckets ...[]byte) error
	// Close terminates the communication with the storage
	Close() error
}

var defaultKey = []byte("entry")

type store struct {
	db *bolt.DB
}

// New creates a new storage instance ready-to-be-used
func New(name string) (Interface, error) {
	db, err := bolt.Open(name, 0600, &bolt.Options{Timeout: time.Second})
	if err != nil {
		return nil, errors.New(errors.Operational, err)
	}
	return store{db}, nil
}

// Make sure we return a failure
func ensureErr(err error) error {
	if err == nil {
		return nil
	}
	_, ok := err.(errors.Failure)
	if !ok {
		err = errors.New(errors.Operational, err)
	}
	return err
}

// getBucket retrieves a bucket based on a slice of ordered identifiers. Each following identifier
// targets a nested bucket in the previous one. If no bucket is found along the path, they are
// created if the write rights are granted by Tx, otherwise, the existing one is used.
func getBucket(tx *bolt.Tx, buckets [][]byte) (*bolt.Bucket, error) {
	if len(buckets) < 1 {
		return nil, errors.New(errors.Structural, "At least one bucket name is required")
	}
	var cursor interface {
		CreateBucketIfNotExists(b []byte) (*bolt.Bucket, error)
		Bucket(b []byte) *bolt.Bucket
	}

	var err error
	cursor = tx
	for _, name := range buckets {
		next := cursor.Bucket(name)
		if next == nil {
			if next, err = cursor.CreateBucketIfNotExists(name); err != nil {
				return nil, errors.New(errors.Operational, err)
			}
		}
		cursor = next
	}
	return cursor.(*bolt.Bucket), nil
}

// ReadAll implements the storage.Interface interface
func (itf store) ReadAll(shape encoding.BinaryUnmarshaler, buckets ...[]byte) (interface{}, error) {
	entryType := reflect.TypeOf(shape)
	if entryType.Kind() != reflect.Ptr {
		return nil, errors.New(errors.Implementation, "Non-pointer shape not supported")
	}

	entries := reflect.MakeSlice(reflect.SliceOf(entryType.Elem()), 0, 0)
	err := itf.db.View(func(tx *bolt.Tx) error {
		bucket, err := getBucket(tx, buckets)
		if err != nil {
			if err.(errors.Failure).Fault == bolt.ErrTxNotWritable {
				return errors.New(errors.NotFound, fmt.Sprintf("Not found %+v", buckets))
			}
			return err
		}
		return bucket.ForEach(func(_ []byte, v []byte) error {
			r := readwriter.New(v)
			r.Read(func(data []byte) {
				entry := reflect.New(entryType.Elem()).Interface()
				entry.(encoding.BinaryUnmarshaler).UnmarshalBinary(data)
				entries = reflect.Append(entries, reflect.ValueOf(entry).Elem())
			})
			return r.Err()
		})
	})
	if err != nil {
		return nil, ensureErr(err)
	}
	if entries.Len() == 0 {
		return nil, errors.New(errors.NotFound, fmt.Sprintf("Not found %+v", buckets))
	}
	return entries.Interface(), nil
}

// Read implements the storage.Interface interface
func (itf store) Read(key []byte, shape encoding.BinaryUnmarshaler, buckets ...[]byte) (interface{}, error) {
	if key == nil {
		key = defaultKey
	}

	entryType := reflect.TypeOf(shape)
	if entryType.Kind() != reflect.Ptr {
		return nil, errors.New(errors.Implementation, "Non-pointer shape not supported")
	}

	// First, lookup the raw entries
	var rawEntry []byte
	err := itf.db.View(func(tx *bolt.Tx) error {
		bucket, err := getBucket(tx, buckets)
		if err != nil {
			if err.(errors.Failure).Fault == bolt.ErrTxNotWritable {
				return errors.New(errors.NotFound, fmt.Sprintf("Not found %+v", key))
			}
			return err
		}
		rawEntry = bucket.Get(key)
		if rawEntry == nil {
			return errors.New(errors.NotFound, fmt.Sprintf("Not found %+v", key))
		}
		return nil
	})

	if err != nil {
		return nil, ensureErr(err)
	}

	// Then, interpret them as instance of 'shape'
	r := readwriter.New(rawEntry)
	entries := reflect.MakeSlice(reflect.SliceOf(entryType.Elem()), 0, 0)
	var nb uint
	for {
		r.Read(func(data []byte) {
			entry := reflect.New(entryType.Elem()).Interface()
			entry.(encoding.BinaryUnmarshaler).UnmarshalBinary(data)
			entries = reflect.Append(entries, reflect.ValueOf(entry).Elem())
			nb++
		})
		if err = r.Err(); err != nil {
			failure, ok := err.(errors.Failure)
			if ok && failure.Nature == errors.Behavioural {
				break
			}
			return nil, errors.New(errors.Operational, err)
		}
	}
	if nb == 0 {
		return nil, errors.New(errors.NotFound, fmt.Sprintf("Not found %+v", key))
	}
	return entries.Interface(), nil
}

// Append implements the storage.Interface interface
func (itf store) Append(key []byte, entries []encoding.BinaryMarshaler, buckets ...[]byte) error {
	if key == nil {
		key = defaultKey
	}
	var marshalled [][]byte

	for _, entry := range entries {
		m, err := entry.MarshalBinary()
		if err != nil {
			return errors.New(errors.Structural, err)
		}
		marshalled = append(marshalled, m)
	}

	err := itf.db.Update(func(tx *bolt.Tx) error {
		bucket, err := getBucket(tx, buckets)
		if err != nil {
			return err
		}
		w := readwriter.New(bucket.Get(key))
		for _, m := range marshalled {
			w.Write(m)
		}
		data, err := w.Bytes()
		if err != nil {
			return errors.New(errors.Structural, err)
		}
		if err := bucket.Put(key, data); err != nil {
			return errors.New(errors.Operational, err)
		}
		return nil
	})

	return ensureErr(err)
}

// Update implements the storage.Interface interface
func (itf store) Update(key []byte, entries []encoding.BinaryMarshaler, buckets ...[]byte) error {
	if key == nil {
		key = defaultKey
	}
	var marshalled [][]byte

	for _, entry := range entries {
		m, err := entry.MarshalBinary()
		if err != nil {
			return errors.New(errors.Structural, err)
		}
		marshalled = append(marshalled, m)
	}

	return ensureErr(itf.db.Update(func(tx *bolt.Tx) error {
		bucket, err := getBucket(tx, buckets)
		if err != nil {
			return err
		}
		if err := bucket.Delete(key); err != nil {
			return errors.New(errors.Operational, err)
		}
		w := readwriter.New(bucket.Get(key))
		for _, m := range marshalled {
			w.Write(m)
		}
		data, err := w.Bytes()
		if err != nil {
			return errors.New(errors.Structural, err)
		}
		if err := bucket.Put(key, data); err != nil {
			return errors.New(errors.Operational, err)
		}
		return nil
	}))
}

// Delete implements the storage.Interface interface
func (itf store) Delete(key []byte, buckets ...[]byte) error {
	return ensureErr(itf.db.Update(func(tx *bolt.Tx) error {
		bucket, err := getBucket(tx, buckets)
		if err != nil {
			return err
		}
		if err := bucket.Delete(key); err != nil {
			return errors.New(errors.Operational, err)
		}
		return nil
	}))
}

// Reset implements the storage.Interface interface
func (itf store) Reset(buckets ...[]byte) (err error) {
	if len(buckets) == 0 {
		return errors.New(errors.Structural, "Expected at least one bucket")
	}

	return ensureErr(itf.db.Update(func(tx *bolt.Tx) error {
		var cursor interface {
			DeleteBucket(name []byte) error
			CreateBucketIfNotExists(name []byte) (*bolt.Bucket, error)
		}

		init, last := buckets[:len(buckets)-1], buckets[len(buckets)-1]

		if len(init) == 0 {
			cursor = tx
		} else {
			cursor, err = getBucket(tx, init)
			if err != nil {
				return err
			}
		}

		if err := cursor.DeleteBucket(last); err != nil {
			return errors.New(errors.Operational, err)
		}
		if _, err := cursor.CreateBucketIfNotExists(last); err != nil {
			return errors.New(errors.Operational, err)
		}
		return nil
	}))
}

// Close implements the storage.Interface interface
func (itf store) Close() error {
	if err := itf.db.Close(); err != nil {
		return errors.New(errors.Operational, err)
	}
	return nil
}

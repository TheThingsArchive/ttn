// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package router

import (
	"encoding"
	"encoding/binary"
	"fmt"
	"sync"
	"time"

	dbutil "github.com/TheThingsNetwork/ttn/core/storage"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	"github.com/TheThingsNetwork/ttn/utils/readwriter"
)

var dbBrokers = []byte("brokers")

// BrkStorage gives a facade to manipulate the router's brokers
type BrkStorage interface {
	read(devAddr []byte) ([]brkEntry, error)
	create(entry brkEntry) error
	//remove(entry brkEntry) error
	done() error
}

type brkEntry struct {
	DevAddr     []byte
	BrokerIndex uint16
	until       time.Time
}

type brkStorage struct {
	sync.Mutex
	ExpiryDelay time.Duration
	db          dbutil.Interface
}

// NewBrkStorage creates a new internal storage for the router
func NewBrkStorage(name string, delay time.Duration) (BrkStorage, error) {
	itf, err := dbutil.New(name)
	if err != nil {
		return nil, errors.New(errors.Operational, err)
	}
	return &brkStorage{db: itf, ExpiryDelay: delay}, nil
}

// read implements the router.BrkStorage interface
func (s *brkStorage) read(devAddr []byte) ([]brkEntry, error) {
	return s._read(devAddr, true)
}

// _read implements the router.BrkStorage interface logic but gives control over mutex to the caller
func (s *brkStorage) _read(devAddr []byte, shouldLock bool) ([]brkEntry, error) {
	if shouldLock {
		s.Lock()
		defer s.Unlock()
	}
	itf, err := s.db.Read(devAddr, &brkEntry{}, dbBrokers)
	if err != nil {
		return nil, err
	}
	entries := itf.([]brkEntry)

	if s.ExpiryDelay != 0 {
		// Get rid of expired entries
		var newEntries []encoding.BinaryMarshaler
		var filtered []brkEntry
		for _, e := range entries {
			if e.until.After(time.Now()) {
				newEntry := new(brkEntry)
				*newEntry = e
				newEntries = append(newEntries, newEntry)
				filtered = append(filtered, e)
			}
		}
		// Replace filtered entries
		if err := s.db.Update(devAddr, newEntries, dbBrokers); err != nil {
			return nil, errors.New(errors.Operational, err)
		}
		entries = filtered
	}

	if len(entries) == 0 {
		return nil, errors.New(errors.NotFound, fmt.Sprintf("No entry for: %v", devAddr))
	}
	return entries, nil
}

// create implements the router.BrkStorage interface
func (s *brkStorage) create(entry brkEntry) error {
	s.Lock()
	defer s.Unlock()

	entries, err := s._read(entry.DevAddr, false)
	if err != nil && err.(errors.Failure).Nature != errors.NotFound {
		return err
	}

	var updates []encoding.BinaryMarshaler
	until, found := time.Now().Add(s.ExpiryDelay), false
	for i, e := range entries {
		if entry.BrokerIndex == e.BrokerIndex {
			// Entry already there, just update the TTL
			entries[i].until = until
			found = true
		}
		updates = append(updates, e)
	}

	if found { // The entry was already existing
		return s.db.Update(entry.DevAddr, updates, dbBrokers)
	}

	// Otherwise, we just happend it
	entry.until = until
	return s.db.Append(entry.DevAddr, []encoding.BinaryMarshaler{entry}, dbBrokers)
}

// done implements the router.BrkStorage interface
func (s *brkStorage) done() error {
	return s.db.Close()
}

// MarshalBinary implements the encoding.BinaryMarshaler
func (e brkEntry) MarshalBinary() ([]byte, error) {
	data, err := e.until.MarshalBinary()
	if err != nil {
		return nil, errors.New(errors.Structural, err)
	}
	rw := readwriter.New(nil)
	rw.Write(e.BrokerIndex)
	rw.Write(e.DevAddr)
	rw.Write(data)
	return rw.Bytes()
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler
func (e *brkEntry) UnmarshalBinary(data []byte) error {
	rw := readwriter.New(data)
	rw.Read(func(data []byte) { e.BrokerIndex = binary.BigEndian.Uint16(data) })
	rw.Read(func(data []byte) {
		e.DevAddr = make([]byte, len(data))
		copy(e.DevAddr, data)
	})
	rw.TryRead(func(data []byte) error { return e.until.UnmarshalBinary(data) })
	return rw.Err()
}

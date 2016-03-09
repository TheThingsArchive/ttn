// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package router

import (
	"fmt"
	"sync"
	"time"

	. "github.com/TheThingsNetwork/ttn/core"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	"github.com/TheThingsNetwork/ttn/utils/readwriter"
	dbutil "github.com/TheThingsNetwork/ttn/utils/storage"
	"github.com/brocaar/lorawan"
)

// Storage gives a facade to manipulate the router database
type Storage interface {
	Lookup(devEUI lorawan.EUI64) ([]entry, error)
	Store(reg RRegistration) error
	Close() error
}

type entry struct {
	Recipient []byte
	until     time.Time
}

type storage struct {
	sync.Mutex
	db          dbutil.Interface
	Name        string
	ExpiryDelay time.Duration
}

// NewStorage creates a new internal storage for the router
func NewStorage(name string, delay time.Duration) (Storage, error) {
	itf, err := dbutil.New(name)
	if err != nil {
		return nil, errors.New(errors.Operational, err)
	}

	return &storage{db: itf, ExpiryDelay: delay, Name: "broker"}, nil
}

// Lookup implements the router.Storage interface
func (s *storage) Lookup(devEUI lorawan.EUI64) ([]entry, error) {
	s.Lock()
	defer s.Unlock()
	itf, err := s.db.Lookup(s.Name, devEUI[:], &entry{})
	if err != nil {
		return nil, err
	}
	entries := itf.([]entry)

	if s.ExpiryDelay != 0 {
		var newEntries []dbutil.Entry
		var filtered []entry
		for _, e := range entries {
			if e.until.After(time.Now()) {
				newEntry := new(entry)
				*newEntry = e
				newEntries = append(newEntries, newEntry)
				filtered = append(filtered, e)
			}
		}
		if err := s.db.Replace(s.Name, devEUI[:], newEntries); err != nil {
			return nil, errors.New(errors.Operational, err)
		}
		entries = filtered
	}

	if len(entries) == 0 {
		return nil, errors.New(errors.NotFound, fmt.Sprintf("No entry for: %v", devEUI[:]))
	}
	return entries, nil
}

// Store implements the router.Storage interface
func (s *storage) Store(reg RRegistration) error {
	devEUI := reg.DevEUI()
	recipient, err := reg.Recipient().MarshalBinary()
	if err != nil {
		return errors.New(errors.Structural, err)
	}

	s.Lock()
	defer s.Unlock()
	return s.db.Store(s.Name, devEUI[:], []dbutil.Entry{&entry{
		Recipient: recipient,
		until:     time.Now().Add(s.ExpiryDelay),
	}})
}

// Close implements the router.Storage interface
func (s *storage) Close() error {
	return s.db.Close()
}

// MarshalBinary implements the encoding.BinaryMarshaler interface
func (e entry) MarshalBinary() ([]byte, error) {
	data, err := e.until.MarshalBinary()
	if err != nil {
		return nil, errors.New(errors.Structural, err)
	}

	rw := readwriter.New(nil)
	rw.Write(e.Recipient)
	rw.Write(data)
	return rw.Bytes()
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface
func (e *entry) UnmarshalBinary(data []byte) error {
	rw := readwriter.New(data)
	rw.Read(func(data []byte) { e.Recipient = data })
	rw.TryRead(func(data []byte) error {
		return e.until.UnmarshalBinary(data)
	})
	return rw.Err()
}

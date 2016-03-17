// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package handler

import (
	"fmt"

	dbutil "github.com/TheThingsNetwork/ttn/core/storage"
	"github.com/TheThingsNetwork/ttn/utils/errors"
)

// PktStorage gives a facade to manipulate the handler packets database
type PktStorage interface {
	Push(appEUI []byte, devEUI []byte, payload pktEntry) error
	Pull(appEUI []byte, devEUI []byte) (pktEntry, error)
	Close() error
}

const dbPackets = "packets"

type pktStorage struct {
	db dbutil.Interface
}

type pktEntry []byte

// NewPktStorage creates a new PktStorage
func NewPktStorage(name string) (PktStorage, error) {
	itf, err := dbutil.New(name)
	if err != nil {
		return nil, errors.New(errors.Operational, err)
	}
	return pktStorage{db: itf}, nil
}

// Push implements the PktStorage interface
func (s pktStorage) Push(appEUI, devEUI []byte, payload pktEntry) error {
	return s.db.Store(dbPackets, append(appEUI, devEUI...), []dbutil.Entry{&payload})
}

// Pull implements the PktStorage interface
func (s pktStorage) Pull(appEUI, devEUI []byte) (pktEntry, error) {
	key := append(appEUI, devEUI...)
	entries, err := s.db.Lookup(dbPackets, key, &pktEntry{})
	if err != nil {
		return nil, err // Operational || NotFound
	}

	payloads, ok := entries.([]pktEntry)
	if !ok {
		return nil, errors.New(errors.Operational, "Unable to retrieve data from db")
	}

	// NOTE: one day, those entries will be more complicated, with a ttl.
	// Here's the place where we should check for that. Cheers.
	if len(payloads) == 0 {
		return nil, errors.New(errors.NotFound, fmt.Sprintf("Entry not found for %v", key))
	}

	payload := payloads[0]

	var newEntries []dbutil.Entry
	for _, p := range payloads[1:] {
		newEntries = append(newEntries, &p)
	}

	if err := s.db.Replace(dbPackets, key, newEntries); err != nil {
		if err := s.db.Replace(dbPackets, key, newEntries); err != nil {
			// TODO This is critical... we've just lost a packet
			return nil, errors.New(errors.Operational, "Unable to restore data in db")
		}
	}

	return payload, nil
}

// Close implements the PktStorage interface
func (s pktStorage) Close() error {
	return s.db.Close()
}

// MarshalBinary implements the encoding.BinaryMarshaler interface
func (e pktEntry) MarshalBinary() ([]byte, error) {
	return e, nil
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface
func (e *pktEntry) UnmarshalBinary(data []byte) error {
	*e = data
	return nil
}

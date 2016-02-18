// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package components

import (
	"time"

	. "github.com/TheThingsNetwork/ttn/core/errors"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	"github.com/boltdb/bolt"
	"github.com/brocaar/lorawan"
)

// BrokerStorage manages the internal persistent state of a broker
type BrokerStorage interface {
	// Close properly ends the connection to the internal database
	Close() error

	// Lookup retrieves all entries associated to a given device
	Lookup(devAddr lorawan.DevAddr) ([]brokerEntry, error)

	// Reset removes all entries stored in the storage
	Reset() error

	// Store creates a new entry and add it to the other entries (if any)
	Store(devAddr lorawan.DevAddr, entry brokerEntry) error
}

type brokerBoltStorage struct {
	*bolt.DB
}

// brokerEntry stores all information that links a handler to a device
type brokerEntry struct {
	Id      string            // The handler / application ID
	NwkSKey lorawan.AES128Key // The network session key associated to the device
	Url     string            // The webook url of the associated handler // NOTE This implies an http protocol. Should review.
}

// NewBrokerStorage a new bolt broker in-memory storage
func NewBrokerStorage() (BrokerStorage, error) {
	db, err := bolt.Open("broker_storage.db", 0600, &bolt.Options{Timeout: time.Second})
	if err != nil {
		return nil, errors.New(ErrFailedOperation, err)
	}

	if err := initDB(db, "devices"); err != nil {
		return nil, err
	}

	return &brokerBoltStorage{DB: db}, nil
}

// Lookup implements the brokerStorage interface
func (s brokerBoltStorage) Lookup(devAddr lorawan.DevAddr) ([]brokerEntry, error) {
	entries, err := lookup(s.DB, "devices", devAddr, &brokerEntry{})
	if err != nil {
		return nil, err
	}
	return entries.([]brokerEntry), nil
}

// Store implements the brokerStorage interface
func (s brokerBoltStorage) Store(devAddr lorawan.DevAddr, entry brokerEntry) error {
	return store(s.DB, "devices", devAddr, &entry)
}

// Close implements the brokerStorage interface
func (s brokerBoltStorage) Close() error {
	return s.DB.Close()
}

// Reset implements the brokerStorage interface
func (s brokerBoltStorage) Reset() error {
	return resetDB(s.DB, "devices")
}

// MarshalBinary implements the entryStorage interface
func (entry brokerEntry) MarshalBinary() ([]byte, error) {
	w := newEntryReadWriter(nil)
	w.Write(entry.Id)
	w.Write(entry.NwkSKey)
	w.Write(entry.Url)
	return w.Bytes()
}

// UnmarshalBinary implements the entryStorage interface
func (entry *brokerEntry) UnmarshalBinary(data []byte) error {
	if entry == nil || len(data) < 3 {
		return errors.New(ErrInvalidStructure, "invalid broker entry")
	}
	r := newEntryReadWriter(data)
	r.Read(func(data []byte) { entry.Id = string(data) })
	r.Read(func(data []byte) { copy(entry.NwkSKey[:], data) })
	r.Read(func(data []byte) { entry.Url = string(data) })
	return r.Err()
}

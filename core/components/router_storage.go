// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package components

import (
	"time"

	"github.com/TheThingsNetwork/ttn/core"
	"github.com/boltdb/bolt"
	"github.com/brocaar/lorawan"
)

// RouterStorage manages the internal persistent state of a router
type RouterStorage interface {
	// Close properly ends the connection to the internal database
	Close() error

	// Lookup retrieves all entries associated to a given device
	Lookup(devAddr lorawan.DevAddr) (routerEntry, error)

	// Reset removes all entries stored in the storage
	Reset() error

	// Store creates a new entry and add it to the other entries (if any)
	Store(devAddr lorawan.DevAddr, entry routerEntry) error
}

type routerBoltStorage struct {
	*bolt.DB
	expiryDelay time.Duration // Entry lifetime delay
}

// routerEntry stores all information that link a device to a broker
type routerEntry struct {
	Recipient core.Recipient // Recipient associated to a device.
	until     time.Time      // The moment until when the entry is still valid
}

// NewRouterStorage creates a new router bolt in-memory storage
func NewRouterStorage() (RouterStorage, error) {
	db, err := bolt.Open("router_storage.db", 0600, &bolt.Options{Timeout: time.Second})
	if err != nil {
		return nil, err
	}

	if err := initDB(db, "brokers"); err != nil {
		return nil, err
	}

	return &routerBoltStorage{DB: db}, nil
}

// Lookup implements the RouterStorage interface
func (s routerBoltStorage) Lookup(devAddr lorawan.DevAddr) (routerEntry, error) {
	entries, err := lookup(s.DB, "brokers", devAddr, &routerEntry{})
	if err != nil {
		return routerEntry{}, err
	}
	routerEntries := entries.([]routerEntry)

	if len(routerEntries) != 1 {
		if err := flush(s.DB, "brokers", devAddr); err != nil {
			return routerEntry{}, err
		}
		return routerEntry{}, ErrNotFound
	}

	if s.expiryDelay != 0 && routerEntries[0].until.Before(time.Now()) {
		if err := flush(s.DB, "brokers", devAddr); err != nil {
			return routerEntry{}, err
		}
		return routerEntry{}, ErrEntryExpired
	}

	return routerEntries[0], nil
}

// Store implements the RouterStorage interface
func (s routerBoltStorage) Store(devAddr lorawan.DevAddr, entry routerEntry) error {
	entry.until = time.Now().Add(s.expiryDelay)
	return store(s.DB, "brokers", devAddr, &entry)
}

// Close implements the RouterStorage interface
func (s routerBoltStorage) Close() error {
	return s.DB.Close()
}

// Reset implements the RouterStorage interface
func (s routerBoltStorage) Reset() error {
	return resetDB(s.DB, "brokers")
}

// MarshalBinary implements the entryStorage interface
func (entry routerEntry) MarshalBinary() ([]byte, error) {
	rawTime, err := entry.until.MarshalBinary()
	if err != nil {
		return nil, err
	}
	rawId := []byte(entry.Recipient.Id.(string))
	rawAddress := []byte(entry.Recipient.Address.(string))

	w := newEntryReadWriter(nil)
	w.Write(rawId)
	w.Write(rawAddress)
	w.Write(rawTime)
	return w.Bytes()
}

// UnmarshalBinary implements the entryStorage interface
func (entry *routerEntry) UnmarshalBinary(data []byte) error {
	if entry == nil || len(data) < 1 {
		return ErrNotUnmarshable
	}
	r := newEntryReadWriter(data[0:])

	var id, address string
	r.Read(func(data []byte) { id = string(data) })
	r.Read(func(data []byte) { address = string(address) })
	entry.Recipient = core.Recipient{
		Id:      id,
		Address: address,
	}
	var err error
	r.Read(func(data []byte) {
		entry.until = time.Time{}
		err = entry.until.UnmarshalBinary(data)
	})
	if err != nil {
		return err
	}
	return r.Err()
}

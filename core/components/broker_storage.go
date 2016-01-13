// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package components

import (
	"github.com/thethingsnetwork/ttn/lorawan"
	"sync"
)

type brokerStorage interface {
	lookup(devAddr lorawan.DevAddr) ([]brokerEntry, error)
	store(devAddr lorawan.DevAddr, entry brokerEntry) error
}

type brokerEntry struct {
	Id     string
	Url    string
	NwsKey lorawan.AES128Key
}

type brokerDB struct {
	sync.RWMutex
	entries map[lorawan.DevAddr][]brokerEntry
}

// NewLocalDB constructs a new local brokerStorage
func NewBrokerStorage() (brokerStorage, error) {
	return &brokerDB{entries: make(map[lorawan.DevAddr][]brokerEntry)}, nil
}

// lookup implements the brokerStorage interface
func (db *brokerDB) lookup(devAddr lorawan.DevAddr) ([]brokerEntry, error) {
	db.RLock()
	entries, ok := db.entries[devAddr]
	db.RUnlock()
	if !ok {
		return nil, ErrDeviceNotFound
	}
	return entries, nil
}

// store implements the brokerStorage interface
func (db *brokerDB) store(devAddr lorawan.DevAddr, entry brokerEntry) error {
	db.Lock()
	defer db.Unlock()
	entries, ok := db.entries[devAddr]
	if !ok {
		db.entries[devAddr] = []brokerEntry{entry}
		return nil
	}
	db.entries[devAddr] = append(entries, entry)
	return nil
}

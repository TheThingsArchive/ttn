// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package components

import (
	"sync"
	"time"

	"github.com/TheThingsNetwork/ttn/core"
	"github.com/brocaar/lorawan"
)

type routerStorage interface {
	lookup(devAddr lorawan.DevAddr) ([]core.Recipient, error)
	store(devAddr lorawan.DevAddr, recipients ...core.Recipient) error
}

type routerDB struct {
	sync.RWMutex
	expiryDelay time.Duration
	entries     map[lorawan.DevAddr]routerEntry
}

type routerEntry struct {
	recipients []core.Recipient
	until      time.Time
}

// NewLocalDB constructs a new local address keeper
func NewRouterStorage(expiryDelay time.Duration) (routerStorage, error) {
	return &routerDB{
		expiryDelay: expiryDelay,
		entries:     make(map[lorawan.DevAddr]routerEntry),
	}, nil
}

// lookup implements the addressKeeper interface
func (db *routerDB) lookup(devAddr lorawan.DevAddr) ([]core.Recipient, error) {
	db.RLock()
	entry, ok := db.entries[devAddr]
	db.RUnlock()
	if !ok {
		return nil, ErrDeviceNotFound
	}

	if db.expiryDelay != 0 && entry.until.Before(time.Now()) {
		db.Lock()
		delete(db.entries, devAddr)
		db.Unlock()
		return nil, ErrEntryExpired
	}

	return entry.recipients, nil
}

// store implements the addressKeeper interface
func (db *routerDB) store(devAddr lorawan.DevAddr, recipients ...core.Recipient) error {
	db.Lock()
	_, ok := db.entries[devAddr]
	if ok {
		db.Unlock()
		return ErrAlreadyExists
	}

	db.entries[devAddr] = routerEntry{
		recipients: recipients,
		until:      time.Now().Add(db.expiryDelay),
	}

	db.Unlock()
	return nil
}

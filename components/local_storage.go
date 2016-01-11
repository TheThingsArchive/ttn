// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package components

import (
	"fmt"
	"github.com/thethingsnetwork/core"
	"github.com/thethingsnetwork/core/lorawan"
	"sync"
	"time"
)

type addressKeeper interface {
	lookup(devAddr lorawan.DevAddr) ([]core.Recipient, error)
	store(devAddr lorawan.DevAddr, recipients ...core.Recipient) error
}

type reddisAddressKeeper struct{} // NOTE In a second time

type localDB struct {
	sync.RWMutex
	expiryDelay time.Duration
	addresses   map[lorawan.DevAddr]localEntry
}

type localEntry struct {
	recipients []core.Recipient
	until      time.Time
}

var ErrDeviceNotFound = fmt.Errorf("Device not found")
var ErrEntryExpired = fmt.Errorf("An entry exists but has expired")

// NewLocalDB constructs a new local address keeper
func NewLocalDB(expiryDelay time.Duration) (*localDB, error) {
	if expiryDelay == 0 {
		return nil, fmt.Errorf("Invalid expiration delay")
	}

	return &localDB{
		expiryDelay: expiryDelay,
		addresses:   make(map[lorawan.DevAddr]localEntry),
	}, nil
}

// lookup implements the addressKeeper interface
func (a *localDB) lookup(devAddr lorawan.DevAddr) ([]core.Recipient, error) {
	a.RLock()
	entry, ok := a.addresses[devAddr]
	a.RUnlock()
	if !ok {
		return nil, ErrDeviceNotFound
	}

	if entry.until.Before(time.Now()) {
		a.Lock()
		delete(a.addresses, devAddr)
		a.Unlock()
		return nil, ErrEntryExpired
	}

	return entry.recipients, nil
}

// store implements the addressKeeper interface
func (a *localDB) store(devAddr lorawan.DevAddr, recipients ...core.Recipient) error {
	a.Lock()
	_, ok := a.addresses[devAddr]
	if ok {
		a.Unlock()
		return fmt.Errorf("An entry already exists for that device")
	}

	a.addresses[devAddr] = localEntry{
		recipients: recipients,
		until:      time.Now().Add(a.expiryDelay),
	}

	a.Unlock()
	return nil
}

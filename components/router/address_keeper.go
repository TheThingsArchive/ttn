// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package router

import (
	"fmt"
	"github.com/thethingsnetwork/core/lorawan/semtech"
	"time"
)

type addressKeeper interface {
	lookup(devAddr semtech.DeviceAddress) ([]core.BrokerAddress, error)
	store(devAddr semtech.DeviceAddress, brosAddr ...core.BrokerAddress) error
}

type reddisAddressKeeper struct{} // In a second time

type localDB struct {
	expiryDelay time.Duration
	addresses   map[semtech.DeviceAddress]localEntry
}

type localEntry struct {
	addr  []core.BrokerAddress
	until time.Time
}

// NewLocalDB constructs a new local address keeper
func NewLocalDB(expiryDelay time.Duration) (*localDB, error) {
	if expiryDelay == 0 {
		return nil, fmt.Errorf("Invalid expiration delay")
	}

	return &localDB{
		expiryDelay: expiryDelay,
		addresses:   make(map[semtech.DeviceAddress]localEntry),
	}, nil
}

// lookup implements the addressKeeper interface
func (a *localDB) lookup(devAddr semtech.DeviceAddress) ([]core.BrokerAddress, error) {
	entry, ok := a.addresses[devAddr]
	if !ok {
		return nil, fmt.Errorf("Device address not found")
	}

	if entry.until.Before(time.Now()) {
		delete(a.addresses, devAddr)
		return nil, fmt.Errorf("Broker address(es) expired")
	}

	return entry.addr, nil
}

// store implements the addressKeeper interface
func (a *localDB) store(devAddr semtech.DeviceAddress, brosAddr ...core.BrokerAddress) error {
	_, ok := a.addresses[devAddr]
	if ok {
		return fmt.Errorf("An entry already exists for that device")
	}

	a.addresses[devAddr] = localEntry{
		addr:  brosAddr,
		until: time.Now().Add(a.expiryDelay),
	}

	return nil
}

// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package router

import (
	"time"

	. "github.com/TheThingsNetwork/ttn/refactor"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	dbutil "github.com/TheThingsNetwork/ttn/utils/storage"
	"github.com/brocaar/lorawan"
)

type Storage interface {
	Lookup(devEUI lorawan.EUI64) (entry, error)
	Store(reg Registration) error
}

type entry struct {
	Recipient Recipient
	until     time.Time
}

type storage struct {
	dbutil.Interface
	Name string
}

// newStorage creates a new internal storage for the router
func newStorage(name string, delay time.Duration) (Storage, error) {
	itf, err := dbutil.New(name)
	if err != nil {
		return nil, errors.New(errors.Operational, err)
	}

	tableName := "brokers"
	if err := itf.Init(tableName); err != nil {
		return nil, errors.New(errors.Operational, err)
	}

	return storage{Interface: itf, Name: tableName}, nil
}

// Lookup implements the router.Storage interface
func (s storage) Lookup(devEUI lorawan.EUI64) (entry, error) {
	return entry{}, nil
}

// Store implements the router.Storage interface
func (s storage) Store(reg Registration) error {
	return nil
}

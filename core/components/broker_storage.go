// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package components

import (
	"time"

	"github.com/boltdb/bolt"
	"github.com/brocaar/lorawan"
)

type BrokerStorage interface {
	Close() error
	Lookup(devAddr lorawan.DevAddr) ([]brokerEntry, error)
	Reset() error
	Store(devAddr lorawan.DevAddr, entry brokerEntry) error
}

type brokerBoltStorage struct {
	*bolt.DB
}

type brokerEntry struct {
	Id      string
	NwkSKey lorawan.AES128Key
	Url     string
}

func NewBrokerStorage() (BrokerStorage, error) {
	db, err := bolt.Open("broker_storage.db", 0600, &bolt.Options{Timeout: time.Second})
	if err != nil {
		return nil, err
	}

	if err := initDB(db, "devices"); err != nil {
		return nil, err
	}

	return &brokerBoltStorage{DB: db}, nil
}

func (s brokerBoltStorage) Lookup(devAddr lorawan.DevAddr) ([]brokerEntry, error) {
	entries, err := lookup(s.DB, "devices", devAddr, &brokerEntry{})
	if err != nil {
		return nil, err
	}
	return entries.([]brokerEntry), nil
}

func (s brokerBoltStorage) Store(devAddr lorawan.DevAddr, entry brokerEntry) error {
	return store(s.DB, "devices", devAddr, &entry)
}

func (s brokerBoltStorage) Close() error {
	return s.DB.Close()
}

func (s brokerBoltStorage) Reset() error {
	return resetDB(s.DB, "devices")
}

func (entry brokerEntry) MarshalBinary() ([]byte, error) {
	w := NewEntryReadWriter(nil)
	w.Write(entry.Id)
	w.Write(entry.NwkSKey)
	w.Write(entry.Url)
	return w.Bytes()
}

func (entry *brokerEntry) UnmarshalBinary(data []byte) error {
	if entry == nil || len(data) < 3 {
		return ErrNotUnmarshable
	}
	r := NewEntryReadWriter(data)
	r.Read(func(data []byte) { entry.Id = string(data) })
	r.Read(func(data []byte) { copy(entry.NwkSKey[:], data) })
	r.Read(func(data []byte) { entry.Url = string(data) })
	return r.Err()
}

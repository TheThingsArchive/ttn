// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package components

import (
	"time"

	"github.com/TheThingsNetwork/ttn/core"
	"github.com/boltdb/bolt"
	"github.com/brocaar/lorawan"
)

type routerStorage interface {
	Lookup(devAddr lorawan.DevAddr) (routerEntry, error)
	Store(devAddr lorawan.DevAddr, entry routerEntry) error
}

type routerBoltStorage struct {
	*bolt.DB
	expiryDelay time.Duration
}

type routerEntry struct {
	Recipients []core.Recipient
	until      time.Time
}

func (s routerBoltStorage) Lookup(devAddr lorawan.DevAddr) (routerEntry, error) {
	entries, err := lookup(s.DB, []byte("brokers"), devAddr, &routerEntry{})
	if err != nil {
		return routerEntry{}, err
	}
	routerEntries := entries.([]routerEntry)

	if len(routerEntries) != 1 {
		if err := flush(s.DB, []byte("brokers"), devAddr); err != nil {
			return routerEntry{}, err
		}
		return routerEntry{}, ErrNotFound
	}

	if s.expiryDelay != 0 && routerEntries[0].until.Before(time.Now()) {
		if err := flush(s.DB, []byte("brokers"), devAddr); err != nil {
			return routerEntry{}, err
		}
		return routerEntry{}, ErrEntryExpired
	}

	return routerEntries[0], nil
}

func (s routerBoltStorage) Store(devAddr lorawan.DevAddr, entry routerEntry) error {
	entry.until = time.Now().Add(s.expiryDelay)
	return store(s.DB, []byte("brokers"), devAddr, &entry)
}

func (entry routerEntry) MarshalBinary() ([]byte, error) {
	w := NewEntryReadWriter(nil)
	w.DirectWrite(uint8(len(entry.Recipients)))
	for _, r := range entry.Recipients {
		rawId := []byte(r.Id.(string))
		rawAddress := []byte(r.Address.(string))
		w.Write(rawId)
		w.Write(rawAddress)
	}
	rawTime, err := entry.until.MarshalBinary()
	if err != nil {
		return nil, err
	}
	w.Write(rawTime)
	return w.Bytes()
}

func (entry *routerEntry) UnmarshalBinary(data []byte) error {
	if entry == nil || len(data) < 1 {
		return ErrNotUnmarshable
	}
	r := NewEntryReadWriter(data[0:])
	for i := 0; i < int(data[0]); i += 1 {
		var id, address string
		r.Read(func(data []byte) { id = string(data) })
		r.Read(func(data []byte) { address = string(address) })
		entry.Recipients = append(entry.Recipients, core.Recipient{
			Id:      id,
			Address: address,
		})
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

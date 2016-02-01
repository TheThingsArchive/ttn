// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package components

import (
	"bytes"
	"encoding/binary"
	"time"

	"github.com/TheThingsNetwork/ttn/core"
	"github.com/boltdb/bolt"
	"github.com/brocaar/lorawan"
)

const routerBucket = "brokers"

type boltStorage struct {
	db          *bolt.DB
	expiryDelay time.Duration
}

func NewRouterBolt(expiryDelay time.Duration) (routerStorage, error) {
	db, err := bolt.Open("router_storage.db", 0600, &bolt.Options{Timeout: time.Second})
	if err != nil {
		return nil, err
	}
	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(routerBucket))
		return err
	})
	if err != nil {
		return nil, err
	}

	return boltStorage{
		expiryDelay: expiryDelay,
		db:          db,
	}, nil
}

func (s boltStorage) lookup(devAddr lorawan.DevAddr) ([]core.Recipient, error) {
	var rawEntry []byte
	entry := routerEntry{}

	err := s.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(routerBucket))
		if bucket == nil {
			return ErrStorageUnreachable
		}
		rawEntry = bucket.Get(devAddr[:])
		if rawEntry == nil {
			return ErrNotFound
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	err = entry.UnmarshalBinary(rawEntry)
	if err != nil {
		return nil, err
	}

	if s.expiryDelay != 0 && entry.until.Before(time.Now()) {
		err := s.db.Update(func(tx *bolt.Tx) error {
			bucket := tx.Bucket([]byte(routerBucket))
			if bucket == nil {
				return ErrStorageUnreachable
			}
			return bucket.Delete(devAddr[:])
		})
		if err != nil {
			return nil, err
		}
		return nil, ErrEntryExpired
	}

	return entry.recipients, nil
}

func (s boltStorage) store(devAddr lorawan.DevAddr, recipients ...core.Recipient) error {
	err := s.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(routerBucket))
		if bucket == nil {
			return ErrStorageUnreachable
		}
		entry := bucket.Get(devAddr[:])
		if entry != nil {
			return ErrAlreadyExists
		}

		return nil
	})

	if err != nil {
		return err
	}

	err = s.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(routerBucket))
		if bucket == nil {
			return ErrStorageUnreachable
		}
		rawEntry, err := routerEntry{
			recipients: recipients,
			until:      time.Now().Add(s.expiryDelay),
		}.MarshalBinary()
		return bucket.Put(devAddr[:], rawEntry)
	})

	return err
}

func (entry routerEntry) MarshalBinary() ([]byte, error) {
	rawTime, err := entry.until.MarshalBinary()
	if err != nil {
		return nil, err
	}
	raw := new(bytes.Buffer)

	writeToRaw := func(content interface{}) {
		if err != nil {
			return
		}
		err = binary.Write(raw, binary.LittleEndian, content)
	}

	writeToRaw(uint64(len(entry.recipients)))

	for _, recipient := range entry.recipients {
		rawId := []byte(recipient.Id.(string))
		rawAddress := []byte(recipient.Address.(string))
		writeToRaw(uint64(len(rawId)))
		writeToRaw(rawId)
		writeToRaw(uint64(len(rawAddress)))
		writeToRaw(rawAddress)
	}

	writeToRaw(rawTime)

	if err != nil {
		return nil, err
	}
	return raw.Bytes(), nil
}

func (entry *routerEntry) UnmarshalBinary(data []byte) error {
	if entry == nil || len(data) == 0 {
		return ErrNotUnmarshable
	}
	cursor := 1
	for i := 0; i < int(data[0]); i += 1 {
		lenId := int(data[cursor])
		cursor += 1
		id := data[cursor : cursor+lenId]
		cursor += lenId
		lenAddr := int(data[cursor])
		cursor := 1
		addr := data[cursor : cursor+lenAddr]
		entry.recipients = append(entry.recipients, core.Recipient{
			Id:      string(id),
			Address: string(addr),
		})
		cursor += lenAddr
	}

	entry.until = time.Time{}
	return entry.until.UnmarshalBinary(data[cursor:])
}

// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package components

import (
	"bytes"
	"encoding/binary"
	"io"
	"time"

	"github.com/boltdb/bolt"
	"github.com/brocaar/lorawan"
)

const brokerBucket = "devices"

type boltBrokerStorage struct {
	db *bolt.DB
}

func NewBrokerBolt(expiryDelay time.Duration) (brokerStorage, error) {
	db, err := bolt.Open("router_storage.db", 0600, &bolt.Options{Timeout: time.Second})
	if err != nil {
		return nil, err
	}
	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(brokerBucket))
		return err
	})
	if err != nil {
		return nil, err
	}

	return boltBrokerStorage{
		db: db,
	}, nil
}

func (s boltBrokerStorage) lookup(devAddr lorawan.DevAddr) ([]brokerEntry, error) {
	var rawEntry []byte
	err := s.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(brokerBucket))
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

	buf := bytes.NewBuffer(rawEntry)
	var entries []brokerEntry
	for {
		lenEntry := new(uint16)
		if err := binary.Read(buf, binary.BigEndian, lenEntry); err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		entry := new(brokerEntry)
		if err := binary.Read(buf, binary.BigEndian, entry); err != nil {
			return nil, err
		}
		entries = append(entries, *entry)
	}

	return entries, nil
}

func (s boltBrokerStorage) store(devAddr lorawan.DevAddr, entry brokerEntry) error {
	marshalled, err := entry.MarshalBinary()
	if err != nil {
		return err
	}

	err = s.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(brokerBucket))
		if bucket == nil {
			return ErrStorageUnreachable
		}
		buf := bytes.NewBuffer(bucket.Get(devAddr[:]))
		binary.Write(buf, binary.BigEndian, len(marshalled))
		binary.Write(buf, binary.BigEndian, marshalled)
		return bucket.Put(devAddr[:], buf.Bytes())
	})

	return err
}

func (entry brokerEntry) MarshalBinary() ([]byte, error) {
	data := new(bytes.Buffer)
	var err error
	writeToData := func(content interface{}) {
		if err != nil {
			return
		}
		err = binary.Write(data, binary.BigEndian, content)
	}

	writeToData(uint16(len(entry.Id)))
	writeToData(entry.Id)
	writeToData(uint16(len(entry.Url)))
	writeToData(entry.Url)
	writeToData(uint16(len(entry.NwsKey)))
	writeToData(entry.NwsKey)
	if err != nil {
		return nil, err
	}
	return data.Bytes(), nil
}

func (entry *brokerEntry) UnmarshalBinary(data []byte) error {
	if entry == nil || len(data) < 3 {
		return ErrNotUnmarshable
	}

	buf := bytes.NewBuffer(data)
	var err error
	readFromData := func(to func(data []byte)) {
		if err != nil {
			return
		}
		lenTo := new(uint16)
		if err = binary.Read(buf, binary.BigEndian, lenTo); err != nil {
			return
		}
		to(buf.Next(int(*lenTo)))
	}

	readFromData(func(data []byte) { entry.Id = string(data) })
	readFromData(func(data []byte) { entry.Url = string(data) })
	readFromData(func(data []byte) { copy(entry.NwsKey[:], data) })

	return err
}

// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package mqtt

import (
	"fmt"

	. "github.com/TheThingsNetwork/ttn/core/errors"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	db "github.com/TheThingsNetwork/ttn/utils/storage"
)

// !!!!!!!!!!!!!!!!!!!!!!!!
// NOTE THIS IS NOT THE ADAPTER ROLE -> SHOULD BE MOVED TO HANDLER
// !!!!!!!!!!!!!!!!!!!!!!!!

// Storage defines an interface for the mqtt adapter
type Storage interface {
	Push(topic string, data []byte) error
	Pull(topic string) ([]byte, error)
}

// type storage materializes a concrete mqtt.Storage
type storage struct {
	db.Interface
	Name string
}

// Storage entry implements the storage.StorageEntry interface
type storageEntry struct {
	Data []byte
}

// NewStorage creates a new mqtt.Storage
func NewStorage(name string) (Storage, error) {
	itf, err := db.New(name)
	if err != nil {
		return nil, errors.New(ErrFailedOperation, err)
	}

	tableName := "mqtt_adapter"
	if err := itf.Init(tableName); err != nil {
		return nil, errors.New(ErrFailedOperation, err)
	}
	return storage{Interface: itf, Name: tableName}, nil
}

// Push implements the Storage interface
func (s storage) Push(topic string, data []byte) error {
	err := s.Store(s.Name, []byte(topic), []db.StorageEntry{&storageEntry{data}})
	if err != nil {
		return errors.New(ErrFailedOperation, err)
	}
	return nil
}

// Pull implements the Storage interface
func (s storage) Pull(topic string) ([]byte, error) {
	entries, err := s.Lookup(s.Name, []byte(topic), &storageEntry{})
	if err != nil {
		return nil, errors.New(ErrFailedOperation, err)
	}

	packets, ok := entries.([]*storageEntry)
	if !ok {
		return nil, errors.New(ErrFailedOperation, "Unable to retrieve data from db")
	}

	// NOTE: one day, those entry will be more complicated, with a ttl.
	// Here's the place where we should check for that. Cheers.
	if len(packets) == 0 {
		return nil, errors.New(ErrWrongBehavior, fmt.Sprintf("Entry not found for %s", topic))
	}

	pkt := packets[0]

	var newEntries []db.StorageEntry
	for _, p := range packets[1:] {
		newEntries = append(newEntries, p)
	}

	if err := s.Replace(s.Name, []byte(topic), newEntries); err != nil {
		return nil, errors.New(ErrFailedOperation, "Unable to restore data in db")
	}

	return pkt.Data, nil
}

// MarshalBinary implements the encoding.BinaryMarshaler interface
func (e storageEntry) MarshalBinary() ([]byte, error) {
	return e.Data, nil
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface
func (e *storageEntry) UnmarshalBinary(data []byte) error {
	if e == nil {
		return errors.New(ErrInvalidStructure, "Unable to unmarshal nil entry")
	}
	e.Data = data
	return nil
}

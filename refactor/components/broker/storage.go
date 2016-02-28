// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package broker

import (
	. "github.com/TheThingsNetwork/ttn/refactor"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	"github.com/TheThingsNetwork/ttn/utils/readwriter"
	dbutil "github.com/TheThingsNetwork/ttn/utils/storage"
	"github.com/brocaar/lorawan"
)

type Storage interface {
	Lookup(devEUI lorawan.EUI64) ([]entry, error)
	Store(reg BRegistration) error
	Close() error
}

type entry struct {
	Recipient []byte
	AppEUI    lorawan.EUI64
	DevEUI    lorawan.EUI64
	NwkSKey   lorawan.AES128Key
}

type storage struct {
	db   dbutil.Interface
	Name string
}

// NewStorage constructs a new broker storage
func NewStorage(name string) (Storage, error) {
	itf, err := dbutil.New(name)
	if err != nil {
		return nil, errors.New(errors.Operational, err)
	}

	tableName := "handlers"
	if err := itf.Init(tableName); err != nil {
		return nil, errors.New(errors.Operational, err)
	}

	return storage{db: itf, Name: tableName}, nil
}

// Lookup implements the broker.Storage interface
func (s storage) Lookup(devEUI lorawan.EUI64) ([]entry, error) {
	entries, err := s.db.Lookup(s.Name, devEUI[:], &entry{})
	if err != nil {
		return nil, errors.New(errors.Operational, err)
	}
	return entries.([]entry), nil
}

// Store implements the broker.Storage interface
func (s storage) Store(reg BRegistration) error {
	data, err := reg.Recipient().MarshalBinary()
	if err != nil {
		return errors.New(errors.Structural, err)
	}

	key := reg.DevEUI()
	return s.db.Store(s.Name, key[:], []dbutil.Entry{
		&entry{
			Recipient: data,
			AppEUI:    reg.AppEUI(),
			DevEUI:    key,
			NwkSKey:   reg.NwkSKey(),
		},
	})
}

// Close implements the broker.Storage interface
func (s storage) Close() error {
	return s.db.Close()
}

// MarshalBinary implements the encoding.BinaryMarshaler interface
func (e entry) MarshalBinary() ([]byte, error) {
	rw := readwriter.New(nil)
	rw.Write(e.Recipient)
	rw.Write(e.AppEUI)
	rw.Write(e.DevEUI)
	rw.Write(e.NwkSKey)
	return rw.Bytes()
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface
func (e *entry) UnmarshalBinary(data []byte) error {
	rw := readwriter.New(data)
	rw.Read(func(data []byte) { e.Recipient = data })
	rw.Read(func(data []byte) { copy(e.AppEUI[:], data) })
	rw.Read(func(data []byte) { copy(e.DevEUI[:], data) })
	rw.Read(func(data []byte) { copy(e.NwkSKey[:], data) })
	return rw.Err()
}

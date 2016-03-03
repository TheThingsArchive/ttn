// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package broker

import (
	"github.com/TheThingsNetwork/ttn/core"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	"github.com/TheThingsNetwork/ttn/utils/readwriter"
	dbutil "github.com/TheThingsNetwork/ttn/utils/storage"
	"github.com/brocaar/lorawan"
)

// Storage gives a facade for manipulating the broker database
type Storage interface {
	LookupDevices(devEUI lorawan.EUI64) ([]devEntry, error)
	LookupApplication(appEUI lorawan.EUI64) (appEntry, error)
	StoreDevice(reg core.BRegistration) error
	StoreApplication(reg core.ARegistration) error
	Close() error
}

type devEntry struct {
	Recipient []byte
	AppEUI    lorawan.EUI64
	DevEUI    lorawan.EUI64
	NwkSKey   lorawan.AES128Key
}

type appEntry struct {
	Recipient []byte
	AppEUI    lorawan.EUI64
}

type storage struct {
	db           dbutil.Interface
	Devices      string
	Applications string
}

// NewStorage constructs a new broker storage
func NewStorage(name string) (Storage, error) {
	itf, err := dbutil.New(name)
	if err != nil {
		return nil, errors.New(errors.Operational, err)
	}

	return storage{db: itf, Devices: "Devices", Applications: "Applications"}, nil
}

// LookupDevices implements the broker.Storage interface
func (s storage) LookupDevices(devEUI lorawan.EUI64) ([]devEntry, error) {
	entries, err := s.db.Lookup(s.Devices, devEUI[:], &devEntry{})
	if err != nil {
		return nil, err
	}
	return entries.([]devEntry), nil
}

// LookupApplication implements the broker.Storage interface
func (s storage) LookupApplication(appEUI lorawan.EUI64) (appEntry, error) {
	itf, err := s.db.Lookup(s.Applications, appEUI[:], &appEntry{})
	if err != nil {
		return appEntry{}, err
	}

	entries := itf.([]appEntry)
	if len(entries) != 1 {
		// NOTE Shall we reset the entry ?
		return appEntry{}, errors.New(errors.Structural, "Invalid application entries")
	}

	return entries[0], nil
}

// StoreDevice implements the broker.Storage interface
func (s storage) StoreDevice(reg core.BRegistration) error {
	data, err := reg.Recipient().MarshalBinary()
	if err != nil {
		return errors.New(errors.Structural, err)
	}

	devEUI := reg.DevEUI()
	return s.db.Store(s.Devices, devEUI[:], []dbutil.Entry{
		&devEntry{
			Recipient: data,
			AppEUI:    reg.AppEUI(),
			DevEUI:    devEUI,
			NwkSKey:   reg.NwkSKey(),
		},
	})
}

// StoreApplication implements the broker.Storage interface
func (s storage) StoreApplication(reg core.ARegistration) error {
	data, err := reg.Recipient().MarshalBinary()
	if err != nil {
		return errors.New(errors.Structural, err)
	}

	appEUI := reg.AppEUI()
	return s.db.Replace(s.Applications, appEUI[:], []dbutil.Entry{
		&appEntry{
			Recipient: data,
			AppEUI:    appEUI,
		},
	})
}

// Close implements the broker.Storage interface
func (s storage) Close() error {
	return s.db.Close()
}

// MarshalBinary implements the encoding.BinaryMarshaler interface
func (e devEntry) MarshalBinary() ([]byte, error) {
	rw := readwriter.New(nil)
	rw.Write(e.Recipient)
	rw.Write(e.AppEUI)
	rw.Write(e.DevEUI)
	rw.Write(e.NwkSKey)
	return rw.Bytes()
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface
func (e *devEntry) UnmarshalBinary(data []byte) error {
	rw := readwriter.New(data)
	rw.Read(func(data []byte) { e.Recipient = data })
	rw.Read(func(data []byte) { copy(e.AppEUI[:], data) })
	rw.Read(func(data []byte) { copy(e.DevEUI[:], data) })
	rw.Read(func(data []byte) { copy(e.NwkSKey[:], data) })
	return rw.Err()
}

// MarshalBinary implements the encoding.BinaryMarshaler interface
func (e appEntry) MarshalBinary() ([]byte, error) {
	rw := readwriter.New(nil)
	rw.Write(e.Recipient)
	rw.Write(e.AppEUI)
	return rw.Bytes()
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface
func (e *appEntry) UnmarshalBinary(data []byte) error {
	rw := readwriter.New(data)
	rw.Read(func(data []byte) { e.Recipient = data })
	rw.Read(func(data []byte) { copy(e.AppEUI[:], data) })
	return rw.Err()
}

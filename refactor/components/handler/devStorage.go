// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package handler

import (
	. "github.com/TheThingsNetwork/ttn/refactor"
	"github.com/TheThingsNetwork/ttn/utils/readwriter"
	dbutil "github.com/TheThingsNetwork/ttn/utils/storage"
	"github.com/brocaar/lorawan"
)

type DevStorage interface {
	Lookup(appEUI lorawan.EUI64, devEUI lorawan.EUI64) (devEntry, error)
	Store(r HRegistration) error
}

type devEntry struct {
	Recipient []byte
	AppSKey   lorawan.AES128Key
	NwkSKey   lorawan.AES128Key
}

type appEntry struct {
	AppKey lorawan.AES128Key
}

type devStorage struct {
	db dbutil.Interface
}

// NewDevStorage creates a new Device Storage for handler
func NewDevStorage(name string) (DevStorage, error) {
	return nil, nil
}

// Lookup implements the handler.DevStorage interface
func (s devStorage) Lookup(appEUI lorawan.EUI64, devEUI lorawan.EUI64) (devEntry, error) {
	return devEntry{}, nil
}

// Store implements the handler.DevStorage interface
func (s devStorage) Store(reg HRegistration) error {
	return nil
}

// Close implements the handler.DevStorage interface
func (s devStorage) Close() error {
	return s.db.Close()
}

// MarshalBinary implements the encoding.BinaryMarshaler interface
func (e devEntry) MarshalBinary() ([]byte, error) {
	rw := readwriter.New(nil)
	rw.Write(e.Recipient)
	rw.Write(e.AppSKey)
	rw.Write(e.NwkSKey)
	return rw.Bytes()
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface
func (e *devEntry) UnmarshalBinary(data []byte) error {
	rw := readwriter.New(data)
	rw.Read(func(data []byte) { e.Recipient = data })
	rw.Read(func(data []byte) { copy(e.AppSKey[:], data) })
	rw.Read(func(data []byte) { copy(e.NwkSKey[:], data) })
	return rw.Err()
}

// MarshalBinary implements the encoding.BinaryMarshaler interface
func (e appEntry) MarshalBinary() ([]byte, error) {
	rw := readwriter.New(nil)
	rw.Write(e.AppKey)
	return rw.Bytes()
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface
func (e *appEntry) UnmarshalBinary(data []byte) error {
	rw := readwriter.New(data)
	rw.Read(func(data []byte) { copy(e.AppKey[:], data) })
	return rw.Err()
}

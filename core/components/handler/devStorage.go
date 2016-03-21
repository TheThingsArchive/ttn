// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package handler

import (
	"encoding"
	"encoding/binary"

	dbutil "github.com/TheThingsNetwork/ttn/core/storage"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	"github.com/TheThingsNetwork/ttn/utils/readwriter"
)

// DevStorage gives a facade to manipulate the handler devices database
type DevStorage interface {
	read(appEUI []byte, devEUI []byte) (devEntry, error)
	upsert(entry devEntry) error
	done() error
}

const dbDevices = "devices"

type devEntry struct {
	AppEUI   []byte
	AppKey   [16]byte
	AppSKey  [16]byte
	DevAddr  []byte
	DevEUI   []byte
	FCntDown uint32
	NwkSKey  [16]byte
}

type devStorage struct {
	db dbutil.Interface
}

// NewDevStorage creates a new Device Storage for handler
func NewDevStorage(name string) (DevStorage, error) {
	itf, err := dbutil.New(name)
	if err != nil {
		return nil, errors.New(errors.Operational, err)
	}

	return &devStorage{db: itf}, nil
}

// read implements the handler.DevStorage interface
func (s *devStorage) read(appEUI []byte, devEUI []byte) (devEntry, error) {
	itf, err := s.db.Read(nil, &devEntry{}, appEUI, devEUI)
	if err != nil {
		return devEntry{}, err
	}
	entries, ok := itf.([]devEntry)
	if !ok || len(entries) != 1 {
		return devEntry{}, errors.New(errors.Structural, "Invalid stored entry")
	}
	return entries[0], nil
}

// upsert implements the handler.DevStorage interface
func (s *devStorage) upsert(entry devEntry) error {
	return s.db.Update(nil, []encoding.BinaryMarshaler{entry}, entry.AppEUI, entry.DevEUI)
}

// done implements the handler.DevStorage interface
func (s *devStorage) done() error {
	return s.db.Close()
}

// MarshalBinary implements the encoding.BinaryMarshaler interface
func (e devEntry) MarshalBinary() ([]byte, error) {
	rw := readwriter.New(nil)
	rw.Write(e.AppKey[:])
	rw.Write(e.AppSKey[:])
	rw.Write(e.NwkSKey[:])
	rw.Write(e.FCntDown)
	rw.Write(e.AppEUI)
	rw.Write(e.DevEUI)
	rw.Write(e.DevAddr)
	return rw.Bytes()
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface
func (e *devEntry) UnmarshalBinary(data []byte) error {
	rw := readwriter.New(data)
	rw.Read(func(data []byte) { copy(e.AppKey[:], data) })
	rw.Read(func(data []byte) { copy(e.AppSKey[:], data) })
	rw.Read(func(data []byte) { copy(e.NwkSKey[:], data) })
	rw.Read(func(data []byte) { e.FCntDown = binary.BigEndian.Uint32(data) })
	rw.Read(func(data []byte) { e.AppEUI = data })
	rw.Read(func(data []byte) { e.DevEUI = data })
	rw.Read(func(data []byte) { e.DevAddr = data })
	return rw.Err()
}

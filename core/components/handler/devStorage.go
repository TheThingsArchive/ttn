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
	readAll(appEUI []byte) ([]devEntry, error)
	upsert(entry devEntry) error
	setDefault(appEUI []byte, entry *devDefaultEntry) error
	getDefault(appEUI []byte) (*devDefaultEntry, error)
	done() error
}

const dbDevices = "devices"

type devEntry struct {
	AppEUI   []byte
	AppKey   *[16]byte
	AppSKey  [16]byte
	DevAddr  []byte
	DevEUI   []byte
	FCntDown uint32
	FCntUp   uint32
	NwkSKey  [16]byte
	DevMode  bool
}

type devDefaultEntry struct {
	AppKey [16]byte
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

func (s *devStorage) read(appEUI []byte, devEUI []byte) (devEntry, error) {
	itf, err := s.db.Read(devEUI, &devEntry{}, appEUI)
	if err != nil {
		return devEntry{}, err
	}
	return itf.([]devEntry)[0], nil // Type and dimensio guaranteed by db.Read()
}

func (s *devStorage) readAll(appEUI []byte) ([]devEntry, error) {
	itf, err := s.db.ReadAll(&devEntry{}, appEUI)
	if err != nil {
		return nil, err
	}
	return itf.([]devEntry), nil
}

func (s *devStorage) upsert(entry devEntry) error {
	return s.db.Update(entry.DevEUI, []encoding.BinaryMarshaler{entry}, entry.AppEUI)
}

func (s *devStorage) setDefault(appEUI []byte, entry *devDefaultEntry) error {
	return s.db.Update([]byte("default"), []encoding.BinaryMarshaler{entry}, appEUI)
}

func (s *devStorage) getDefault(appEUI []byte) (*devDefaultEntry, error) {
	itf, err := s.db.Read([]byte("default"), &devDefaultEntry{}, appEUI)
	if err != nil {
		if ferr, ok := err.(errors.Failure); ok && ferr.Nature == errors.NotFound {
			return nil, nil
		}
		return nil, err
	}
	return &itf.([]devDefaultEntry)[0], nil
}

// done implements the handler.DevStorage interface
func (s *devStorage) done() error {
	return s.db.Close()
}

// MarshalBinary implements the encoding.BinaryMarshaler interface
func (e devEntry) MarshalBinary() ([]byte, error) {
	rw := readwriter.New(nil)
	if e.AppKey != nil {
		rw.Write(e.AppKey[:])
	} else {
		rw.Write([]byte{})
	}
	rw.Write(e.AppSKey[:])
	rw.Write(e.NwkSKey[:])
	rw.Write(e.FCntUp)
	rw.Write(e.FCntDown)
	rw.Write(e.DevMode)
	rw.Write(e.AppEUI)
	rw.Write(e.DevEUI)
	rw.Write(e.DevAddr)
	return rw.Bytes()
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface
func (e *devEntry) UnmarshalBinary(data []byte) error {
	rw := readwriter.New(data)
	rw.Read(func(data []byte) {
		if len(data) == 16 {
			e.AppKey = new([16]byte)
			copy(e.AppKey[:], data)
		} else {
			e.AppKey = nil
		}
	})
	rw.Read(func(data []byte) { copy(e.AppSKey[:], data) })
	rw.Read(func(data []byte) { copy(e.NwkSKey[:], data) })
	rw.Read(func(data []byte) { e.FCntUp = binary.BigEndian.Uint32(data) })
	rw.Read(func(data []byte) { e.FCntDown = binary.BigEndian.Uint32(data) })
	rw.Read(func(data []byte) { e.DevMode = (data[0] != 0) })
	rw.Read(func(data []byte) {
		e.AppEUI = make([]byte, len(data))
		copy(e.AppEUI, data)
	})
	rw.Read(func(data []byte) {
		e.DevEUI = make([]byte, len(data))
		copy(e.DevEUI, data)
	})
	rw.Read(func(data []byte) {
		e.DevAddr = make([]byte, len(data))
		copy(e.DevAddr, data)
	})
	return rw.Err()
}

// MarshalBinary implements the encoding.BinaryMarshaler interface
func (e devDefaultEntry) MarshalBinary() ([]byte, error) {
	rw := readwriter.New(nil)
	rw.Write(e.AppKey[:])
	return rw.Bytes()
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface
func (e *devDefaultEntry) UnmarshalBinary(data []byte) error {
	rw := readwriter.New(data)
	rw.Read(func(data []byte) { copy(e.AppKey[:], data) })
	return rw.Err()
}

// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package broker

import (
	"encoding"

	dbutil "github.com/TheThingsNetwork/ttn/core/storage"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	"github.com/TheThingsNetwork/ttn/utils/readwriter"
)

// NOTE: This is a partial duplication of handler.DevStorage

// AppStorage gives a facade for manipulating the broker applications infos
type AppStorage interface {
	read(appEUI []byte) (appEntry, error)
	upsert(entry appEntry) error
	done() error
}

type appEntry struct {
	Dialer Dialer
	AppEUI []byte
}

type appStorage struct {
	db dbutil.Interface
}

// NewAppStorage constructs a new broker controller
func NewAppStorage(name string) (AppStorage, error) {
	itf, err := dbutil.New(name)
	if err != nil {
		return nil, errors.New(errors.Operational, err)
	}

	return &appStorage{db: itf}, nil
}

// read implements the AppStorage interface
func (s *appStorage) read(appEUI []byte) (appEntry, error) {
	itf, err := s.db.Read(nil, &appEntry{}, appEUI)
	if err != nil {
		return appEntry{}, err
	}
	entries, ok := itf.([]appEntry)
	if !ok || len(entries) != 1 {
		return appEntry{}, errors.New(errors.Structural, "Invalid stored entry")
	}
	return entries[0], nil
}

// upsert implements the AppStorage interface
func (s *appStorage) upsert(entry appEntry) error {
	return s.db.Update(nil, []encoding.BinaryMarshaler{entry}, entry.AppEUI)
}

// done implements the AppStorage interface {
func (s *appStorage) done() error {
	return s.db.Close()
}

// MarshalBinary implements the encoding.BinaryMarshaler interfaceA
func (e appEntry) MarshalBinary() ([]byte, error) {
	rw := readwriter.New(nil)
	rw.Write(e.AppEUI)
	rw.Write(e.Dialer.MarshalSafely())
	return rw.Bytes()
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface
func (e *appEntry) UnmarshalBinary(data []byte) error {
	rw := readwriter.New(data)
	rw.Read(func(data []byte) {
		e.AppEUI = make([]byte, len(data))
		copy(e.AppEUI, data)
	})
	rw.Read(func(data []byte) {
		e.Dialer = NewDialer(data)
	})
	return rw.Err()
}

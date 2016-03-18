// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package handler

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"sync"

	dbutil "github.com/TheThingsNetwork/ttn/core/storage"
	"github.com/TheThingsNetwork/ttn/utils/errors"
)

// DevStorage gives a facade to manipulate the handler devices database
type DevStorage interface {
	UpdateFCnt(appEUI []byte, devEUI []byte, fcnt uint32) error
	Lookup(appEUI []byte, devEUI []byte) (devEntry, error)
	StorePersonalized(appEUI []byte, devAddr [4]byte, appSKey, nwkSKey [16]byte) error
	Close() error
}

const dbDevices = "devices"

type devEntry struct {
	DevAddr  [4]byte
	AppSKey  [16]byte
	NwkSKey  [16]byte
	FCntDown uint32
}

type devStorage struct {
	sync.RWMutex
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

// Lookup implements the handler.DevStorage interface
func (s *devStorage) Lookup(appEUI []byte, devEUI []byte) (devEntry, error) {
	return s.lookup(appEUI, devEUI, true)
}

// lookup allow other method to re-use lookup while holding the lock
func (s *devStorage) lookup(appEUI []byte, devEUI []byte, shouldLock bool) (devEntry, error) {
	if shouldLock {
		s.RLock()
		defer s.RUnlock()
	}
	itf, err := s.db.Lookup(fmt.Sprintf("%x.%x", appEUI, devEUI), []byte(dbDevices), &devEntry{})
	if err != nil {
		return devEntry{}, err // Operational || NotFound
	}
	entries, ok := itf.([]devEntry)
	if !ok || len(entries) != 1 {
		return devEntry{}, errors.New(errors.Structural, "Invalid stored entry")
	}
	return entries[0], nil
}

// StorePersonalized implements the handler.DevStorage interface
func (s *devStorage) StorePersonalized(appEUI []byte, devAddr [4]byte, appSKey, nwkSKey [16]byte) error {
	devEUI := make([]byte, 8, 8)
	copy(devEUI[4:], devAddr[:])
	e := []dbutil.Entry{
		&devEntry{
			AppSKey: appSKey,
			NwkSKey: nwkSKey,
			DevAddr: devAddr,
		},
	}
	s.Lock()
	defer s.Unlock()
	return s.db.Replace(fmt.Sprintf("%x.%x", appEUI, devEUI), []byte(dbDevices), e)
}

// UpdateFCnt implements the handler.DevStorage interface
func (s *devStorage) UpdateFCnt(appEUI []byte, devEUI []byte, fcnt uint32) error {
	s.Lock()
	defer s.Unlock()
	devEntry, err := s.lookup(appEUI, devEUI, false)
	if err != nil {
		return err
	}
	devEntry.FCntDown = fcnt
	return s.db.Replace(fmt.Sprintf("%x.%x", appEUI, devEUI), []byte(dbDevices), []dbutil.Entry{&devEntry})
}

// Close implements the handler.DevStorage interface
func (s *devStorage) Close() error {
	return s.db.Close()
}

// MarshalBinary implements the encoding.BinaryMarshaler interface
func (e devEntry) MarshalBinary() ([]byte, error) {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, e.DevAddr[:]) // 4
	binary.Write(buf, binary.BigEndian, e.AppSKey[:]) // 16
	binary.Write(buf, binary.BigEndian, e.NwkSKey[:]) // 16
	binary.Write(buf, binary.BigEndian, e.FCntDown)   // 4
	if len(buf.Bytes()) != 40 {
		return nil, errors.New(errors.Structural, "Unable to marshal devEntry")
	}
	return buf.Bytes(), nil
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface
func (e *devEntry) UnmarshalBinary(data []byte) error {
	buf := bytes.NewBuffer(data)
	binary.Read(buf, binary.BigEndian, &e.DevAddr)
	binary.Read(buf, binary.BigEndian, &e.AppSKey)
	binary.Read(buf, binary.BigEndian, &e.NwkSKey)
	return binary.Read(buf, binary.BigEndian, &e.FCntDown)
}

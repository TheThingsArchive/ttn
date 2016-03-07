// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package broker

import (
	"bytes"
	"encoding/binary"
	"sync"

	"github.com/TheThingsNetwork/ttn/core"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	"github.com/TheThingsNetwork/ttn/utils/readwriter"
	dbutil "github.com/TheThingsNetwork/ttn/utils/storage"
	"github.com/brocaar/lorawan"
)

// NetworkController gives a facade for manipulating the broker databases and devices
type NetworkController interface {
	UpdateFCnt(appEUI lorawan.EUI64, devEUI lorawan.EUI64, fcnt uint32, dir string) error
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
	FCntUp    uint32
	FCntDown  uint32
}

type appEntry struct {
	Recipient []byte
	AppEUI    lorawan.EUI64
}

type controller struct {
	sync.RWMutex
	db           dbutil.Interface
	Devices      string
	Applications string
}

// NewNetworkController constructs a new broker controller
func NewNetworkController(name string) (NetworkController, error) {
	itf, err := dbutil.New(name)
	if err != nil {
		return nil, errors.New(errors.Operational, err)
	}

	return &controller{db: itf, Devices: "Devices", Applications: "Applications"}, nil
}

// LookupDevices implements the broker.NetworkController interface
func (s *controller) LookupDevices(devEUI lorawan.EUI64) ([]devEntry, error) {
	s.RLock()
	defer s.RUnlock()
	entries, err := s.db.Lookup(s.Devices, devEUI[:], &devEntry{})
	if err != nil {
		return nil, err
	}
	return entries.([]devEntry), nil
}

// LookupApplication implements the broker.NetworkController interface
func (s *controller) LookupApplication(appEUI lorawan.EUI64) (appEntry, error) {
	s.RLock()
	defer s.RUnlock()
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

// UpdateFCnt implements the broker.NetworkController interface
func (s *controller) UpdateFCnt(appEUI lorawan.EUI64, devEUI lorawan.EUI64, fcnt uint32, dir string) error {
	s.Lock()
	defer s.Unlock()
	itf, err := s.db.Lookup(s.Devices, devEUI[:], &devEntry{})
	if err != nil {
		return err
	}
	entries := itf.([]devEntry)

	var newEntries []dbutil.Entry
	for _, e := range entries {
		entry := new(devEntry)
		*entry = e
		if entry.AppEUI == appEUI {
			switch dir {
			case "up":
				entry.FCntUp = fcnt
			case "down":
				entry.FCntDown = fcnt
			default:
				return errors.New(errors.Implementation, "Unreckognized direction")
			}
		}
		newEntries = append(newEntries, entry)
	}

	return s.db.Replace(s.Devices, devEUI[:], newEntries)
}

// StoreDevice implements the broker.NetworkController interface
func (s *controller) StoreDevice(reg core.BRegistration) error {
	s.Lock()
	defer s.Unlock()
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

// StoreApplication implements the broker.NetworkController interface
func (s *controller) StoreApplication(reg core.ARegistration) error {
	s.Lock()
	defer s.Unlock()
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

// Close implements the broker.NetworkController interface
func (s *controller) Close() error {
	return s.db.Close()
}

// MarshalBinary implements the encoding.BinaryMarshaler interface
func (e devEntry) MarshalBinary() ([]byte, error) {
	rw := readwriter.New(nil)
	rw.Write(e.Recipient)
	rw.Write(e.AppEUI)
	rw.Write(e.DevEUI)
	rw.Write(e.NwkSKey)
	rw.Write(e.FCntUp)
	rw.Write(e.FCntDown)
	return rw.Bytes()
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface
func (e *devEntry) UnmarshalBinary(data []byte) error {
	rw := readwriter.New(data)
	rw.Read(func(data []byte) { e.Recipient = data })
	rw.Read(func(data []byte) { copy(e.AppEUI[:], data) })
	rw.Read(func(data []byte) { copy(e.DevEUI[:], data) })
	rw.Read(func(data []byte) { copy(e.NwkSKey[:], data) })
	rw.TryRead(func(data []byte) error {
		buf := new(bytes.Buffer)
		buf.Write(data)
		fcnt := new(uint32)
		if err := binary.Read(buf, binary.BigEndian, fcnt); err != nil {
			return err
		}
		e.FCntUp = *fcnt
		return nil
	})
	rw.TryRead(func(data []byte) error {
		buf := new(bytes.Buffer)
		buf.Write(data)
		fcnt := new(uint32)
		if err := binary.Read(buf, binary.BigEndian, fcnt); err != nil {
			return err
		}
		e.FCntDown = *fcnt
		return nil
	})
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

// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package broker

import (
	"bytes"
	"encoding"
	"encoding/binary"
	"math"
	"reflect"
	"sync"

	dbutil "github.com/TheThingsNetwork/ttn/core/storage"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	"github.com/TheThingsNetwork/ttn/utils/readwriter"
)

// NetworkController gives a facade for manipulating the broker databases and devices
type NetworkController interface {
	read(devAddr []byte) ([]devEntry, error)
	readNonces(appEUI []byte, devEUI []byte) (noncesEntry, error)
	upsertNonces(entry noncesEntry) error
	upsert(entry devEntry) error
	wholeCounter(devCnt uint32, entryCnt uint32) (uint32, error)
	done() error
}

type devEntry struct {
	AppEUI  []byte
	DevEUI  []byte
	DevAddr []byte
	Dialer  Dialer
	FCntUp  uint32
	NwkSKey [16]byte
}

type noncesEntry struct {
	AppEUI    []byte
	DevEUI    []byte
	DevNonces [][]byte
}

type controller struct {
	sync.RWMutex
	db dbutil.Interface
}

var dbDevices = []byte("devices")

// NewNetworkController constructs a new broker controller
func NewNetworkController(name string) (NetworkController, error) {
	itf, err := dbutil.New(name)
	if err != nil {
		return nil, errors.New(errors.Operational, err)
	}

	return &controller{db: itf}, nil
}

// read implements the NetworkController interface
func (s *controller) read(devAddr []byte) ([]devEntry, error) {
	entries, err := s.db.Read(devAddr, &devEntry{}, dbDevices)
	if err != nil {
		return nil, err
	}
	return entries.([]devEntry), nil
}

// wholeCounter implements the broker.NetworkController interface
func (s *controller) wholeCounter(devCnt uint32, entryCnt uint32) (uint32, error) {
	upperSup := int(math.Pow(2, 16))
	diff := int(devCnt) - (int(entryCnt) % upperSup)
	var offset int
	if diff >= 0 {
		offset = diff
	} else {
		offset = upperSup + diff
	}
	if offset > upperSup/4 {
		return 0, errors.New(errors.Structural, "Gap too big, counter is errored")
	}
	return entryCnt + uint32(offset), nil
}

// upsert implements the broker.NetworkController interface
func (s *controller) upsert(update devEntry) error {
	s.Lock()
	defer s.Unlock()
	itf, err := s.db.Read(update.DevAddr, &devEntry{}, dbDevices)
	if err != nil {
		if err.(errors.Failure).Nature != errors.NotFound {
			return err
		}
		return s.db.Update(update.DevAddr, []encoding.BinaryMarshaler{update}, dbDevices)
	}
	entries := itf.([]devEntry)

	var newEntries []encoding.BinaryMarshaler
	var replaced bool
	for _, e := range entries {
		entry := new(devEntry)
		*entry = e
		if reflect.DeepEqual(entry.AppEUI, update.AppEUI) && reflect.DeepEqual(entry.DevEUI, update.DevEUI) {
			newEntries = append(newEntries, update)
			replaced = true
		} else {
			newEntries = append(newEntries, entry)
		}
	}
	if !replaced {
		newEntries = append(newEntries, update)
	}
	return s.db.Update(update.DevAddr, newEntries, dbDevices)
}

// readNonces implements the broker.NetworkController interface
func (s *controller) readNonces(appEUI []byte, devEUI []byte) (noncesEntry, error) {
	itf, err := s.db.Read(nil, &noncesEntry{}, appEUI, devEUI)
	if err != nil {
		return noncesEntry{}, err
	}
	return itf.([]noncesEntry)[0], nil // Storage guarantee to have one entry
}

// upsertNonces implements the broker.NetworkController interface
func (s *controller) upsertNonces(entry noncesEntry) error {
	return s.db.Update(nil, []encoding.BinaryMarshaler{entry}, entry.AppEUI, entry.DevEUI)
}

// done implements the broker.NetworkController interface
func (s *controller) done() error {
	return s.db.Close()
}

// MarshalBinary implements the encoding.BinaryMarshaler interface
func (e devEntry) MarshalBinary() ([]byte, error) {
	rw := readwriter.New(nil)
	rw.Write(e.AppEUI)
	rw.Write(e.DevEUI)
	rw.Write(e.DevAddr)
	rw.Write(e.NwkSKey[:])
	rw.Write(e.FCntUp)
	rw.Write(e.Dialer.MarshalSafely())
	return rw.Bytes()
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface
func (e *devEntry) UnmarshalBinary(data []byte) error {
	rw := readwriter.New(data)
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
	rw.Read(func(data []byte) { copy(e.NwkSKey[:], data) })
	rw.Read(func(data []byte) { e.FCntUp = binary.BigEndian.Uint32(data) })
	rw.Read(func(data []byte) { e.Dialer = NewDialer(data) })
	return rw.Err()
}

// MarshalBinary implements the encoding.BinaryMarshaler interface
func (e noncesEntry) MarshalBinary() ([]byte, error) {
	rw := readwriter.New(nil)
	rw.Write(e.AppEUI)
	rw.Write(e.DevEUI)
	buf := new(bytes.Buffer)
	for _, n := range e.DevNonces {
		_, _ = buf.Write(n)
	}
	rw.Write(buf.Bytes())
	return rw.Bytes()
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface
func (e *noncesEntry) UnmarshalBinary(data []byte) error {
	rw := readwriter.New(data)
	rw.Read(func(data []byte) {
		e.AppEUI = make([]byte, len(data))
		copy(e.AppEUI, data)
	})
	rw.Read(func(data []byte) {
		e.DevEUI = make([]byte, len(data))
		copy(e.DevEUI, data)
	})
	rw.Read(func(data []byte) {
		n := len(data) / 2 // DevNonce -> 2-bytes
		for i := 0; i < int(n); i++ {
			devNonce := make([]byte, 2, 2)
			copy(devNonce, data[2*i:2*i+2])
			e.DevNonces = append(e.DevNonces, devNonce)
		}
	})
	return rw.Err()
}

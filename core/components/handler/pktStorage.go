// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package handler

import (
	"encoding"
	"sync"
	"time"

	dbutil "github.com/TheThingsNetwork/ttn/core/storage"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	"github.com/TheThingsNetwork/ttn/utils/readwriter"
)

// PktStorage gives a facade to manipulate the handler packets database
type PktStorage interface {
	enqueue(entry pktEntry) error
	dequeue(appEUI []byte, devEUI []byte) (pktEntry, error)
	peek(appEUI []byte, devEUI []byte) (pktEntry, error)
	done() error
}

const dbPackets = "packets"

type pktStorage struct {
	sync.RWMutex
	size uint
	db   dbutil.Interface
}

type pktEntry struct {
	AppEUI  []byte
	DevEUI  []byte
	Payload []byte
	TTL     time.Time
}

// NewPktStorage creates a new PktStorage
func NewPktStorage(name string, size uint) (PktStorage, error) {
	itf, err := dbutil.New(name)
	if err != nil {
		return nil, errors.New(errors.Operational, err)
	}
	return &pktStorage{db: itf, size: size}, nil
}

func filterExpired(entries []pktEntry) []pktEntry {
	var filtered []pktEntry
	now := time.Now()
	for _, e := range entries {
		if e.TTL.After(now) {
			filtered = append(filtered, e)
		}
	}
	return filtered
}

func pop(entries []pktEntry) (pktEntry, []encoding.BinaryMarshaler) {
	head := entries[0]
	var tail []encoding.BinaryMarshaler
	for _, e := range entries[1:] {
		tail = append(tail, e)
	}
	return head, tail
}

// enqueue implements the PktStorage interface
func (s *pktStorage) enqueue(entry pktEntry) error {
	s.Lock()
	defer s.Unlock()
	itf, err := s.db.Read(entry.DevEUI, &pktEntry{}, entry.AppEUI)
	if err != nil && err.(errors.Failure).Nature != errors.NotFound {
		return err
	}
	var entries []pktEntry
	if itf != nil {
		entries = filterExpired(itf.([]pktEntry))
	}
	if len(entries) >= int(s.size) {
		_, tail := pop(entries)
		return s.db.Update(entry.DevEUI, append(tail, entry), entry.AppEUI)
	}
	// NOTE: We append, even if there're still expired entries, we'll filter them
	// during dequeuing
	return s.db.Append(entry.DevEUI, []encoding.BinaryMarshaler{entry}, entry.AppEUI)
}

// dequeue implements the PktStorage interface
func (s *pktStorage) dequeue(appEUI []byte, devEUI []byte) (pktEntry, error) {
	s.Lock()
	defer s.Unlock()
	itf, err := s.db.Read(devEUI, &pktEntry{}, appEUI)
	if err != nil {
		return pktEntry{}, err
	}
	entries := itf.([]pktEntry)
	filtered := filterExpired(entries)

	if len(filtered) < 1 { // No entry left, return NotFound
		if len(entries) != len(filtered) { // Get rid of expired entries
			var replaces []encoding.BinaryMarshaler
			for _, e := range filtered {
				replaces = append(replaces, e)
			}
			_ = s.db.Update(devEUI, replaces, appEUI)
		}
		return pktEntry{}, errors.New(errors.NotFound, "There's no available entry")
	}

	// Otherwise dequeue the first one and send it
	head, tail := pop(filtered)
	if err := s.db.Update(devEUI, tail, appEUI); err != nil {
		return pktEntry{}, err
	}
	return head, nil
}

// peek implements the PktStorage interface
func (s *pktStorage) peek(appEUI []byte, devEUI []byte) (pktEntry, error) {
	s.RLock()
	defer s.RUnlock()
	itf, err := s.db.Read(devEUI, &pktEntry{}, appEUI)
	if err != nil {
		return pktEntry{}, err
	}
	entries := filterExpired(itf.([]pktEntry))
	if len(entries) < 1 {
		return pktEntry{}, errors.New(errors.NotFound, "There's no available entry")
	}
	return entries[0], nil
}

// done implements the PktStorage interface
func (s *pktStorage) done() error {
	return s.db.Close()
}

// MarshalBinary implements the encoding.BinaryMarshaler interface
func (e pktEntry) MarshalBinary() ([]byte, error) {
	data, err := e.TTL.MarshalBinary()
	if err != nil {
		return nil, errors.New(errors.Structural, err)
	}
	rw := readwriter.New(nil)
	rw.Write(e.AppEUI)
	rw.Write(e.DevEUI)
	rw.Write(e.Payload)
	rw.Write(data)
	return rw.Bytes()
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface
func (e *pktEntry) UnmarshalBinary(data []byte) error {
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
		e.Payload = make([]byte, len(data))
		copy(e.Payload, data)
	})
	rw.TryRead(func(data []byte) error { return e.TTL.UnmarshalBinary(data) })
	return rw.Err()
}

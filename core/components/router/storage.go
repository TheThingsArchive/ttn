// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package router

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"sync"
	"time"

	"github.com/KtorZ/rpc/core"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	dbutil "github.com/TheThingsNetwork/ttn/utils/storage"
)

// Storage gives a facade to manipulate the router database
type Storage interface {
	Lookup(devAddr []byte) ([]entry, error)
	Store(devAddr []byte, brokerIndex int) error
	LookupStats(gid []byte) (core.StatsMetadata, error)
	UpdateStats(gid []byte, metadata core.StatsMetadata) error
	Close() error
}

type entry struct {
	BrokerIndex int
	until       time.Time
}

type storage struct {
	sync.Mutex
	db          dbutil.Interface
	ExpiryDelay time.Duration
}

const (
	dbBrokers = "brokers"
	dbGateway = "gateways"
)

// NewStorage creates a new internal storage for the router
func NewStorage(name string, delay time.Duration) (Storage, error) {
	itf, err := dbutil.New(name)
	if err != nil {
		return nil, errors.New(errors.Operational, err)
	}

	return &storage{db: itf, ExpiryDelay: delay}, nil
}

// UpdateStats implements the router.Storage interface
func (s *storage) UpdateStats(gid []byte, metadata core.StatsMetadata) error {
	return s.db.Replace(dbGateway, gid, []dbutil.Entry{&metadata})
}

// LookupStats implements the router.Storage interface
func (s *storage) LookupStats(gid []byte) (core.StatsMetadata, error) {
	itf, err := s.db.Lookup(dbGateway, gid, &core.StatsMetadata{})
	if err != nil {
		return core.StatsMetadata{}, err
	}
	entries := itf.([]core.StatsMetadata)
	if len(entries) == 0 {
		return core.StatsMetadata{}, errors.New(errors.NotFound, "Not entry found for given gateway")
	}
	return entries[0], nil
}

// Lookup implements the router.Storage interface
func (s *storage) Lookup(devAddr []byte) ([]entry, error) {
	s.Lock()
	defer s.Unlock()
	itf, err := s.db.Lookup(dbBrokers, devAddr, &entry{})
	if err != nil {
		return nil, err
	}
	entries := itf.([]entry)

	if s.ExpiryDelay != 0 {
		var newEntries []dbutil.Entry
		var filtered []entry
		for _, e := range entries {
			if e.until.After(time.Now()) {
				newEntry := new(entry)
				*newEntry = e
				newEntries = append(newEntries, newEntry)
				filtered = append(filtered, e)
			}
		}
		if err := s.db.Replace(dbBrokers, devAddr, newEntries); err != nil {
			return nil, errors.New(errors.Operational, err)
		}
		entries = filtered
	}

	if len(entries) == 0 {
		return nil, errors.New(errors.NotFound, fmt.Sprintf("No entry for: %v", devAddr))
	}
	return entries, nil
}

// Store implements the router.Storage interface
func (s *storage) Store(devAddr []byte, brokerIndex int) error {
	s.Lock()
	defer s.Unlock()
	return s.db.Store(dbBrokers, devAddr, []dbutil.Entry{&entry{
		BrokerIndex: brokerIndex,
		until:       time.Now().Add(s.ExpiryDelay),
	}})
}

// Close implements the router.Storage interface
func (s *storage) Close() error {
	return s.db.Close()
}

// MarshalBinary implements the encoding.BinaryMarshaler interface
func (e entry) MarshalBinary() ([]byte, error) {
	data, err := e.until.MarshalBinary()
	if err != nil {
		return nil, errors.New(errors.Structural, err)
	}

	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, uint16(e.BrokerIndex))
	binary.Write(buf, binary.BigEndian, data)
	return buf.Bytes(), nil
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface
func (e *entry) UnmarshalBinary(data []byte) error {
	buf := bytes.NewBuffer(data)

	// e.until
	tdata := new([]byte)
	binary.Read(buf, binary.BigEndian, tdata)
	if err := e.until.UnmarshalBinary(*tdata); err != nil {
		return errors.New(errors.Structural, err)
	}

	// e.Broker
	index := new(uint16)
	binary.Read(buf, binary.BigEndian, index)
	e.BrokerIndex = int(*index)

	return nil
}

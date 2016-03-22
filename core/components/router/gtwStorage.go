// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package router

import (
	"encoding"

	"github.com/TheThingsNetwork/ttn/core"
	dbutil "github.com/TheThingsNetwork/ttn/core/storage"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	"github.com/TheThingsNetwork/ttn/utils/readwriter"
)

var dbGateways = []byte("gateways")

// GtwStorage gives a facade to manipulate the router's gateways data
type GtwStorage interface {
	read(gid []byte) (gtwEntry, error)
	upsert(entry gtwEntry) error
	done() error
}

type gtwEntry struct {
	GatewayID []byte
	Metadata  core.StatsMetadata
}

type gtwStorage struct {
	db dbutil.Interface
}

// NewGtwStorage creates a new internal storage for the router
func NewGtwStorage(name string) (GtwStorage, error) {
	itf, err := dbutil.New(name)
	if err != nil {
		return nil, errors.New(errors.Operational, err)
	}
	return &gtwStorage{db: itf}, nil
}

// read implements the router.GtwStorage interface {
func (s *gtwStorage) read(gid []byte) (gtwEntry, error) {
	itf, err := s.db.Read(gid, &gtwEntry{}, dbGateways)
	if err != nil {
		return gtwEntry{}, err
	}
	return itf.([]gtwEntry)[0], nil // Storage guarantee at least one entry
}

// upsert implements the router.GtwStorage interface
func (s *gtwStorage) upsert(entry gtwEntry) error {
	return s.db.Update(entry.GatewayID, []encoding.BinaryMarshaler{entry}, dbGateways)
}

// done implements the router.GtwStorage interface
func (s *gtwStorage) done() error {
	return s.db.Close()
}

// MarshalBinary implements the encoding.BinaryMarshaler interface
func (e gtwEntry) MarshalBinary() ([]byte, error) {
	data, err := e.Metadata.MarshalBinary()
	if err != nil {
		return nil, errors.New(errors.Structural, err)
	}
	rw := readwriter.New(nil)
	rw.Write(e.GatewayID)
	rw.Write(data)
	return rw.Bytes()
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface
func (e *gtwEntry) UnmarshalBinary(data []byte) error {
	rw := readwriter.New(data)
	rw.Read(func(data []byte) {
		e.GatewayID = make([]byte, len(data))
		copy(e.GatewayID, data)
	})
	rw.TryRead(func(data []byte) error { return e.Metadata.UnmarshalBinary(data) })
	return rw.Err()
}

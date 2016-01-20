// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package components

import (
	"sync"

	"github.com/TheThingsNetwork/ttn/core"
	"github.com/brocaar/lorawan"
)

type handlerStorage interface {
	store(lorawan.EUI64, handlerEntry) error
	partition([]core.Packet) ([]handlerPartition, error)
}

type handlerPartition struct {
	AppEUI  lorawan.EUI64
	NwsKey  lorawan.AES128Key
	AppKey  lorawan.AES128Key
	DevAddr lorawan.DevAddr
	Packets []core.Packet
}

type handlerEntry struct {
	NwsKey  lorawan.AES128Key
	AppKey  lorawan.AES128Key
	DevAddr lorawan.DevAddr
}

type handlerDB struct {
	sync.RWMutex
	entries map[lorawan.EUI64]handlerEntry
}

func NewHandlerDB() handlerStorage {
	return &handlerDB{entries: make(map[lorawan.EUI64]handlerEntry)}
}

func (db *handlerDB) store(appEUI lorawan.EUI64, entry handlerEntry) error {
	return nil
}

func (db *handlerDB) partition(packets []core.Packet) ([]handlerPartition, error) {
	return nil, nil
}

// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package router

import (
	. "github.com/TheThingsNetwork/ttn/refactor"
	"github.com/brocaar/lorawan"
)

type Storage interface {
	Lookup(devEUI lorawan.EUI64) (entry, error)
	Store(reg Registration) error
	Close() error
}

type entry struct {
	Recipient Recipient
}

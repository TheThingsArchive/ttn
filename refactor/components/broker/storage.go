// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package broker

import (
	. "github.com/TheThingsNetwork/ttn/refactor"
	"github.com/brocaar/lorawan"
)

type Storage interface {
	Lookup(devEUI lorawan.EUI64) ([]entry, error)
	Store(reg BRegistration) error
	Close() error
}

type entry struct {
	Recipient Recipient
	AppEUI    lorawan.EUI64
	DevEUI    lorawan.EUI64
	NwkSKey   lorawan.AES128Key
}

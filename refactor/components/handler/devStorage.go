// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package handler

import (
	. "github.com/TheThingsNetwork/ttn/refactor"
	"github.com/brocaar/lorawan"
)

type devStorage interface {
	Lookup(appEUI lorawan.EUI64, devEUI lorawan.EUI64) (devEntry, error)
	Store(r HRegistration) error
}

type devEntry struct {
	AppSKey   lorawan.AES128Key
	NwkSKey   lorawan.AES128Key
	Recipient Recipient
}

type appEntry struct {
	AppKey lorawan.AES128Key
}

// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package udp

import (
	core "github.com/TheThingsNetwork/ttn/refactor"
	. "github.com/TheThingsNetwork/ttn/refactor/errors"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	"github.com/brocaar/lorawan"
)

// udpRegistration is a blank type which implements the core.Registration interface
type udpRegistration struct{}

// Recipient implements the core.Registration inteface
func (r udpRegistration) Recipient() core.Recipient {
	return core.Recipient{}
}

// AppEUI implements the core.Registration interface
func (r udpRegistration) AppEUI() (lorawan.EUI64, error) {
	return lorawan.EUI64{}, errors.New(ErrNotSupported, "NextRegistration not supported on udp adapter")
}

// DevEUI implements the core.Registration interface
func (r udpRegistration) DevEUI() (lorawan.EUI64, error) {
	return lorawan.EUI64{}, errors.New(ErrNotSupported, "NextRegistration not supported on udp adapter")
}

// AppSKey implements the core.Registration interface
func (r udpRegistration) AppSKey() (lorawan.AES128Key, error) {
	return lorawan.AES128Key{}, errors.New(ErrNotSupported, "NextRegistration not supported on udp adapter")
}

// NwkSKey implements the core.Registration interface
func (r udpRegistration) NwkSKey() (lorawan.AES128Key, error) {
	return lorawan.AES128Key{}, errors.New(ErrNotSupported, "NextRegistration not supported on udp adapter")
}

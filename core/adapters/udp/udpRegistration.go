// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package udp

import (
	"github.com/TheThingsNetwork/ttn/core"
	"github.com/brocaar/lorawan"
)

// udpRegistration is a blank type which implements the core.RRegistration interface
type udpRegistration struct{}

// Recipient implements the core.RRegistration inteface
func (r udpRegistration) Recipient() core.Recipient {
	return nil
}

// DevEUI implements the core.RRegistration interface
func (r udpRegistration) DevEUI() lorawan.EUI64 {
	return lorawan.EUI64{}
}

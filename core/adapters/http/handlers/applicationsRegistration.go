// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package handlers

import (
	"github.com/TheThingsNetwork/ttn/core"
	. "github.com/TheThingsNetwork/ttn/core/adapters/http"
	"github.com/brocaar/lorawan"
)

// applicationsRegistration implements the core.ARegistration interface
type applicationsRegistration struct {
	recipient Recipient
	appEUI    lorawan.EUI64
}

// Recipient implements the core.ARegistration interface
func (r applicationsRegistration) Recipient() core.Recipient {
	return r.recipient
}

// AppEUI implements the core.ARegistration interface
func (r applicationsRegistration) AppEUI() lorawan.EUI64 {
	return r.appEUI
}

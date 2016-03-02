// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package broker

import (
	"testing"

	. "github.com/TheThingsNetwork/ttn/core"
	. "github.com/TheThingsNetwork/ttn/core/mocks"
	//"github.com/brocaar/lorawan"
)

// ----- CHECK utilities
func CheckEntries(t *testing.T, want []entry, got []entry) {
	Check(t, want, got, "Entries")
}

func CheckRegistrations(t *testing.T, want Registration, got Registration) {
	Check(t, want, got, "Registrations")
}

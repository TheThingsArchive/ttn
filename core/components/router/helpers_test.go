// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package router

import (
	"testing"
	"time"

	. "github.com/TheThingsNetwork/ttn/core/mocks"
	. "github.com/TheThingsNetwork/ttn/utils/testing"
)

func CheckEntries(t *testing.T, want entry, got entry) {
	tmin := want.until.Add(-time.Second)
	tmax := want.until.Add(time.Second)
	if !tmin.Before(got.until) || !got.until.Before(tmax) {
		Ko(t, "Unexpected expiry time.\nWant: %s\nGot:  %s", want.until, got.until)
	}
	Check(t, want.Recipient, got.Recipient, "Recipients")
}

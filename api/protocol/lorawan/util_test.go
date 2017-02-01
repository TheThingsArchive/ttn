// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package lorawan

import (
	"testing"

	"github.com/TheThingsNetwork/ttn/core/types"
	. "github.com/smartystreets/assertions"
)

func TestSetValidateMIC(t *testing.T) {
	a := New(t)

	m := new(Message)
	m.InitUplink()

	key := types.NwkSKey([16]byte{})

	{
		err := m.SetMIC(key)
		a.So(err, ShouldBeNil)
	}

	{
		err := m.ValidateMIC(key)
		a.So(err, ShouldBeNil)
	}
}

func TestEncryptDecrypt(t *testing.T) {
	a := New(t)

	m := new(Message)
	mac := m.InitUplink()
	mac.FPort = 1

	payload := []byte{1, 2, 3, 4}
	mac.FrmPayload = payload

	key := types.AppSKey([16]byte{})

	{
		err := m.EncryptFRMPayload(key)
		a.So(err, ShouldBeNil)
	}

	{
		err := m.DecryptFRMPayload(key)
		a.So(err, ShouldBeNil)
		a.So(m.GetMacPayload().FrmPayload, ShouldResemble, payload)
	}
}

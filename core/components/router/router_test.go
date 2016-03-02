// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package router

import (
	"testing"
	//"time"

	. "github.com/TheThingsNetwork/ttn/core/mocks"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	. "github.com/TheThingsNetwork/ttn/utils/errors/checks"
	"github.com/TheThingsNetwork/ttn/utils/pointer"
	. "github.com/TheThingsNetwork/ttn/utils/testing"
)

func TestRegister(t *testing.T) {
	{
		Desc(t, "Register an entry")

		// Build
		an := NewMockAckNacker()
		store := newMockStorage()
		r := NewMockRegistration()

		// Operate
		router := New(store, GetLogger(t, "Router"))
		err := router.Register(r, an)

		// Check
		CheckErrors(t, nil, err)
		CheckAcks(t, true, an.InAck)
		CheckRegistrations(t, store.InStore, r)
	}

	// -------------------

	{
		Desc(t, "Register an entry | Store failed")

		// Build
		an := NewMockAckNacker()
		store := newMockStorage()
		store.Failures["Store"] = errors.New(errors.Structural, "Mock Error: Store Failed")
		r := NewMockRegistration()

		// Operate
		router := New(store, GetLogger(t, "Router"))
		err := router.Register(r, an)

		// Check
		CheckErrors(t, pointer.String(string(errors.Structural)), err)
		CheckAcks(t, false, an.InAck)
		CheckRegistrations(t, store.InStore, r)
	}
}

// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package broker

import (
	"testing"

	. "github.com/TheThingsNetwork/ttn/core"
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
		broker := New(store, GetLogger(t, "Router"))
		err := broker.Register(r, an)

		// Check
		CheckErrors(t, nil, err)
		CheckAcks(t, true, an.InAck)
		CheckRegistrations(t, r, store.InStore)
	}

	// -------------------

	{
		Desc(t, "Register an entry | store failed")

		// Build
		an := NewMockAckNacker()
		store := newMockStorage()
		store.Failures["Store"] = errors.New(errors.Structural, "Mock Error: Store Failed")
		r := NewMockRegistration()

		// Operate
		broker := New(store, GetLogger(t, "Router"))
		err := broker.Register(r, an)

		// Check
		CheckErrors(t, pointer.String(string(errors.Structural)), err)
		CheckAcks(t, false, an.InAck)
		CheckRegistrations(t, r, store.InStore)
	}

	// -------------------

	{
		Desc(t, "Register an entry | Wrong registration")

		// Build
		an := NewMockAckNacker()
		store := newMockStorage()
		r := NewMockRRegistration()

		// Operate
		broker := New(store, GetLogger(t, "Router"))
		err := broker.Register(r, an)

		// Check
		CheckErrors(t, pointer.String(string(errors.Structural)), err)
		CheckAcks(t, false, an.InAck)
		CheckRegistrations(t, nil, store.InStore)
	}
}

func TestHandleDown(t *testing.T) {
	{
		Desc(t, "Try Handle Down")

		// Build
		an := NewMockAckNacker()
		adapter := NewMockAdapter()
		store := newMockStorage()

		// Operate
		broker := New(store, GetLogger(t, "Broker"))
		err := broker.HandleDown([]byte{1, 2, 3}, an, adapter)

		// Check
		CheckErrors(t, pointer.String(string(errors.Implementation)), err)
		CheckAcks(t, false, an.InAck)
		CheckRegistrations(t, nil, store.InStore)
		CheckSent(t, nil, adapter.InSendPacket)
		CheckRecipients(t, nil, adapter.InSendRecipients)
	}
}

func TestHandleUp(t *testing.T) {

}

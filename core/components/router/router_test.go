// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package router

import (
	"testing"
	"time"

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
		r := NewMockRRegistration()

		// Operate
		router := New(store, GetLogger(t, "Router"))
		err := router.Register(r, an)

		// Check
		CheckErrors(t, nil, err)
		CheckAcks(t, true, an.InAck)
		CheckRegistrations(t, r, store.InStore)
	}

	// -------------------

	{
		Desc(t, "Register an entry, wrong registration type")

		// Build
		an := NewMockAckNacker()
		store := newMockStorage()
		r := NewMockARegistration()

		// Operate
		router := New(store, GetLogger(t, "Router"))
		err := router.Register(r, an)

		// Check
		CheckErrors(t, pointer.String(string(errors.Structural)), err)
		CheckAcks(t, false, an.InAck)
		CheckRegistrations(t, nil, store.InStore)
	}

	// -------------------

	{
		Desc(t, "Register an entry | Store failed")

		// Build
		an := NewMockAckNacker()
		store := newMockStorage()
		store.Failures["Store"] = errors.New(errors.Structural, "Mock Error: Store Failed")
		r := NewMockRRegistration()

		// Operate
		router := New(store, GetLogger(t, "Router"))
		err := router.Register(r, an)

		// Check
		CheckErrors(t, pointer.String(string(errors.Structural)), err)
		CheckAcks(t, false, an.InAck)
		CheckRegistrations(t, r, store.InStore)
	}
}

func TestHandleUp(t *testing.T) {
	{
		Desc(t, "Send an unknown packet | No downlink")

		// Build
		an := NewMockAckNacker()
		adapter := NewMockAdapter()
		adapter.OutSend = nil
		store := newMockStorage()
		store.Failures["Lookup"] = errors.New(errors.NotFound, "Mock Error: Not Found")
		data, err := newRPacket(
			[4]byte{2, 3, 2, 3},
			"Payload",
			[]byte{1, 2, 3, 4, 5, 6, 7, 8},
		).MarshalBinary()
		bpacket := newBPacket([4]byte{2, 3, 2, 3}, "Payload")

		// Operate
		router := New(store, GetLogger(t, "Router"))
		err = router.HandleUp(data, an, adapter)

		// Check
		CheckErrors(t, nil, err)
		CheckAcks(t, true, an.InAck)
		CheckRegistrations(t, nil, store.InStore)
		CheckSent(t, bpacket, adapter.InSendPacket)
		CheckRecipients(t, nil, adapter.InSendRecipients)
	}

	// -------------------

	{
		Desc(t, "Send an unknown packet | With Downlink")

		// Build
		an := NewMockAckNacker()
		resp := newRPacket(
			[4]byte{2, 3, 2, 3},
			"Response",
			[]byte{1, 2, 3, 4, 5, 6, 7, 8},
		)
		dataResp, _ := resp.MarshalBinary()
		adapter := NewMockAdapter()
		adapter.OutSend = dataResp
		store := newMockStorage()
		store.Failures["Lookup"] = errors.New(errors.NotFound, "Mock Error: Not Found")
		data, err := newRPacket(
			[4]byte{2, 3, 2, 3},
			"Payload",
			[]byte{1, 2, 3, 4, 5, 6, 7, 8},
		).MarshalBinary()
		bpacket := newBPacket([4]byte{2, 3, 2, 3}, "Payload")

		// Operate
		router := New(store, GetLogger(t, "Router"))
		err = router.HandleUp(data, an, adapter)

		// Check
		CheckErrors(t, nil, err)
		CheckAcks(t, resp, an.InAck)
		CheckRegistrations(t, nil, store.InStore)
		CheckSent(t, bpacket, adapter.InSendPacket)
		CheckRecipients(t, nil, adapter.InSendRecipients)
	}

	// -------------------

	{
		Desc(t, "Send invalid data")

		// Build
		an := NewMockAckNacker()
		adapter := NewMockAdapter()
		store := newMockStorage()
		store.Failures["Lookup"] = errors.New(errors.NotFound, "Mock Error: Not Found")

		// Operate
		router := New(store, GetLogger(t, "Router"))
		err := router.HandleUp([]byte{1, 2, 3}, an, adapter)

		// Check
		CheckErrors(t, pointer.String(string(errors.Structural)), err)
		CheckAcks(t, false, an.InAck)
		CheckRegistrations(t, nil, store.InStore)
		CheckSent(t, nil, adapter.InSendPacket)
		CheckRecipients(t, nil, adapter.InSendRecipients)
	}

	// -------------------

	{
		Desc(t, "Send an unknown packet | No downlink | Storage fail lookup ")

		// Build
		an := NewMockAckNacker()
		adapter := NewMockAdapter()
		adapter.OutSend = nil
		store := newMockStorage()
		store.Failures["Lookup"] = errors.New(errors.Operational, "Mock Error: Lookup failed")
		data, err := newRPacket(
			[4]byte{2, 3, 2, 3},
			"Payload",
			[]byte{1, 2, 3, 4, 5, 6, 7, 8},
		).MarshalBinary()

		// Operate
		router := New(store, GetLogger(t, "Router"))
		err = router.HandleUp(data, an, adapter)

		// Check
		CheckErrors(t, pointer.String(string(errors.Operational)), err)
		CheckAcks(t, false, an.InAck)
		CheckRegistrations(t, nil, store.InStore)
		CheckSent(t, nil, adapter.InSendPacket)
		CheckRecipients(t, nil, adapter.InSendRecipients)
	}

	// -------------------

	{
		Desc(t, "Send known packet | No downlink")

		// Build
		an := NewMockAckNacker()
		adapter := NewMockAdapter()
		adapter.OutSend = nil
		recipient := NewMockRecipient()
		dataRecipient, _ := recipient.MarshalBinary()
		store := newMockStorage()
		store.OutLookup = []entry{
			{
				Recipient: dataRecipient,
				until:     time.Now().Add(time.Hour),
			},
		}
		data, err := newRPacket(
			[4]byte{2, 3, 2, 3},
			"Payload",
			[]byte{1, 2, 3, 4, 5, 6, 7, 8},
		).MarshalBinary()
		bpacket := newBPacket([4]byte{2, 3, 2, 3}, "Payload")

		// Operate
		router := New(store, GetLogger(t, "Router"))
		err = router.HandleUp(data, an, adapter)

		// Check
		CheckErrors(t, nil, err)
		CheckAcks(t, true, an.InAck)
		CheckRegistrations(t, nil, store.InStore)
		CheckSent(t, bpacket, adapter.InSendPacket)
		CheckRecipients(t, []Recipient{recipient}, adapter.InSendRecipients)
	}

	// -------------------

	{
		Desc(t, "Send unknown packet | Get wrong recipient from db")

		// Build
		an := NewMockAckNacker()
		adapter := NewMockAdapter()
		adapter.Failures["GetRecipient"] = errors.New(errors.Structural, "Mock Error: Invalid recipient")
		recipient := NewMockRecipient()
		dataRecipient, _ := recipient.MarshalBinary()
		store := newMockStorage()
		store.OutLookup = []entry{
			{
				Recipient: dataRecipient,
				until:     time.Now().Add(time.Hour),
			},
		}
		data, err := newRPacket(
			[4]byte{2, 3, 2, 3},
			"Payload",
			[]byte{1, 2, 3, 4, 5, 6, 7, 8},
		).MarshalBinary()

		// Operate
		router := New(store, GetLogger(t, "Router"))
		err = router.HandleUp(data, an, adapter)

		// Check
		CheckErrors(t, pointer.String(string(errors.Structural)), err)
		CheckAcks(t, false, an.InAck)
		CheckRegistrations(t, nil, store.InStore)
		CheckSent(t, nil, adapter.InSendPacket)
		CheckRecipients(t, nil, adapter.InSendRecipients)
	}

	// -------------------

	{
		Desc(t, "Send unknown packet | Sending fails")

		// Build
		an := NewMockAckNacker()
		adapter := NewMockAdapter()
		adapter.Failures["Send"] = errors.New(errors.Operational, "Mock Error: Unable to send")
		store := newMockStorage()
		store.Failures["Lookup"] = errors.New(errors.NotFound, "Mock Error: Not found")
		data, err := newRPacket(
			[4]byte{2, 3, 2, 3},
			"Payload",
			[]byte{1, 2, 3, 4, 5, 6, 7, 8},
		).MarshalBinary()
		bpacket := newBPacket([4]byte{2, 3, 2, 3}, "Payload")

		// Operate
		router := New(store, GetLogger(t, "Router"))
		err = router.HandleUp(data, an, adapter)

		// Check
		CheckErrors(t, pointer.String(string(errors.Operational)), err)
		CheckAcks(t, false, an.InAck)
		CheckRegistrations(t, nil, store.InStore)
		CheckSent(t, bpacket, adapter.InSendPacket)
		CheckRecipients(t, nil, adapter.InSendRecipients)
	}

	// -------------------

	{
		Desc(t, "Send unknown packet | Get invalid downlink response")

		// Build
		an := NewMockAckNacker()
		adapter := NewMockAdapter()
		adapter.OutSend = []byte{1, 2, 3}
		store := newMockStorage()
		store.Failures["Lookup"] = errors.New(errors.NotFound, "Mock Error: Not found")
		data, err := newRPacket(
			[4]byte{2, 3, 2, 3},
			"Payload",
			[]byte{1, 2, 3, 4, 5, 6, 7, 8},
		).MarshalBinary()
		bpacket := newBPacket([4]byte{2, 3, 2, 3}, "Payload")

		// Operate
		router := New(store, GetLogger(t, "Router"))
		err = router.HandleUp(data, an, adapter)

		// Check
		CheckErrors(t, pointer.String(string(errors.Operational)), err)
		CheckAcks(t, false, an.InAck)
		CheckRegistrations(t, nil, store.InStore)
		CheckSent(t, bpacket, adapter.InSendPacket)
		CheckRecipients(t, nil, adapter.InSendRecipients)
	}
}

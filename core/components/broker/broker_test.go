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
	"github.com/brocaar/lorawan"
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
	{
		Desc(t, "Send an unknown packet")

		// Build
		an := NewMockAckNacker()
		adapter := NewMockAdapter()
		store := newMockStorage()
		store.Failures["Lookup"] = errors.New(errors.Behavioural, "Mock Error: Not Found")
		data, _ := newBPacket(
			[4]byte{2, 3, 2, 3},
			"Payload",
			[16]byte{1, 1, 2, 2, 3, 3, 4, 4, 5, 5, 6, 6, 7, 7, 8, 8},
			5,
		).MarshalBinary()

		// Operate
		broker := New(store, GetLogger(t, "Broker"))
		err := broker.HandleUp(data, an, adapter)

		// Check
		CheckErrors(t, pointer.String(string(errors.Behavioural)), err)
		CheckAcks(t, false, an.InAck)
		CheckRegistrations(t, nil, store.InStore)
		CheckSent(t, nil, adapter.InSendPacket)
		CheckRecipients(t, nil, adapter.InSendRecipients)
	}

	// -------------------

	{
		Desc(t, "Send an invalid packet")

		// Build
		an := NewMockAckNacker()
		adapter := NewMockAdapter()
		store := newMockStorage()

		// Operate
		broker := New(store, GetLogger(t, "Broker"))
		err := broker.HandleUp([]byte{1, 2, 3}, an, adapter)

		// Check
		CheckErrors(t, pointer.String(string(errors.Structural)), err)
		CheckAcks(t, false, an.InAck)
		CheckRegistrations(t, nil, store.InStore)
		CheckSent(t, nil, adapter.InSendPacket)
		CheckRecipients(t, nil, adapter.InSendRecipients)
	}

	// -------------------

	{
		Desc(t, "Send packet, get 2 entries, no valid MIC")

		// Build
		an := NewMockAckNacker()
		adapter := NewMockAdapter()
		store := newMockStorage()
		store.OutLookup = []entry{
			{
				Recipient: []byte{1, 2, 3},
				AppEUI:    lorawan.EUI64([8]byte{1, 2, 3, 4, 5, 6, 7, 8}),
				DevEUI:    lorawan.EUI64([8]byte{1, 2, 3, 4, 5, 6, 7, 8}),
				NwkSKey:   lorawan.AES128Key([16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}),
			},
			{
				Recipient: []byte{1, 2, 3},
				AppEUI:    lorawan.EUI64([8]byte{1, 1, 1, 1, 5, 5, 5, 5}),
				DevEUI:    lorawan.EUI64([8]byte{4, 4, 4, 4, 5, 5, 5, 5}),
				NwkSKey:   lorawan.AES128Key([16]byte{1, 2, 3, 4, 5, 6, 7, 8, 6, 6, 11, 12, 13, 14, 12, 16}),
			},
		}
		data, _ := newBPacket(
			[4]byte{2, 3, 2, 3},
			"Payload",
			[16]byte{1, 1, 2, 2, 3, 3, 4, 4, 5, 5, 6, 6, 7, 7, 8, 8},
			5,
		).MarshalBinary()

		// Operate
		broker := New(store, GetLogger(t, "Broker"))
		err := broker.HandleUp(data, an, adapter)

		// Check
		CheckErrors(t, pointer.String(string(errors.Behavioural)), err)
		CheckAcks(t, false, an.InAck)
		CheckRegistrations(t, nil, store.InStore)
		CheckSent(t, nil, adapter.InSendPacket)
		CheckRecipients(t, nil, adapter.InSendRecipients)
	}

	// -------------------

	{
		Desc(t, "Send packet, get 2 entries, 1 valid MIC | No downlink")

		// Build
		an := NewMockAckNacker()
		recipient := NewMockRecipient()
		adapter := NewMockAdapter()
		adapter.OutSend = nil
		adapter.OutGetRecipient = recipient
		store := newMockStorage()
		store.OutLookup = []entry{
			{
				Recipient: []byte{1, 2, 3},
				AppEUI:    lorawan.EUI64([8]byte{1, 2, 3, 4, 5, 6, 7, 8}),
				DevEUI:    lorawan.EUI64([8]byte{1, 2, 3, 4, 5, 6, 7, 8}),
				NwkSKey:   lorawan.AES128Key([16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}),
			},
			{
				Recipient: []byte{1, 2, 3},
				AppEUI:    lorawan.EUI64([8]byte{1, 1, 1, 1, 5, 5, 5, 5}),
				DevEUI:    lorawan.EUI64([8]byte{4, 4, 4, 4, 2, 3, 2, 3}),
				NwkSKey:   lorawan.AES128Key([16]byte{1, 1, 2, 2, 3, 3, 4, 4, 5, 5, 6, 6, 7, 7, 8, 8}),
			},
		}
		bpacket := newBPacket(
			[4]byte{2, 3, 2, 3},
			"Payload",
			[16]byte{1, 1, 2, 2, 3, 3, 4, 4, 5, 5, 6, 6, 7, 7, 8, 8},
			5,
		)
		data, _ := bpacket.MarshalBinary()
		hpacket, _ := NewHPacket(
			store.OutLookup[1].AppEUI,
			store.OutLookup[1].DevEUI,
			bpacket.Payload(),
			bpacket.Metadata(),
		)

		// Operate
		broker := New(store, GetLogger(t, "Broker"))
		err := broker.HandleUp(data, an, adapter)

		// Check
		CheckErrors(t, nil, err)
		CheckAcks(t, true, an.InAck)
		CheckRegistrations(t, nil, store.InStore)
		CheckSent(t, hpacket, adapter.InSendPacket)
		CheckRecipients(t, []Recipient{recipient}, adapter.InSendRecipients)
	}

	// -------------------

	{
		Desc(t, "Send packet, get 2 entries, 1 valid MIC | Fails to get recipient")

		// Build
		an := NewMockAckNacker()
		adapter := NewMockAdapter()
		adapter.OutSend = nil
		adapter.Failures["GetRecipient"] = errors.New(errors.Structural, "Mock Error: Unable to get recipient")
		store := newMockStorage()
		store.OutLookup = []entry{
			{
				Recipient: []byte{1, 2, 3},
				AppEUI:    lorawan.EUI64([8]byte{1, 2, 3, 4, 5, 6, 7, 8}),
				DevEUI:    lorawan.EUI64([8]byte{1, 2, 3, 4, 5, 6, 7, 8}),
				NwkSKey:   lorawan.AES128Key([16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}),
			},
			{
				Recipient: []byte{1, 2, 3},
				AppEUI:    lorawan.EUI64([8]byte{1, 1, 1, 1, 5, 5, 5, 5}),
				DevEUI:    lorawan.EUI64([8]byte{4, 4, 4, 4, 2, 3, 2, 3}),
				NwkSKey:   lorawan.AES128Key([16]byte{1, 1, 2, 2, 3, 3, 4, 4, 5, 5, 6, 6, 7, 7, 8, 8}),
			},
		}
		bpacket := newBPacket(
			[4]byte{2, 3, 2, 3},
			"Payload",
			[16]byte{1, 1, 2, 2, 3, 3, 4, 4, 5, 5, 6, 6, 7, 7, 8, 8},
			5,
		)
		data, _ := bpacket.MarshalBinary()

		// Operate
		broker := New(store, GetLogger(t, "Broker"))
		err := broker.HandleUp(data, an, adapter)

		// Check
		CheckErrors(t, pointer.String(string(errors.Structural)), err)
		CheckAcks(t, false, an.InAck)
		CheckRegistrations(t, nil, store.InStore)
		CheckSent(t, nil, adapter.InSendPacket)
		CheckRecipients(t, nil, adapter.InSendRecipients)
	}

	// -------------------

	{
		Desc(t, "Send packet, get 2 entries, 1 valid MIC | Fails to send")

		// Build
		an := NewMockAckNacker()
		recipient := NewMockRecipient()
		adapter := NewMockAdapter()
		adapter.OutGetRecipient = recipient
		adapter.Failures["Send"] = errors.New(errors.Operational, "Mock Error: Unable to send")
		store := newMockStorage()
		store.OutLookup = []entry{
			{
				Recipient: []byte{1, 2, 3},
				AppEUI:    lorawan.EUI64([8]byte{1, 2, 3, 4, 5, 6, 7, 8}),
				DevEUI:    lorawan.EUI64([8]byte{1, 2, 3, 4, 5, 6, 7, 8}),
				NwkSKey:   lorawan.AES128Key([16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}),
			},
			{
				Recipient: []byte{1, 2, 3},
				AppEUI:    lorawan.EUI64([8]byte{1, 1, 1, 1, 5, 5, 5, 5}),
				DevEUI:    lorawan.EUI64([8]byte{4, 4, 4, 4, 2, 3, 2, 3}),
				NwkSKey:   lorawan.AES128Key([16]byte{1, 1, 2, 2, 3, 3, 4, 4, 5, 5, 6, 6, 7, 7, 8, 8}),
			},
		}
		bpacket := newBPacket(
			[4]byte{2, 3, 2, 3},
			"Payload",
			[16]byte{1, 1, 2, 2, 3, 3, 4, 4, 5, 5, 6, 6, 7, 7, 8, 8},
			5,
		)
		data, _ := bpacket.MarshalBinary()
		hpacket, _ := NewHPacket(
			store.OutLookup[1].AppEUI,
			store.OutLookup[1].DevEUI,
			bpacket.Payload(),
			bpacket.Metadata(),
		)

		// Operate
		broker := New(store, GetLogger(t, "Broker"))
		err := broker.HandleUp(data, an, adapter)

		// Check
		CheckErrors(t, pointer.String(string(errors.Operational)), err)
		CheckAcks(t, false, an.InAck)
		CheckRegistrations(t, nil, store.InStore)
		CheckSent(t, hpacket, adapter.InSendPacket)
		CheckRecipients(t, []Recipient{recipient}, adapter.InSendRecipients)
	}
}

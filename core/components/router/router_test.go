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
		m := newMockDutyManager()

		// Operate
		router := New(store, m, GetLogger(t, "Router"))
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
		m := newMockDutyManager()

		// Operate
		router := New(store, m, GetLogger(t, "Router"))
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
		m := newMockDutyManager()

		// Operate
		router := New(store, m, GetLogger(t, "Router"))
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
		inPacket := newRPacket(
			[4]byte{2, 3, 2, 3},
			"Payload",
			[]byte{1, 2, 3, 4, 5, 6, 7, 8},
			Metadata{Freq: pointer.Float64(865.5)},
		)
		data, _ := inPacket.MarshalBinary()
		bpacket := newBPacket(
			[4]byte{2, 3, 2, 3},
			"Payload",
			Metadata{Freq: pointer.Float64(865.5), DutyRX1: pointer.Uint(0), DutyRX2: pointer.Uint(0)},
		)
		m := newMockDutyManager()

		// Operate
		router := New(store, m, GetLogger(t, "Router"))
		err := router.HandleUp(data, an, adapter)

		// Check
		CheckErrors(t, nil, err)
		CheckAcks(t, true, an.InAck)
		CheckRegistrations(t, nil, store.InStore)
		CheckSent(t, bpacket, adapter.InSendPacket)
		CheckRecipients(t, nil, adapter.InSendRecipients)
		CheckIDs(t, inPacket.GatewayID(), m.InLookupId)
		CheckIDs(t, nil, m.InUpdateId)
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
			Metadata{
				Freq: pointer.Float64(868.42),
				Size: pointer.Uint(14),
				Codr: pointer.String("4/5"),
				Datr: pointer.String("SF8BW125"),
			},
		)
		dataResp, _ := resp.MarshalBinary()
		adapter := NewMockAdapter()
		adapter.OutSend = dataResp
		store := newMockStorage()
		store.Failures["Lookup"] = errors.New(errors.NotFound, "Mock Error: Not Found")
		inPacket := newRPacket(
			[4]byte{2, 3, 2, 3},
			"Payload",
			[]byte{1, 2, 3, 4, 5, 6, 7, 8},
			Metadata{Freq: pointer.Float64(865.5)},
		)
		data, _ := inPacket.MarshalBinary()
		bpacket := newBPacket(
			[4]byte{2, 3, 2, 3},
			"Payload",
			Metadata{Freq: pointer.Float64(865.5), DutyRX1: pointer.Uint(0), DutyRX2: pointer.Uint(0)},
		)
		m := newMockDutyManager()

		// Operate
		router := New(store, m, GetLogger(t, "Router"))
		err := router.HandleUp(data, an, adapter)

		// Check
		CheckErrors(t, nil, err)
		CheckAcks(t, resp, an.InAck)
		CheckRegistrations(t, nil, store.InStore)
		CheckSent(t, bpacket, adapter.InSendPacket)
		CheckRecipients(t, nil, adapter.InSendRecipients)
		CheckIDs(t, inPacket.GatewayID(), m.InLookupId)
		CheckIDs(t, inPacket.GatewayID(), m.InUpdateId)
	}

	// -------------------

	{
		Desc(t, "Send invalid data")

		// Build
		an := NewMockAckNacker()
		adapter := NewMockAdapter()
		store := newMockStorage()
		store.Failures["Lookup"] = errors.New(errors.NotFound, "Mock Error: Not Found")
		m := newMockDutyManager()

		// Operate
		router := New(store, m, GetLogger(t, "Router"))
		err := router.HandleUp([]byte{1, 2, 3}, an, adapter)

		// Check
		CheckErrors(t, pointer.String(string(errors.Structural)), err)
		CheckAcks(t, false, an.InAck)
		CheckRegistrations(t, nil, store.InStore)
		CheckSent(t, nil, adapter.InSendPacket)
		CheckRecipients(t, nil, adapter.InSendRecipients)
		CheckIDs(t, nil, m.InUpdateId)
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
		inPacket := newRPacket(
			[4]byte{2, 3, 2, 3},
			"Payload",
			[]byte{1, 2, 3, 4, 5, 6, 7, 8},
			Metadata{Freq: pointer.Float64(865.5)},
		)
		data, _ := inPacket.MarshalBinary()
		m := newMockDutyManager()

		// Operate
		router := New(store, m, GetLogger(t, "Router"))
		err := router.HandleUp(data, an, adapter)

		// Check
		CheckErrors(t, pointer.String(string(errors.Operational)), err)
		CheckAcks(t, false, an.InAck)
		CheckRegistrations(t, nil, store.InStore)
		CheckSent(t, nil, adapter.InSendPacket)
		CheckRecipients(t, nil, adapter.InSendRecipients)
		CheckIDs(t, nil, m.InUpdateId)
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
		inPacket := newRPacket(
			[4]byte{2, 3, 2, 3},
			"Payload",
			[]byte{1, 2, 3, 4, 5, 6, 7, 8},
			Metadata{Freq: pointer.Float64(865.5)},
		)
		data, _ := inPacket.MarshalBinary()
		bpacket := newBPacket(
			[4]byte{2, 3, 2, 3},
			"Payload",
			Metadata{Freq: pointer.Float64(865.5), DutyRX1: pointer.Uint(0), DutyRX2: pointer.Uint(0)},
		)
		m := newMockDutyManager()

		// Operate
		router := New(store, m, GetLogger(t, "Router"))
		err := router.HandleUp(data, an, adapter)

		// Check
		CheckErrors(t, nil, err)
		CheckAcks(t, true, an.InAck)
		CheckRegistrations(t, nil, store.InStore)
		CheckSent(t, bpacket, adapter.InSendPacket)
		CheckRecipients(t, []Recipient{recipient}, adapter.InSendRecipients)
		CheckIDs(t, inPacket.GatewayID(), m.InLookupId)
		CheckIDs(t, nil, m.InUpdateId)
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
		inPacket := newRPacket(
			[4]byte{2, 3, 2, 3},
			"Payload",
			[]byte{1, 2, 3, 4, 5, 6, 7, 8},
			Metadata{Freq: pointer.Float64(865.5)},
		)
		data, _ := inPacket.MarshalBinary()
		m := newMockDutyManager()

		// Operate
		router := New(store, m, GetLogger(t, "Router"))
		err := router.HandleUp(data, an, adapter)

		// Check
		CheckErrors(t, pointer.String(string(errors.Structural)), err)
		CheckAcks(t, false, an.InAck)
		CheckRegistrations(t, nil, store.InStore)
		CheckSent(t, nil, adapter.InSendPacket)
		CheckRecipients(t, nil, adapter.InSendRecipients)
		CheckIDs(t, inPacket.GatewayID(), m.InLookupId)
		CheckIDs(t, nil, m.InUpdateId)
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
		inPacket := newRPacket(
			[4]byte{2, 3, 2, 3},
			"Payload",
			[]byte{1, 2, 3, 4, 5, 6, 7, 8},
			Metadata{Freq: pointer.Float64(865.5)},
		)
		data, _ := inPacket.MarshalBinary()
		bpacket := newBPacket(
			[4]byte{2, 3, 2, 3},
			"Payload",
			Metadata{Freq: pointer.Float64(865.5), DutyRX1: pointer.Uint(0), DutyRX2: pointer.Uint(0)},
		)
		m := newMockDutyManager()

		// Operate
		router := New(store, m, GetLogger(t, "Router"))
		err := router.HandleUp(data, an, adapter)

		// Check
		CheckErrors(t, pointer.String(string(errors.Operational)), err)
		CheckAcks(t, false, an.InAck)
		CheckRegistrations(t, nil, store.InStore)
		CheckSent(t, bpacket, adapter.InSendPacket)
		CheckRecipients(t, nil, adapter.InSendRecipients)
		CheckIDs(t, inPacket.GatewayID(), m.InLookupId)
		CheckIDs(t, nil, m.InUpdateId)
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
		inPacket := newRPacket(
			[4]byte{2, 3, 2, 3},
			"Payload",
			[]byte{1, 2, 3, 4, 5, 6, 7, 8},
			Metadata{Freq: pointer.Float64(865.5)},
		)
		data, _ := inPacket.MarshalBinary()
		bpacket := newBPacket(
			[4]byte{2, 3, 2, 3},
			"Payload",
			Metadata{Freq: pointer.Float64(865.5), DutyRX1: pointer.Uint(0), DutyRX2: pointer.Uint(0)},
		)
		m := newMockDutyManager()

		// Operate
		router := New(store, m, GetLogger(t, "Router"))
		err := router.HandleUp(data, an, adapter)

		// Check
		CheckErrors(t, pointer.String(string(errors.Operational)), err)
		CheckAcks(t, false, an.InAck)
		CheckRegistrations(t, nil, store.InStore)
		CheckSent(t, bpacket, adapter.InSendPacket)
		CheckRecipients(t, nil, adapter.InSendRecipients)
		CheckIDs(t, inPacket.GatewayID(), m.InLookupId)
		CheckIDs(t, nil, m.InUpdateId)
	}

	// -------------------

	{
		Desc(t, "Send a known packet, get not found, and broadcast")

		// Build
		an := NewMockAckNacker()
		adapter := newMockRouterAdapter()
		adapter.OutSend = nil
		adapter.Failures["Send"] = []error{
			errors.New(errors.NotFound, "Mock Error"),
		}
		recipient := NewMockRecipient()
		dataRecipient, _ := recipient.MarshalBinary()
		store := newMockStorage()
		store.OutLookup = []entry{
			{
				Recipient: dataRecipient,
				until:     time.Now().Add(time.Hour),
			},
		}
		inPacket := newRPacket(
			[4]byte{2, 3, 2, 3},
			"Payload",
			[]byte{1, 2, 3, 4, 5, 6, 7, 8},
			Metadata{Freq: pointer.Float64(865.5)},
		)
		data, _ := inPacket.MarshalBinary()
		bpacket := newBPacket(
			[4]byte{2, 3, 2, 3},
			"Payload",
			Metadata{Freq: pointer.Float64(865.5), DutyRX1: pointer.Uint(0), DutyRX2: pointer.Uint(0)},
		)
		m := newMockDutyManager()

		// Operate
		router := New(store, m, GetLogger(t, "Router"))
		err := router.HandleUp(data, an, adapter)

		// Check
		CheckErrors(t, nil, err)
		CheckAcks(t, true, an.InAck)
		CheckRegistrations(t, nil, store.InStore)
		CheckSent(t, bpacket, adapter.InSendPacket)
		CheckRecipients(t, nil, adapter.InSendRecipients)
	}

	// -------------------

	{
		Desc(t, "Send a known packet, get not found, and broadcast, still not found")

		// Build
		an := NewMockAckNacker()
		adapter := newMockRouterAdapter()
		adapter.OutSend = nil
		adapter.Failures["Send"] = []error{
			errors.New(errors.NotFound, "Mock Error"),
			errors.New(errors.NotFound, "Mock Error"),
		}
		recipient := NewMockRecipient()
		dataRecipient, _ := recipient.MarshalBinary()
		store := newMockStorage()
		store.OutLookup = []entry{
			{
				Recipient: dataRecipient,
				until:     time.Now().Add(time.Hour),
			},
		}
		inPacket := newRPacket(
			[4]byte{2, 3, 2, 3},
			"Payload",
			[]byte{1, 2, 3, 4, 5, 6, 7, 8},
			Metadata{Freq: pointer.Float64(865.5)},
		)
		data, _ := inPacket.MarshalBinary()
		bpacket := newBPacket(
			[4]byte{2, 3, 2, 3},
			"Payload",
			Metadata{Freq: pointer.Float64(865.5), DutyRX1: pointer.Uint(0), DutyRX2: pointer.Uint(0)},
		)
		m := newMockDutyManager()

		// Operate
		router := New(store, m, GetLogger(t, "Router"))
		err := router.HandleUp(data, an, adapter)

		// Check
		CheckErrors(t, pointer.String(string(errors.NotFound)), err)
		CheckAcks(t, false, an.InAck)
		CheckRegistrations(t, nil, store.InStore)
		CheckSent(t, bpacket, adapter.InSendPacket)
		CheckRecipients(t, nil, adapter.InSendRecipients)
		CheckIDs(t, inPacket.GatewayID(), m.InLookupId)
		CheckIDs(t, nil, m.InUpdateId)
	}

	// -------------------

	{
		Desc(t, "Send an unknown packet | No Downlink | Fail to lookup gateway")

		// Build
		an := NewMockAckNacker()
		adapter := NewMockAdapter()
		adapter.OutSend = nil
		store := newMockStorage()
		store.Failures["Lookup"] = errors.New(errors.NotFound, "Mock Error: Not Found")
		inPacket := newRPacket(
			[4]byte{2, 3, 2, 3},
			"Payload",
			[]byte{1, 2, 3, 4, 5, 6, 7, 8},
			Metadata{Freq: pointer.Float64(865.5)},
		)
		data, _ := inPacket.MarshalBinary()
		bpacket := newBPacket(
			[4]byte{2, 3, 2, 3},
			"Payload",
			Metadata{Freq: pointer.Float64(865.5), DutyRX1: pointer.Uint(0), DutyRX2: pointer.Uint(0)},
		)
		m := newMockDutyManager()
		m.Failures["Lookup"] = errors.New(errors.NotFound, "Mock Error")

		// Operate
		router := New(store, m, GetLogger(t, "Router"))
		err := router.HandleUp(data, an, adapter)

		// Check
		CheckErrors(t, nil, err)
		CheckAcks(t, true, an.InAck)
		CheckRegistrations(t, nil, store.InStore)
		CheckSent(t, bpacket, adapter.InSendPacket)
		CheckRecipients(t, nil, adapter.InSendRecipients)
		CheckIDs(t, inPacket.GatewayID(), m.InLookupId)
		CheckIDs(t, nil, m.InUpdateId)
	}

	// -------------------

	{
		Desc(t, "Send an unknown packet | Missing Metadata in downlink")

		// Build
		an := NewMockAckNacker()
		resp := newRPacket(
			[4]byte{2, 3, 2, 3},
			"Response",
			[]byte{1, 2, 3, 4, 5, 6, 7, 8},
			Metadata{
				Size: pointer.Uint(14),
				Codr: pointer.String("4/5"),
				Datr: pointer.String("SF8BW125"),
			},
		)
		dataResp, _ := resp.MarshalBinary()
		adapter := NewMockAdapter()
		adapter.OutSend = dataResp
		store := newMockStorage()
		store.Failures["Lookup"] = errors.New(errors.NotFound, "Mock Error: Not Found")
		inPacket := newRPacket(
			[4]byte{2, 3, 2, 3},
			"Payload",
			[]byte{1, 2, 3, 4, 5, 6, 7, 8},
			Metadata{Freq: pointer.Float64(865.5)},
		)
		data, _ := inPacket.MarshalBinary()
		bpacket := newBPacket(
			[4]byte{2, 3, 2, 3},
			"Payload",
			Metadata{Freq: pointer.Float64(865.5), DutyRX1: pointer.Uint(0), DutyRX2: pointer.Uint(0)},
		)
		m := newMockDutyManager()

		// Operate
		router := New(store, m, GetLogger(t, "Router"))
		err := router.HandleUp(data, an, adapter)

		// Check
		CheckErrors(t, pointer.String(string(errors.Operational)), err)
		CheckAcks(t, false, an.InAck)
		CheckRegistrations(t, nil, store.InStore)
		CheckSent(t, bpacket, adapter.InSendPacket)
		CheckRecipients(t, nil, adapter.InSendRecipients)
		CheckIDs(t, inPacket.GatewayID(), m.InLookupId)
		CheckIDs(t, nil, m.InUpdateId)
	}

	// -------------------

	{
		Desc(t, "Send an unknown packet | Fail to update metadata")

		// Build
		an := NewMockAckNacker()
		resp := newRPacket(
			[4]byte{2, 3, 2, 3},
			"Response",
			[]byte{1, 2, 3, 4, 5, 6, 7, 8},
			Metadata{
				Freq: pointer.Float64(868.14),
				Size: pointer.Uint(14),
				Codr: pointer.String("4/5"),
				Datr: pointer.String("SF8BW125"),
			},
		)
		dataResp, _ := resp.MarshalBinary()
		adapter := NewMockAdapter()
		adapter.OutSend = dataResp
		store := newMockStorage()
		store.Failures["Lookup"] = errors.New(errors.NotFound, "Mock Error: Not Found")
		inPacket := newRPacket(
			[4]byte{2, 3, 2, 3},
			"Payload",
			[]byte{1, 2, 3, 4, 5, 6, 7, 8},
			Metadata{Freq: pointer.Float64(865.5)},
		)
		data, _ := inPacket.MarshalBinary()
		bpacket := newBPacket(
			[4]byte{2, 3, 2, 3},
			"Payload",
			Metadata{Freq: pointer.Float64(865.5), DutyRX1: pointer.Uint(0), DutyRX2: pointer.Uint(0)},
		)
		m := newMockDutyManager()
		m.Failures["Update"] = errors.New(errors.Operational, "Mock Error: Update")

		// Operate
		router := New(store, m, GetLogger(t, "Router"))
		err := router.HandleUp(data, an, adapter)

		// Check
		CheckErrors(t, pointer.String(string(errors.Operational)), err)
		CheckAcks(t, false, an.InAck)
		CheckRegistrations(t, nil, store.InStore)
		CheckSent(t, bpacket, adapter.InSendPacket)
		CheckRecipients(t, nil, adapter.InSendRecipients)
		CheckIDs(t, inPacket.GatewayID(), m.InLookupId)
		CheckIDs(t, inPacket.GatewayID(), m.InUpdateId)
	}

	// -------------------

	{
		Desc(t, "Send an unknown packet | No Metadata")

		// Build
		an := NewMockAckNacker()
		adapter := NewMockAdapter()
		store := newMockStorage()
		store.Failures["Lookup"] = errors.New(errors.NotFound, "Mock Error: Not Found")
		inPacket := newRPacket(
			[4]byte{2, 3, 2, 3},
			"Payload",
			[]byte{1, 2, 3, 4, 5, 6, 7, 8},
			Metadata{},
		)
		data, _ := inPacket.MarshalBinary()
		m := newMockDutyManager()

		// Operate
		router := New(store, m, GetLogger(t, "Router"))
		err := router.HandleUp(data, an, adapter)

		// Check
		CheckErrors(t, pointer.String(string(errors.Structural)), err)
		CheckAcks(t, false, an.InAck)
		CheckRegistrations(t, nil, store.InStore)
		CheckSent(t, nil, adapter.InSendPacket)
		CheckRecipients(t, nil, adapter.InSendRecipients)
		CheckIDs(t, nil, m.InLookupId)
		CheckIDs(t, nil, m.InUpdateId)
	}

	// -------------------

	{
		Desc(t, "Send an unknown packet | Unsupported frequency")

		// Build
		an := NewMockAckNacker()
		adapter := NewMockAdapter()
		store := newMockStorage()
		store.Failures["Lookup"] = errors.New(errors.NotFound, "Mock Error: Not Found")
		inPacket := newRPacket(
			[4]byte{2, 3, 2, 3},
			"Payload",
			[]byte{1, 2, 3, 4, 5, 6, 7, 8},
			Metadata{Freq: pointer.Float64(333.5)},
		)
		data, _ := inPacket.MarshalBinary()
		m := newMockDutyManager()

		// Operate
		router := New(store, m, GetLogger(t, "Router"))
		err := router.HandleUp(data, an, adapter)

		// Check
		CheckErrors(t, pointer.String(string(errors.Structural)), err)
		CheckAcks(t, false, an.InAck)
		CheckRegistrations(t, nil, store.InStore)
		CheckSent(t, nil, adapter.InSendPacket)
		CheckRecipients(t, nil, adapter.InSendRecipients)
		CheckIDs(t, inPacket.GatewayID(), m.InLookupId)
		CheckIDs(t, nil, m.InUpdateId)
	}

	// -------------------

	{
		Desc(t, "Send a valid stat packet")

		// Build
		an := NewMockAckNacker()
		adapter := NewMockAdapter()
		store := newMockStorage()
		inPacket, _ := NewSPacket(
			[]byte{1, 2, 3, 4, 5, 6, 7, 8},
			Metadata{Alti: pointer.Int(14)},
		)
		data, _ := inPacket.MarshalBinary()
		m := newMockDutyManager()

		// Operate
		router := New(store, m, GetLogger(t, "Router"))
		err := router.HandleUp(data, an, adapter)

		// Check
		CheckErrors(t, nil, err)
		CheckAcks(t, true, an.InAck)
		CheckRegistrations(t, nil, store.InStore)
		CheckStats(t, inPacket, store.InUpdateStats)
		CheckSent(t, nil, adapter.InSendPacket)
		CheckRecipients(t, nil, adapter.InSendRecipients)
		CheckIDs(t, nil, m.InLookupId)
		CheckIDs(t, nil, m.InUpdateId)
		CheckIDs(t, nil, store.InLookupStats)
	}

	// -------------------

	{
		Desc(t, "Send a valid stat packet | unable to update")

		// Build
		an := NewMockAckNacker()
		adapter := NewMockAdapter()
		store := newMockStorage()
		store.Failures["UpdateStats"] = errors.New(errors.Operational, "Mock Error")
		inPacket, _ := NewSPacket(
			[]byte{1, 2, 3, 4, 5, 6, 7, 8},
			Metadata{Alti: pointer.Int(14)},
		)
		data, _ := inPacket.MarshalBinary()
		m := newMockDutyManager()

		// Operate
		router := New(store, m, GetLogger(t, "Router"))
		err := router.HandleUp(data, an, adapter)

		// Check
		CheckErrors(t, pointer.String(string(errors.Operational)), err)
		CheckAcks(t, false, an.InAck)
		CheckRegistrations(t, nil, store.InStore)
		CheckStats(t, inPacket, store.InUpdateStats)
		CheckSent(t, nil, adapter.InSendPacket)
		CheckRecipients(t, nil, adapter.InSendRecipients)
		CheckIDs(t, nil, m.InLookupId)
		CheckIDs(t, nil, m.InUpdateId)
		CheckIDs(t, nil, store.InLookupStats)
	}
}

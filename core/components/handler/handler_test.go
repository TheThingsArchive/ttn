// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package handler

import (
	"testing"
	"time"

	. "github.com/TheThingsNetwork/ttn/core"
	"github.com/TheThingsNetwork/ttn/core/dutycycle"
	. "github.com/TheThingsNetwork/ttn/core/mocks"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	. "github.com/TheThingsNetwork/ttn/utils/errors/checks"
	"github.com/TheThingsNetwork/ttn/utils/pointer"
	. "github.com/TheThingsNetwork/ttn/utils/testing"
	"github.com/brocaar/lorawan"
)

func TestRegister(t *testing.T) {
	{
		Desc(t, "Register valid HRegistration")

		// Build
		devStorage := newMockDevStorage()
		pktStorage := newMockPktStorage()
		an := NewMockAckNacker()
		r := NewMockHRegistration()
		broker := NewMockJSONRecipient()
		br := NewMockBRegistration()
		br.OutRecipient = broker
		br.OutDevEUI = r.DevEUI()
		br.OutAppEUI = r.AppEUI()
		br.OutNwkSKey = r.NwkSKey()
		sub := NewMockSubscriber()

		// Operate
		handler := New(devStorage, pktStorage, broker, GetLogger(t, "Handler"))
		err := handler.Register(r, an, sub)

		// Check
		CheckErrors(t, nil, err)
		CheckPushed(t, nil, pktStorage.InPush)
		CheckPersonalized(t, r, devStorage.InStorePersonalized)
		CheckSubscriptions(t, br, sub.InSubscribeRegistration)
	}

	// --------------------

	{
		Desc(t, "Register invalid HRegistration")

		// Build
		devStorage := newMockDevStorage()
		pktStorage := newMockPktStorage()
		an := NewMockAckNacker()
		broker := NewMockJSONRecipient()
		sub := NewMockSubscriber()

		// Operate
		handler := New(devStorage, pktStorage, broker, GetLogger(t, "Handler"))
		err := handler.Register(nil, an, sub)

		// Checks
		CheckErrors(t, pointer.String(string(errors.Structural)), err)
		CheckPushed(t, nil, pktStorage.InPush)
		CheckPersonalized(t, nil, devStorage.InStorePersonalized)
		CheckSubscriptions(t, nil, sub.InSubscribeRegistration)
	}

	// --------------------

	{
		Desc(t, "Register valid HRegistration | devStorage fails")

		// Build
		devStorage := newMockDevStorage()
		devStorage.Failures["StorePersonalized"] = errors.New(errors.Operational, "Mock Error")
		pktStorage := newMockPktStorage()
		an := NewMockAckNacker()
		r := NewMockHRegistration()
		broker := NewMockJSONRecipient()
		sub := NewMockSubscriber()

		// Operate
		handler := New(devStorage, pktStorage, broker, GetLogger(t, "Handler"))
		err := handler.Register(r, an, sub)

		// Check
		CheckErrors(t, pointer.String(string(errors.Operational)), err)
		CheckPushed(t, nil, pktStorage.InPush)
		CheckPersonalized(t, r, devStorage.InStorePersonalized)
	}
}

func TestHandleDown(t *testing.T) {
	{
		Desc(t, "Handle downlink APacket")

		// Build
		devStorage := newMockDevStorage()
		pktStorage := newMockPktStorage()
		an := NewMockAckNacker()
		adapter := NewMockAdapter()
		pkt, _ := NewAPacket(
			[8]byte{1, 1, 1, 1, 1, 1, 1, 1},
			[8]byte{2, 2, 2, 2, 2, 2, 2, 2},
			[]byte("TheThingsNetwork"),
			[]Metadata{},
		)
		data, _ := pkt.MarshalBinary()
		broker := NewMockJSONRecipient()

		// Operate
		handler := New(devStorage, pktStorage, broker, GetLogger(t, "Handler"))
		err := handler.HandleDown(data, an, adapter)

		// Check
		CheckErrors(t, nil, err)
		CheckPushed(t, pkt, pktStorage.InPush)
		CheckPersonalized(t, nil, devStorage.InStorePersonalized)
		CheckAcks(t, true, an.InAck)
		CheckSent(t, nil, adapter.InSendPacket)
		CheckRecipients(t, nil, adapter.InSendRecipients)
	}

	// --------------------

	{
		Desc(t, "Handle downlink wrong data")

		// Build
		devStorage := newMockDevStorage()
		pktStorage := newMockPktStorage()
		an := NewMockAckNacker()
		adapter := NewMockAdapter()
		broker := NewMockJSONRecipient()

		// Operate
		handler := New(devStorage, pktStorage, broker, GetLogger(t, "Handler"))
		err := handler.HandleDown([]byte{1, 2, 3}, an, adapter)

		// Check
		CheckErrors(t, pointer.String(string(errors.Structural)), err)
		CheckPushed(t, nil, pktStorage.InPush)
		CheckPersonalized(t, nil, devStorage.InStorePersonalized)
		CheckAcks(t, false, an.InAck)
		CheckSent(t, nil, adapter.InSendPacket)
		CheckRecipients(t, nil, adapter.InSendRecipients)
	}

	// --------------------

	{
		Desc(t, "Handle downlink wrong packet type")

		// Build
		devStorage := newMockDevStorage()
		pktStorage := newMockPktStorage()
		an := NewMockAckNacker()
		adapter := NewMockAdapter()
		pkt := NewJPacket(
			lorawan.EUI64([8]byte{1, 1, 1, 1, 1, 1, 1, 1}),
			lorawan.EUI64([8]byte{2, 2, 2, 2, 2, 2, 2, 2}),
			[2]byte{14, 42},
			Metadata{},
		)
		data, _ := pkt.MarshalBinary()
		broker := NewMockJSONRecipient()

		// Operate
		handler := New(devStorage, pktStorage, broker, GetLogger(t, "Handler"))
		err := handler.HandleDown(data, an, adapter)

		// Check
		CheckErrors(t, pointer.String(string(errors.Implementation)), err)
		CheckPushed(t, nil, pktStorage.InPush)
		CheckPersonalized(t, nil, devStorage.InStorePersonalized)
		CheckAcks(t, false, an.InAck)
		CheckSent(t, nil, adapter.InSendPacket)
		CheckRecipients(t, nil, adapter.InSendRecipients)
	}
}

func TestHandleUp(t *testing.T) {
	{
		Desc(t, "Handle uplink with 1 packet | No Associated App")

		// Build
		devStorage := newMockDevStorage()
		devStorage.Failures["Lookup"] = errors.New(errors.NotFound, "Mock: Not Found")
		pktStorage := newMockPktStorage()
		pktStorage.Failures["Pull"] = errors.New(errors.NotFound, "Mock: Not Found")
		an := NewMockAckNacker()
		adapter := NewMockAdapter()
		inPkt := newHPacket(
			[8]byte{1, 1, 1, 1, 1, 1, 1, 1},
			[8]byte{2, 2, 2, 2, 2, 2, 2, 2},
			"Payload",
			Metadata{
				Datr: pointer.String("SF7BW125"),
			},
			10,
			[16]byte{1, 1, 1, 1, 2, 2, 2, 2, 1, 1, 1, 1, 2, 2, 2, 2},
		)
		dataIn, _ := inPkt.MarshalBinary()
		broker := NewMockJSONRecipient()

		// Operate
		handler := New(devStorage, pktStorage, broker, GetLogger(t, "Handler"))
		err := handler.HandleUp(dataIn, an, adapter)

		// Check
		CheckErrors(t, pointer.String(string(errors.NotFound)), err)
		CheckPushed(t, nil, pktStorage.InPush)
		CheckPersonalized(t, nil, devStorage.InStorePersonalized)
		CheckAcks(t, false, an.InAck)
		CheckSent(t, nil, adapter.InSendPacket)
		CheckRecipients(t, nil, adapter.InSendRecipients)
	}

	{
		Desc(t, "Handle uplink with invalid data")

		// Build
		devStorage := newMockDevStorage()
		pktStorage := newMockPktStorage()
		an := NewMockAckNacker()
		adapter := NewMockAdapter()
		broker := NewMockJSONRecipient()

		// Operate
		handler := New(devStorage, pktStorage, broker, GetLogger(t, "Handler"))
		err := handler.HandleUp([]byte{1, 2, 3}, an, adapter)

		// Check
		CheckErrors(t, pointer.String(string(errors.Structural)), err)
		CheckPushed(t, nil, pktStorage.InPush)
		CheckPersonalized(t, nil, devStorage.InStorePersonalized)
		CheckAcks(t, false, an.InAck)
		CheckSent(t, nil, adapter.InSendPacket)
		CheckRecipients(t, nil, adapter.InSendRecipients)
	}

	// --------------------

	{
		Desc(t, "Handle uplink with 1 packet | No downlink ready")

		// Build
		an := NewMockAckNacker()
		adapter := NewMockAdapter()
		dataRecipient, _ := adapter.OutGetRecipient.MarshalBinary()
		inPkt := newHPacket(
			[8]byte{1, 1, 1, 1, 1, 1, 1, 1},
			[8]byte{2, 2, 2, 2, 2, 2, 2, 2},
			"Payload",
			Metadata{
				Datr: pointer.String("SF7BW125"),
			},
			10,
			[16]byte{1, 1, 1, 1, 2, 2, 2, 2, 1, 1, 1, 1, 2, 2, 2, 2},
		)
		dataIn, _ := inPkt.MarshalBinary()
		pktSent, _ := NewAPacket(
			inPkt.AppEUI(),
			inPkt.DevEUI(),
			[]byte("Payload"),
			[]Metadata{inPkt.Metadata()},
		)
		devStorage := newMockDevStorage()
		devStorage.OutLookup = devEntry{
			Recipient: dataRecipient,
			DevAddr:   lorawan.DevAddr([4]byte{2, 2, 2, 2}),
			AppSKey:   [16]byte{1, 1, 1, 1, 2, 2, 2, 2, 1, 1, 1, 1, 2, 2, 2, 2},
			NwkSKey:   [16]byte{4, 4, 4, 4, 3, 3, 3, 3, 4, 4, 4, 4, 3, 3, 3, 3},
		}
		pktStorage := newMockPktStorage()
		broker := NewMockJSONRecipient()

		// Operate
		handler := New(devStorage, pktStorage, broker, GetLogger(t, "Handler"))
		err := handler.HandleUp(dataIn, an, adapter)

		// Check
		CheckErrors(t, nil, err)
		CheckPushed(t, nil, pktStorage.InPush)
		CheckPersonalized(t, nil, devStorage.InStorePersonalized)
		CheckAcks(t, true, an.InAck)
		CheckSent(t, pktSent, adapter.InSendPacket)
		CheckRecipients(t, []Recipient{adapter.OutGetRecipient}, adapter.InSendRecipients)
	}

	// --------------------

	{
		Desc(t, "Handle uplink with 2 packets in a row | No downlink ready")

		// Build
		recipient := NewMockJSONRecipient()
		dataRecipient, _ := recipient.MarshalBinary()

		// First Packet
		adapter1 := NewMockAdapter()
		adapter1.OutGetRecipient = recipient
		an1 := NewMockAckNacker()
		inPkt1 := newHPacket(
			[8]byte{1, 1, 1, 1, 1, 1, 1, 1},
			[8]byte{2, 2, 2, 2, 2, 2, 2, 2},
			"Payload",
			Metadata{
				Datr:    pointer.String("SF7BW125"),
				DutyRX1: pointer.Uint(uint(dutycycle.StateWarning)),
				DutyRX2: pointer.Uint(uint(dutycycle.StateWarning)),
				Rssi:    pointer.Int(-20),
				Lsnr:    pointer.Float64(5.0),
			},
			10,
			[16]byte{1, 1, 1, 1, 2, 2, 2, 2, 1, 1, 1, 1, 2, 2, 2, 2},
		)
		dataIn1, _ := inPkt1.MarshalBinary()

		// Second Packet
		adapter2 := NewMockAdapter()
		adapter2.OutGetRecipient = recipient
		an2 := NewMockAckNacker()
		inPkt2 := newHPacket(
			[8]byte{1, 1, 1, 1, 1, 1, 1, 1},
			[8]byte{2, 2, 2, 2, 2, 2, 2, 2},
			"Payload",
			Metadata{
				Datr:    pointer.String("SF7BW125"),
				DutyRX1: pointer.Uint(uint(dutycycle.StateAvailable)),
				DutyRX2: pointer.Uint(uint(dutycycle.StateAvailable)),
				Rssi:    pointer.Int(-20),
				Lsnr:    pointer.Float64(5.0),
			},
			10,
			[16]byte{1, 1, 1, 1, 2, 2, 2, 2, 1, 1, 1, 1, 2, 2, 2, 2},
		)
		dataIn2, _ := inPkt2.MarshalBinary()

		// Expected Sent
		pktSent, _ := NewAPacket(
			inPkt1.AppEUI(),
			inPkt1.DevEUI(),
			[]byte("Payload"),
			[]Metadata{inPkt1.Metadata(), inPkt2.Metadata()},
		)

		devStorage := newMockDevStorage()
		devStorage.OutLookup = devEntry{
			Recipient: dataRecipient,
			DevAddr:   lorawan.DevAddr([4]byte{2, 2, 2, 2}),
			AppSKey:   [16]byte{1, 1, 1, 1, 2, 2, 2, 2, 1, 1, 1, 1, 2, 2, 2, 2},
			NwkSKey:   [16]byte{4, 4, 4, 4, 3, 3, 3, 3, 4, 4, 4, 4, 3, 3, 3, 3},
		}
		pktStorage := newMockPktStorage()
		broker := NewMockJSONRecipient()

		// Operate
		handler := New(devStorage, pktStorage, broker, GetLogger(t, "Handler"))
		cherr := make(chan bool)
		go func() {
			var ok bool
			defer func(ok *bool) { cherr <- *ok }(&ok)
			err := handler.HandleUp(dataIn1, an1, adapter1)
			// Check
			CheckErrors(t, nil, err)
			CheckAcks(t, true, an1.InAck)
			CheckSent(t, pktSent, adapter1.InSendPacket) // We actually transfer to the first bundle
			CheckRecipients(t, []Recipient{recipient}, adapter1.InSendRecipients)
			ok = true
		}()

		go func() {
			<-time.After(time.Millisecond * 50)
			var ok bool
			defer func(ok *bool) { cherr <- *ok }(&ok)
			err := handler.HandleUp(dataIn2, an2, adapter2)
			// Check
			CheckErrors(t, nil, err)
			CheckAcks(t, true, an2.InAck)
			CheckSent(t, nil, adapter2.InSendPacket)
			CheckRecipients(t, nil, adapter2.InSendRecipients)
			ok = true
		}()

		// Check
		ok1 := <-cherr
		ok2 := <-cherr
		if !(ok1 && ok2) {
			return
		}
		CheckPushed(t, nil, pktStorage.InPush)
		CheckPersonalized(t, nil, devStorage.InStorePersonalized)
	}

	// --------------------

	{
		Desc(t, "Handle uplink with 1 packet | One downlink response")

		// Build
		recipient := NewMockJSONRecipient()
		dataRecipient, _ := recipient.MarshalBinary()
		an := NewMockAckNacker()
		adapter := NewMockAdapter()
		adapter.OutGetRecipient = recipient
		tmst := time.Now()
		inPkt := newHPacket(
			[8]byte{1, 1, 1, 1, 1, 1, 1, 1},
			[8]byte{2, 2, 2, 2, 2, 2, 2, 2},
			"Payload",
			Metadata{
				Datr:    pointer.String("SF7BW125"),
				Freq:    pointer.Float64(865.5),
				Tmst:    pointer.Uint(uint(tmst.Unix() * 1000)),
				Codr:    pointer.String("4/5"),
				DutyRX1: pointer.Uint(uint(dutycycle.StateAvailable)),
				DutyRX2: pointer.Uint(uint(dutycycle.StateAvailable)),
				Rssi:    pointer.Int(-20),
				Lsnr:    pointer.Float64(5.0),
			},
			10,
			[16]byte{1, 1, 1, 1, 2, 2, 2, 2, 1, 1, 1, 1, 2, 2, 2, 2},
		)
		dataIn, _ := inPkt.MarshalBinary()
		pktSent, _ := NewAPacket(
			inPkt.AppEUI(),
			inPkt.DevEUI(),
			[]byte("Payload"),
			[]Metadata{inPkt.Metadata()},
		)
		brkResp := newBPacket(
			[4]byte{2, 2, 2, 2},
			"Downlink",
			Metadata{
				Datr: pointer.String("SF7BW125"),
				Freq: pointer.Float64(865.5),
				Tmst: pointer.Uint(uint(tmst.Add(time.Second).Unix() * 1000)),
				Codr: pointer.String("4/5"),
				Size: pointer.Uint(21),
			},
			4,
			[16]byte{1, 1, 1, 1, 2, 2, 2, 2, 1, 1, 1, 1, 2, 2, 2, 2},
		)
		appResp, _ := NewAPacket(
			inPkt.AppEUI(),
			inPkt.DevEUI(),
			[]byte("Downlink"),
			[]Metadata{},
		)

		devStorage := newMockDevStorage()
		devStorage.OutLookup = devEntry{
			Recipient: dataRecipient,
			DevAddr:   lorawan.DevAddr([4]byte{2, 2, 2, 2}),
			AppSKey:   [16]byte{1, 1, 1, 1, 2, 2, 2, 2, 1, 1, 1, 1, 2, 2, 2, 2},
			NwkSKey:   [16]byte{4, 4, 4, 4, 3, 3, 3, 3, 4, 4, 4, 4, 3, 3, 3, 3},
			FCntDown:  3,
		}
		pktStorage := newMockPktStorage()
		pktStorage.OutPull = appResp
		broker := NewMockJSONRecipient()

		// Operate
		handler := New(devStorage, pktStorage, broker, GetLogger(t, "Handler"))
		err := handler.HandleUp(dataIn, an, adapter)

		// Check
		CheckErrors(t, nil, err)
		CheckPushed(t, nil, pktStorage.InPush)
		CheckPersonalized(t, nil, devStorage.InStorePersonalized)
		CheckAcks(t, brkResp, an.InAck)
		CheckSent(t, pktSent, adapter.InSendPacket)
		CheckRecipients(t, []Recipient{recipient}, adapter.InSendRecipients)
	}

	// ---------------

	{
		Desc(t, "Handle a late uplink | No downlink ready")

		// Build
		recipient := NewMockJSONRecipient()
		dataRecipient, _ := recipient.MarshalBinary()
		an2 := NewMockAckNacker()
		an1 := NewMockAckNacker()
		adapter1 := NewMockAdapter()
		adapter1.OutGetRecipient = recipient
		adapter2 := NewMockAdapter()
		adapter2.OutGetRecipient = recipient
		inPkt := newHPacket(
			[8]byte{1, 1, 1, 1, 1, 1, 1, 1},
			[8]byte{2, 2, 2, 2, 2, 2, 2, 2},
			"Payload",
			Metadata{
				Datr: pointer.String("SF7BW125"),
			},
			10,
			[16]byte{1, 1, 1, 1, 2, 2, 2, 2, 1, 1, 1, 1, 2, 2, 2, 2},
		)
		dataIn, _ := inPkt.MarshalBinary()
		pktSent, _ := NewAPacket(
			inPkt.AppEUI(),
			inPkt.DevEUI(),
			[]byte("Payload"),
			[]Metadata{inPkt.Metadata()},
		)
		devStorage := newMockDevStorage()
		devStorage.OutLookup = devEntry{
			Recipient: dataRecipient,
			DevAddr:   lorawan.DevAddr([4]byte{2, 2, 2, 2}),
			AppSKey:   [16]byte{1, 1, 1, 1, 2, 2, 2, 2, 1, 1, 1, 1, 2, 2, 2, 2},
			NwkSKey:   [16]byte{4, 4, 4, 4, 3, 3, 3, 3, 4, 4, 4, 4, 3, 3, 3, 3},
		}
		pktStorage := newMockPktStorage()
		broker := NewMockJSONRecipient()

		// Operate
		handler := New(devStorage, pktStorage, broker, GetLogger(t, "Handler"))
		cherr := make(chan bool)
		go func() {
			var ok bool
			defer func(ok *bool) { cherr <- *ok }(&ok)
			err := handler.HandleUp(dataIn, an1, adapter1)
			// Check
			CheckErrors(t, nil, err)
			CheckAcks(t, true, an1.InAck)
			CheckSent(t, pktSent, adapter1.InSendPacket)
			ok = true
		}()
		go func() {
			var ok bool
			defer func(ok *bool) { cherr <- *ok }(&ok)
			<-time.After(2 * bufferDelay)
			err := handler.HandleUp(dataIn, an2, adapter2)
			// Check
			CheckErrors(t, pointer.String(string(errors.Behavioural)), err)
			CheckAcks(t, false, an2.InAck)
			CheckSent(t, nil, adapter2.InSendPacket)
			ok = true
		}()

		// Check
		if !(<-cherr && <-cherr) {
			return
		}
		CheckPushed(t, nil, pktStorage.InPush)
		CheckPersonalized(t, nil, devStorage.InStorePersonalized)
	}

	// --------------------

	{
		Desc(t, "Handle uplink with 1 packet | No downlink ready | Adapter fail sending ")

		// Build
		recipient := NewMockJSONRecipient()
		dataRecipient, _ := recipient.MarshalBinary()
		an := NewMockAckNacker()
		adapter := NewMockAdapter()
		adapter.OutGetRecipient = recipient
		adapter.Failures["Send"] = errors.New(errors.Operational, "Mock Error: Unable to send")
		inPkt := newHPacket(
			[8]byte{1, 1, 1, 1, 1, 1, 1, 1},
			[8]byte{2, 2, 2, 2, 2, 2, 2, 2},
			"Payload",
			Metadata{
				Datr: pointer.String("SF7BW125"),
			},
			10,
			[16]byte{1, 1, 1, 1, 2, 2, 2, 2, 1, 1, 1, 1, 2, 2, 2, 2},
		)
		dataIn, _ := inPkt.MarshalBinary()
		pktSent, _ := NewAPacket(
			inPkt.AppEUI(),
			inPkt.DevEUI(),
			[]byte("Payload"),
			[]Metadata{inPkt.Metadata()},
		)

		devStorage := newMockDevStorage()
		devStorage.OutLookup = devEntry{
			Recipient: dataRecipient,
			DevAddr:   lorawan.DevAddr([4]byte{2, 2, 2, 2}),
			AppSKey:   [16]byte{1, 1, 1, 1, 2, 2, 2, 2, 1, 1, 1, 1, 2, 2, 2, 2},
			NwkSKey:   [16]byte{4, 4, 4, 4, 3, 3, 3, 3, 4, 4, 4, 4, 3, 3, 3, 3},
		}
		pktStorage := newMockPktStorage()
		broker := NewMockJSONRecipient()

		// Operate
		handler := New(devStorage, pktStorage, broker, GetLogger(t, "Handler"))
		err := handler.HandleUp(dataIn, an, adapter)

		// Check
		CheckErrors(t, pointer.String(string(errors.Operational)), err)
		CheckPushed(t, nil, pktStorage.InPush)
		CheckPersonalized(t, nil, devStorage.InStorePersonalized)
		CheckAcks(t, false, an.InAck)
		CheckSent(t, pktSent, adapter.InSendPacket)
		CheckRecipients(t, []Recipient{recipient}, adapter.InSendRecipients)
	}

	// --------------------

	{
		Desc(t, "Handle uplink with 1 packet | No downlink ready | Adapter fail GetRecipient")

		// Build
		recipient := NewMockJSONRecipient()
		dataRecipient, _ := recipient.MarshalBinary()
		an := NewMockAckNacker()
		adapter := NewMockAdapter()
		adapter.OutGetRecipient = recipient
		adapter.Failures["GetRecipient"] = errors.New(errors.Structural, "Mock Error: Unable to get recipient")
		inPkt := newHPacket(
			[8]byte{1, 1, 1, 1, 1, 1, 1, 1},
			[8]byte{2, 2, 2, 2, 2, 2, 2, 2},
			"Payload",
			Metadata{
				Datr: pointer.String("SF7BW125"),
			},
			10,
			[16]byte{1, 1, 1, 1, 2, 2, 2, 2, 1, 1, 1, 1, 2, 2, 2, 2},
		)
		dataIn, _ := inPkt.MarshalBinary()

		devStorage := newMockDevStorage()
		devStorage.OutLookup = devEntry{
			Recipient: dataRecipient,
			DevAddr:   lorawan.DevAddr([4]byte{2, 2, 2, 2}),
			AppSKey:   [16]byte{1, 1, 1, 1, 2, 2, 2, 2, 1, 1, 1, 1, 2, 2, 2, 2},
			NwkSKey:   [16]byte{4, 4, 4, 4, 3, 3, 3, 3, 4, 4, 4, 4, 3, 3, 3, 3},
		}
		pktStorage := newMockPktStorage()
		broker := NewMockJSONRecipient()

		// Operate
		handler := New(devStorage, pktStorage, broker, GetLogger(t, "Handler"))
		err := handler.HandleUp(dataIn, an, adapter)

		// Check
		CheckErrors(t, pointer.String(string(errors.Structural)), err)
		CheckPushed(t, nil, pktStorage.InPush)
		CheckPersonalized(t, nil, devStorage.InStorePersonalized)
		CheckAcks(t, false, an.InAck)
		CheckSent(t, nil, adapter.InSendPacket)
		CheckRecipients(t, nil, adapter.InSendRecipients)
	}

	// --------------------

	{
		Desc(t, "Handle uplink with 1 packet | No downlink ready | PktStorage fails to pull")

		// Build
		recipient := NewMockJSONRecipient()
		dataRecipient, _ := recipient.MarshalBinary()
		an := NewMockAckNacker()
		adapter := NewMockAdapter()
		adapter.OutGetRecipient = recipient
		inPkt := newHPacket(
			[8]byte{1, 1, 1, 1, 1, 1, 1, 1},
			[8]byte{2, 2, 2, 2, 2, 2, 2, 2},
			"Payload",
			Metadata{
				DutyRX1: pointer.Uint(uint(dutycycle.StateAvailable)),
				DutyRX2: pointer.Uint(uint(dutycycle.StateAvailable)),
				Rssi:    pointer.Int(-20),
				Lsnr:    pointer.Float64(5.0),
				Datr:    pointer.String("SF7BW125"),
			},
			10,
			[16]byte{1, 1, 1, 1, 2, 2, 2, 2, 1, 1, 1, 1, 2, 2, 2, 2},
		)
		dataIn, _ := inPkt.MarshalBinary()
		pktSent, _ := NewAPacket(
			inPkt.AppEUI(),
			inPkt.DevEUI(),
			[]byte("Payload"),
			[]Metadata{inPkt.Metadata()},
		)

		devStorage := newMockDevStorage()
		devStorage.OutLookup = devEntry{
			Recipient: dataRecipient,
			DevAddr:   lorawan.DevAddr([4]byte{2, 2, 2, 2}),
			AppSKey:   [16]byte{1, 1, 1, 1, 2, 2, 2, 2, 1, 1, 1, 1, 2, 2, 2, 2},
			NwkSKey:   [16]byte{4, 4, 4, 4, 3, 3, 3, 3, 4, 4, 4, 4, 3, 3, 3, 3},
		}
		pktStorage := newMockPktStorage()
		pktStorage.Failures["Pull"] = errors.New(errors.Operational, "Mock Error: Failed to Pull")
		broker := NewMockJSONRecipient()

		// Operate
		handler := New(devStorage, pktStorage, broker, GetLogger(t, "Handler"))
		err := handler.HandleUp(dataIn, an, adapter)

		// Check
		CheckErrors(t, pointer.String(string(errors.Operational)), err)
		CheckPushed(t, nil, pktStorage.InPush)
		CheckPersonalized(t, nil, devStorage.InStorePersonalized)
		CheckAcks(t, false, an.InAck)
		CheckSent(t, pktSent, adapter.InSendPacket)
		CheckRecipients(t, []Recipient{recipient}, adapter.InSendRecipients)
	}

	// --------------------

	{
		Desc(t, "Handle uplink with 2 different packets from same app | No downlink ready")

		// Build
		recipient := NewMockJSONRecipient()
		dataRecipient, _ := recipient.MarshalBinary()

		// First Packet
		adapter1 := NewMockAdapter()
		adapter1.OutGetRecipient = recipient
		an1 := NewMockAckNacker()
		inPkt1 := newHPacket(
			[8]byte{1, 1, 1, 1, 1, 1, 1, 1},
			[8]byte{2, 2, 2, 2, 2, 2, 2, 2},
			"PayloadPacket1",
			Metadata{
				Datr:    pointer.String("SF7BW125"),
				DutyRX1: pointer.Uint(uint(dutycycle.StateAvailable)),
				DutyRX2: pointer.Uint(uint(dutycycle.StateAvailable)),
				Rssi:    pointer.Int(-20),
				Lsnr:    pointer.Float64(5.0),
			},
			10,
			[16]byte{1, 1, 1, 1, 2, 2, 2, 2, 1, 1, 1, 1, 2, 2, 2, 2},
		)
		dataIn1, _ := inPkt1.MarshalBinary()

		// Second Packet
		adapter2 := NewMockAdapter()
		adapter2.OutGetRecipient = recipient
		an2 := NewMockAckNacker()
		inPkt2 := newHPacket(
			[8]byte{1, 1, 1, 1, 1, 1, 1, 1},
			[8]byte{2, 2, 2, 2, 2, 2, 2, 2},
			"PayloadPacket2",
			Metadata{
				Datr:    pointer.String("SF7BW125"),
				DutyRX1: pointer.Uint(uint(dutycycle.StateAvailable)),
				DutyRX2: pointer.Uint(uint(dutycycle.StateAvailable)),
				Rssi:    pointer.Int(-20),
				Lsnr:    pointer.Float64(5.0),
			},
			11,
			[16]byte{1, 1, 1, 1, 2, 2, 2, 2, 1, 1, 1, 1, 2, 2, 2, 2},
		)
		dataIn2, _ := inPkt2.MarshalBinary()

		// Expected responses
		pktSent1, _ := NewAPacket(
			inPkt1.AppEUI(),
			inPkt1.DevEUI(),
			[]byte("PayloadPacket1"),
			[]Metadata{inPkt1.Metadata()},
		)

		pktSent2, _ := NewAPacket(
			inPkt1.AppEUI(),
			inPkt1.DevEUI(),
			[]byte("PayloadPacket2"),
			[]Metadata{inPkt2.Metadata()},
		)

		devStorage := newMockDevStorage()
		devStorage.OutLookup = devEntry{
			Recipient: dataRecipient,
			DevAddr:   lorawan.DevAddr([4]byte{2, 2, 2, 2}),
			AppSKey:   [16]byte{1, 1, 1, 1, 2, 2, 2, 2, 1, 1, 1, 1, 2, 2, 2, 2},
			NwkSKey:   [16]byte{4, 4, 4, 4, 3, 3, 3, 3, 4, 4, 4, 4, 3, 3, 3, 3},
		}
		pktStorage := newMockPktStorage()
		broker := NewMockJSONRecipient()

		// Operate #1
		handler := New(devStorage, pktStorage, broker, GetLogger(t, "Handler"))
		err := handler.HandleUp(dataIn1, an1, adapter1)

		// Check #1
		CheckErrors(t, nil, err)
		CheckAcks(t, true, an1.InAck)
		CheckSent(t, pktSent1, adapter1.InSendPacket)
		CheckRecipients(t, []Recipient{recipient}, adapter1.InSendRecipients)
		CheckPushed(t, nil, pktStorage.InPush)
		CheckPersonalized(t, nil, devStorage.InStorePersonalized)

		// Operate #2
		<-time.After(150 * time.Millisecond)
		err = handler.HandleUp(dataIn2, an2, adapter2)

		// Check
		CheckErrors(t, nil, err)
		CheckAcks(t, true, an2.InAck)
		CheckSent(t, pktSent2, adapter2.InSendPacket)
		CheckRecipients(t, []Recipient{recipient}, adapter2.InSendRecipients)
		CheckPushed(t, nil, pktStorage.InPush)
		CheckPersonalized(t, nil, devStorage.InStorePersonalized)
	}

	// --------------------

	{
		Desc(t, "Handle uplink with 1 packet | One downlink response, mising metadata in uplink")

		// Build
		recipient := NewMockJSONRecipient()
		dataRecipient, _ := recipient.MarshalBinary()
		an := NewMockAckNacker()
		adapter := NewMockAdapter()
		adapter.OutGetRecipient = recipient
		inPkt := newHPacket(
			[8]byte{1, 1, 1, 1, 1, 1, 1, 1},
			[8]byte{2, 2, 2, 2, 2, 2, 2, 2},
			"Payload",
			Metadata{
				Datr:    pointer.String("SF7BW125"),
				Freq:    pointer.Float64(865.5),
				DutyRX1: pointer.Uint(uint(dutycycle.StateAvailable)),
				DutyRX2: pointer.Uint(uint(dutycycle.StateAvailable)),
				Rssi:    pointer.Int(-20),
				Lsnr:    pointer.Float64(5.0),
			},
			10,
			[16]byte{1, 1, 1, 1, 2, 2, 2, 2, 1, 1, 1, 1, 2, 2, 2, 2},
		)
		dataIn, _ := inPkt.MarshalBinary()
		pktSent, _ := NewAPacket(
			inPkt.AppEUI(),
			inPkt.DevEUI(),
			[]byte("Payload"),
			[]Metadata{inPkt.Metadata()},
		)
		appResp, _ := NewAPacket(
			inPkt.AppEUI(),
			inPkt.DevEUI(),
			[]byte("Downlink"),
			[]Metadata{},
		)

		devStorage := newMockDevStorage()
		devStorage.OutLookup = devEntry{
			Recipient: dataRecipient,
			DevAddr:   lorawan.DevAddr([4]byte{2, 2, 2, 2}),
			AppSKey:   [16]byte{1, 1, 1, 1, 2, 2, 2, 2, 1, 1, 1, 1, 2, 2, 2, 2},
			NwkSKey:   [16]byte{4, 4, 4, 4, 3, 3, 3, 3, 4, 4, 4, 4, 3, 3, 3, 3},
			FCntDown:  3,
		}
		pktStorage := newMockPktStorage()
		pktStorage.OutPull = appResp
		broker := NewMockJSONRecipient()

		// Operate
		handler := New(devStorage, pktStorage, broker, GetLogger(t, "Handler"))
		err := handler.HandleUp(dataIn, an, adapter)

		// Check
		CheckErrors(t, pointer.String(string(errors.Structural)), err)
		CheckPushed(t, nil, pktStorage.InPush)
		CheckPersonalized(t, nil, devStorage.InStorePersonalized)
		CheckAcks(t, false, an.InAck)
		CheckSent(t, pktSent, adapter.InSendPacket)
		CheckRecipients(t, []Recipient{recipient}, adapter.InSendRecipients)
	}

	// --------------------

	{
		Desc(t, "Handle uplink with 1 packet | One downlink response | Only RX2 available")

		// Build
		recipient := NewMockJSONRecipient()
		dataRecipient, _ := recipient.MarshalBinary()
		an := NewMockAckNacker()
		adapter := NewMockAdapter()
		adapter.OutGetRecipient = recipient
		tmst := time.Now()
		inPkt := newHPacket(
			[8]byte{1, 1, 1, 1, 1, 1, 1, 1},
			[8]byte{2, 2, 2, 2, 2, 2, 2, 2},
			"Payload",
			Metadata{
				Datr:    pointer.String("SF7BW125"),
				Freq:    pointer.Float64(865.5),
				Tmst:    pointer.Uint(uint(tmst.Unix() * 1000)),
				Codr:    pointer.String("4/5"),
				DutyRX1: pointer.Uint(uint(dutycycle.StateBlocked)),
				DutyRX2: pointer.Uint(uint(dutycycle.StateAvailable)),
				Rssi:    pointer.Int(-20),
				Lsnr:    pointer.Float64(5.0),
			},
			10,
			[16]byte{1, 1, 1, 1, 2, 2, 2, 2, 1, 1, 1, 1, 2, 2, 2, 2},
		)
		dataIn, _ := inPkt.MarshalBinary()
		pktSent, _ := NewAPacket(
			inPkt.AppEUI(),
			inPkt.DevEUI(),
			[]byte("Payload"),
			[]Metadata{inPkt.Metadata()},
		)
		brkResp := newBPacket(
			[4]byte{2, 2, 2, 2},
			"Downlink",
			Metadata{
				Datr: pointer.String("SF9BW125"),
				Freq: pointer.Float64(869.5),
				Tmst: pointer.Uint(uint(tmst.Add(2*time.Second).Unix() * 1000)),
				Codr: pointer.String("4/5"),
				Size: pointer.Uint(21),
			},
			15,
			[16]byte{1, 1, 1, 1, 2, 2, 2, 2, 1, 1, 1, 1, 2, 2, 2, 2},
		)
		appResp, _ := NewAPacket(
			inPkt.AppEUI(),
			inPkt.DevEUI(),
			[]byte("Downlink"),
			[]Metadata{},
		)

		devStorage := newMockDevStorage()
		devStorage.OutLookup = devEntry{
			Recipient: dataRecipient,
			DevAddr:   lorawan.DevAddr([4]byte{2, 2, 2, 2}),
			AppSKey:   [16]byte{1, 1, 1, 1, 2, 2, 2, 2, 1, 1, 1, 1, 2, 2, 2, 2},
			NwkSKey:   [16]byte{4, 4, 4, 4, 3, 3, 3, 3, 4, 4, 4, 4, 3, 3, 3, 3},
			FCntDown:  14,
		}
		pktStorage := newMockPktStorage()
		pktStorage.OutPull = appResp
		broker := NewMockJSONRecipient()

		// Operate
		handler := New(devStorage, pktStorage, broker, GetLogger(t, "Handler"))
		err := handler.HandleUp(dataIn, an, adapter)

		// Check
		CheckErrors(t, nil, err)
		CheckPushed(t, nil, pktStorage.InPush)
		CheckPersonalized(t, nil, devStorage.InStorePersonalized)
		CheckAcks(t, brkResp, an.InAck)
		CheckSent(t, pktSent, adapter.InSendPacket)
		CheckRecipients(t, []Recipient{recipient}, adapter.InSendRecipients)
	}
}

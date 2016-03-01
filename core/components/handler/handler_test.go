// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package handler

import (
	"fmt"
	"reflect"
	"sync"
	"testing"
	"time"

	. "github.com/TheThingsNetwork/ttn/core"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	. "github.com/TheThingsNetwork/ttn/utils/errors/checks"
	"github.com/TheThingsNetwork/ttn/utils/pointer"
	. "github.com/TheThingsNetwork/ttn/utils/testing"
	"github.com/brocaar/lorawan"
)

func TestRegister(t *testing.T) {
	{
		Desc(t, "Register valid HRegistration")

		devStorage := newMockDevStorage()
		pktStorage := newMockPktStorage()
		an := newMockAckNacker()
		handler := New(devStorage, pktStorage, GetLogger(t, "Handler"))
		r := newMockRegistration(
			[8]byte{1, 1, 1, 1, 1, 1, 1, 1},
			[8]byte{2, 2, 2, 2, 2, 2, 2, 2},
			newMockRecipient("recipient"),
		)

		err := handler.Register(r, an)

		CheckErrors(t, nil, err)
		CheckPushed(t, nil, pktStorage.Pushed)
		CheckPersonalized(t, r, devStorage.Personalized)
	}

	// --------------------

	{
		Desc(t, "Register invalid HRegistration")

		devStorage := newMockDevStorage()
		pktStorage := newMockPktStorage()
		an := newMockAckNacker()
		handler := New(devStorage, pktStorage, GetLogger(t, "Handler"))

		err := handler.Register(nil, an)

		CheckErrors(t, pointer.String(string(errors.Structural)), err)
		CheckPushed(t, nil, pktStorage.Pushed)
		CheckPersonalized(t, nil, devStorage.Personalized)
	}

	// --------------------

	{
		Desc(t, "Register valid HRegistration | devStorage fails")

		devStorage := newMockDevStorage()
		pktStorage := newMockPktStorage()
		an := newMockAckNacker()
		handler := New(devStorage, pktStorage, GetLogger(t, "Handler"))
		r := newMockRegistration(
			[8]byte{1, 1, 1, 1, 1, 1, 1, 1},
			[8]byte{2, 2, 2, 2, 2, 2, 2, 2},
			newMockRecipient("recipient"),
		)

		devStorage.Failures["StorePersonalized"] = errors.New(errors.Operational, "Mock Error")
		err := handler.Register(r, an)

		CheckErrors(t, pointer.String(string(errors.Operational)), err)
		CheckPushed(t, nil, pktStorage.Pushed)
		CheckPersonalized(t, r, devStorage.Personalized)
	}
}

func TestHandleDown(t *testing.T) {
	{
		Desc(t, "Handle downlink APacket")

		devStorage := newMockDevStorage()
		pktStorage := newMockPktStorage()
		an := newMockAckNacker()
		adapter := newMockAdapter()
		handler := New(devStorage, pktStorage, GetLogger(t, "Handler"))
		pkt, _ := NewAPacket(
			[8]byte{1, 1, 1, 1, 1, 1, 1, 1},
			[8]byte{2, 2, 2, 2, 2, 2, 2, 2},
			[]byte("TheThingsNetwork"),
			[]Metadata{},
		)

		data, _ := pkt.MarshalBinary()
		err := handler.HandleDown(data, an, adapter)

		CheckErrors(t, nil, err)
		CheckPushed(t, pkt, pktStorage.Pushed)
		CheckPersonalized(t, nil, devStorage.Personalized)
		CheckAcks(t, true, an.Acked)
		CheckSent(t, nil, adapter.SentPkt)
		CheckRecipients(t, nil, adapter.SentRecipients)
	}

	// --------------------

	{
		Desc(t, "Handle downlink wrong data")

		devStorage := newMockDevStorage()
		pktStorage := newMockPktStorage()
		an := newMockAckNacker()
		adapter := newMockAdapter()
		handler := New(devStorage, pktStorage, GetLogger(t, "Handler"))

		err := handler.HandleDown([]byte{1, 2, 3}, an, adapter)

		CheckErrors(t, pointer.String(string(errors.Structural)), err)
		CheckPushed(t, nil, pktStorage.Pushed)
		CheckPersonalized(t, nil, devStorage.Personalized)
		CheckAcks(t, false, an.Acked)
		CheckSent(t, nil, adapter.SentPkt)
		CheckRecipients(t, nil, adapter.SentRecipients)
	}

	// --------------------

	{
		Desc(t, "Handle downlink wrong packet type")

		devStorage := newMockDevStorage()
		pktStorage := newMockPktStorage()
		an := newMockAckNacker()
		adapter := newMockAdapter()
		handler := New(devStorage, pktStorage, GetLogger(t, "Handler"))
		pkt := NewJPacket(
			lorawan.EUI64([8]byte{1, 1, 1, 1, 1, 1, 1, 1}),
			lorawan.EUI64([8]byte{2, 2, 2, 2, 2, 2, 2, 2}),
			[2]byte{14, 42},
			Metadata{},
		)
		data, _ := pkt.MarshalBinary()

		err := handler.HandleDown(data, an, adapter)

		CheckErrors(t, pointer.String(string(errors.Implementation)), err)
		CheckPushed(t, nil, pktStorage.Pushed)
		CheckPersonalized(t, nil, devStorage.Personalized)
		CheckAcks(t, false, an.Acked)
		CheckSent(t, nil, adapter.SentPkt)
		CheckRecipients(t, nil, adapter.SentRecipients)
	}
}

func TestHandleUp(t *testing.T) {
	{
		Desc(t, "Handle uplink with 1 packet | No Associated App")

		devStorage := newMockDevStorage()
		pktStorage := newMockPktStorage()
		an := newMockAckNacker()
		adapter := newMockAdapter()
		handler := New(devStorage, pktStorage, GetLogger(t, "Handler"))
		inPkt := newHPacket(
			[8]byte{1, 1, 1, 1, 1, 1, 1, 1},
			[8]byte{2, 2, 2, 2, 2, 2, 2, 2},
			"Payload",
			Metadata{
				Duty: pointer.Uint(5),
				Rssi: pointer.Int(-25),
			},
			10,
			[16]byte{1, 1, 1, 1, 2, 2, 2, 2, 1, 1, 1, 1, 2, 2, 2, 2},
		)
		dataIn, _ := inPkt.MarshalBinary()

		devStorage.Failures["Lookup"] = errors.New(errors.Behavioural, "Mock: Not Found")
		pktStorage.Failures["Pull"] = errors.New(errors.Behavioural, "Mock: Not Found")
		err := handler.HandleUp(dataIn, an, adapter)

		CheckErrors(t, pointer.String(string(errors.Behavioural)), err)
		CheckPushed(t, nil, pktStorage.Pushed)
		CheckPersonalized(t, nil, devStorage.Personalized)
		CheckAcks(t, false, an.Acked)
		CheckSent(t, nil, adapter.SentPkt)
		CheckRecipients(t, nil, adapter.SentRecipients)
	}

	{
		Desc(t, "Handle uplink with invalid data")

		devStorage := newMockDevStorage()
		pktStorage := newMockPktStorage()
		an := newMockAckNacker()
		adapter := newMockAdapter()
		handler := New(devStorage, pktStorage, GetLogger(t, "Handler"))

		err := handler.HandleUp([]byte{1, 2, 3}, an, adapter)

		CheckErrors(t, pointer.String(string(errors.Structural)), err)
		CheckPushed(t, nil, pktStorage.Pushed)
		CheckPersonalized(t, nil, devStorage.Personalized)
		CheckAcks(t, false, an.Acked)
		CheckSent(t, nil, adapter.SentPkt)
		CheckRecipients(t, nil, adapter.SentRecipients)
	}

	// --------------------

	{
		Desc(t, "Handle uplink with 1 packet | No downlink ready")

		devStorage := newMockDevStorage()
		pktStorage := newMockPktStorage()
		an := newMockAckNacker()
		adapter := newMockAdapter()
		handler := New(devStorage, pktStorage, GetLogger(t, "Handler"))
		inPkt := newHPacket(
			[8]byte{1, 1, 1, 1, 1, 1, 1, 1},
			[8]byte{2, 2, 2, 2, 2, 2, 2, 2},
			"Payload",
			Metadata{
				Duty: pointer.Uint(5),
				Rssi: pointer.Int(-25),
			},
			10,
			[16]byte{1, 1, 1, 1, 2, 2, 2, 2, 1, 1, 1, 1, 2, 2, 2, 2},
		)
		dataIn, _ := inPkt.MarshalBinary()
		recipient := newMockRecipient("TowardsInfinity")
		dataRecipient, _ := recipient.MarshalBinary()
		pktSent, _ := NewAPacket(
			inPkt.AppEUI(),
			inPkt.DevEUI(),
			[]byte("Payload"),
			[]Metadata{inPkt.Metadata()},
		)

		adapter.Recipient = recipient
		devStorage.LookupEntry = devEntry{
			Recipient: dataRecipient,
			DevAddr:   lorawan.DevAddr([4]byte{2, 2, 2, 2}),
			AppSKey:   [16]byte{1, 1, 1, 1, 2, 2, 2, 2, 1, 1, 1, 1, 2, 2, 2, 2},
			NwkSKey:   [16]byte{4, 4, 4, 4, 3, 3, 3, 3, 4, 4, 4, 4, 3, 3, 3, 3},
		}
		err := handler.HandleUp(dataIn, an, adapter)

		CheckErrors(t, nil, err)
		CheckPushed(t, nil, pktStorage.Pushed)
		CheckPersonalized(t, nil, devStorage.Personalized)
		CheckAcks(t, true, an.Acked)
		CheckSent(t, pktSent, adapter.SentPkt)
		CheckRecipients(t, []Recipient{recipient}, adapter.SentRecipients)
	}

	// --------------------

	{
		Desc(t, "Handle uplink with 2 packets in a row | No downlink ready")

		// Handler
		devStorage := newMockDevStorage()
		pktStorage := newMockPktStorage()
		handler := New(devStorage, pktStorage, GetLogger(t, "Handler"))

		// Recipient
		recipient := newMockRecipient("TowardsInfinity")
		dataRecipient, _ := recipient.MarshalBinary()

		// First Packet
		adapter1 := newMockAdapter()
		adapter1.Recipient = recipient
		an1 := newMockAckNacker()
		inPkt1 := newHPacket(
			[8]byte{1, 1, 1, 1, 1, 1, 1, 1},
			[8]byte{2, 2, 2, 2, 2, 2, 2, 2},
			"Payload",
			Metadata{
				Duty: pointer.Uint(75),
				Rssi: pointer.Int(-25),
			},
			10,
			[16]byte{1, 1, 1, 1, 2, 2, 2, 2, 1, 1, 1, 1, 2, 2, 2, 2},
		)
		dataIn1, _ := inPkt1.MarshalBinary()

		// Second Packet
		adapter2 := newMockAdapter()
		adapter2.Recipient = recipient
		an2 := newMockAckNacker()
		inPkt2 := newHPacket(
			[8]byte{1, 1, 1, 1, 1, 1, 1, 1},
			[8]byte{2, 2, 2, 2, 2, 2, 2, 2},
			"Payload",
			Metadata{
				Duty: pointer.Uint(5),
				Rssi: pointer.Int(0),
			},
			10,
			[16]byte{1, 1, 1, 1, 2, 2, 2, 2, 1, 1, 1, 1, 2, 2, 2, 2},
		)
		dataIn2, _ := inPkt2.MarshalBinary()

		// Expected response
		pktSent, _ := NewAPacket(
			inPkt1.AppEUI(),
			inPkt1.DevEUI(),
			[]byte("Payload"),
			[]Metadata{inPkt1.Metadata(), inPkt2.Metadata()},
		)

		// Fake response from the storage
		done := sync.WaitGroup{}
		done.Add(2)
		devStorage.LookupEntry = devEntry{
			Recipient: dataRecipient,
			DevAddr:   lorawan.DevAddr([4]byte{2, 2, 2, 2}),
			AppSKey:   [16]byte{1, 1, 1, 1, 2, 2, 2, 2, 1, 1, 1, 1, 2, 2, 2, 2},
			NwkSKey:   [16]byte{4, 4, 4, 4, 3, 3, 3, 3, 4, 4, 4, 4, 3, 3, 3, 3},
		}
		go func() {
			defer done.Done()
			err := handler.HandleUp(dataIn1, an1, adapter1)
			CheckErrors(t, nil, err)
			CheckAcks(t, true, an1.Acked)
			CheckSent(t, nil, adapter1.SentPkt)
			CheckRecipients(t, nil, adapter1.SentRecipients)
		}()

		go func() {
			<-time.After(time.Millisecond * 50)
			defer done.Done()
			err := handler.HandleUp(dataIn2, an2, adapter2)
			CheckErrors(t, nil, err)
			CheckAcks(t, true, an2.Acked)
			CheckSent(t, pktSent, adapter2.SentPkt) // Adapter2 because the adapter of the best bundle even if they are supposed to be identical
			CheckRecipients(t, []Recipient{recipient}, adapter2.SentRecipients)
		}()

		done.Wait()
		CheckPushed(t, nil, pktStorage.Pushed)
		CheckPersonalized(t, nil, devStorage.Personalized)
	}

	// --------------------

	{
		Desc(t, "Handle uplink with 1 packet | One downlink response")

		devStorage := newMockDevStorage()
		pktStorage := newMockPktStorage()
		an := newMockAckNacker()
		adapter := newMockAdapter()
		handler := New(devStorage, pktStorage, GetLogger(t, "Handler"))
		inPkt := newHPacket(
			[8]byte{1, 1, 1, 1, 1, 1, 1, 1},
			[8]byte{2, 2, 2, 2, 2, 2, 2, 2},
			"Payload",
			Metadata{
				Duty: pointer.Uint(5),
				Rssi: pointer.Int(-25),
			},
			10,
			[16]byte{1, 1, 1, 1, 2, 2, 2, 2, 1, 1, 1, 1, 2, 2, 2, 2},
		)
		dataIn, _ := inPkt.MarshalBinary()
		recipient := newMockRecipient("TowardsInfinity")
		dataRecipient, _ := recipient.MarshalBinary()
		pktSent, _ := NewAPacket(
			inPkt.AppEUI(),
			inPkt.DevEUI(),
			[]byte("Payload"),
			[]Metadata{inPkt.Metadata()},
		)
		brkResp := newBPacket(
			[4]byte{2, 2, 2, 2},
			"Downlink",
			Metadata{},
			11,
			[16]byte{1, 1, 1, 1, 2, 2, 2, 2, 1, 1, 1, 1, 2, 2, 2, 2},
		)
		appResp, _ := NewAPacket(
			inPkt.AppEUI(),
			inPkt.DevEUI(),
			[]byte("Downlink"),
			[]Metadata{},
		)

		adapter.Recipient = recipient
		devStorage.LookupEntry = devEntry{
			Recipient: dataRecipient,
			DevAddr:   lorawan.DevAddr([4]byte{2, 2, 2, 2}),
			AppSKey:   [16]byte{1, 1, 1, 1, 2, 2, 2, 2, 1, 1, 1, 1, 2, 2, 2, 2},
			NwkSKey:   [16]byte{4, 4, 4, 4, 3, 3, 3, 3, 4, 4, 4, 4, 3, 3, 3, 3},
		}
		pktStorage.PullEntry = appResp
		err := handler.HandleUp(dataIn, an, adapter)

		CheckErrors(t, nil, err)
		CheckPushed(t, nil, pktStorage.Pushed)
		CheckPersonalized(t, nil, devStorage.Personalized)
		CheckAcks(t, brkResp, an.Acked)
		CheckSent(t, pktSent, adapter.SentPkt)
		CheckRecipients(t, []Recipient{recipient}, adapter.SentRecipients)
	}

	// ---------------

	{
		Desc(t, "Handle a late uplink | No downlink ready")

		devStorage := newMockDevStorage()
		pktStorage := newMockPktStorage()
		an2 := newMockAckNacker()
		an1 := newMockAckNacker()
		adapter1 := newMockAdapter()
		adapter2 := newMockAdapter()
		handler := New(devStorage, pktStorage, GetLogger(t, "Handler"))
		inPkt := newHPacket(
			[8]byte{1, 1, 1, 1, 1, 1, 1, 1},
			[8]byte{2, 2, 2, 2, 2, 2, 2, 2},
			"Payload",
			Metadata{
				Duty: pointer.Uint(5),
				Rssi: pointer.Int(-25),
			},
			10,
			[16]byte{1, 1, 1, 1, 2, 2, 2, 2, 1, 1, 1, 1, 2, 2, 2, 2},
		)
		dataIn, _ := inPkt.MarshalBinary()
		recipient := newMockRecipient("TowardsInfinity")
		dataRecipient, _ := recipient.MarshalBinary()
		pktSent, _ := NewAPacket(
			inPkt.AppEUI(),
			inPkt.DevEUI(),
			[]byte("Payload"),
			[]Metadata{inPkt.Metadata()},
		)

		adapter1.Recipient = recipient
		adapter2.Recipient = recipient
		devStorage.LookupEntry = devEntry{
			Recipient: dataRecipient,
			DevAddr:   lorawan.DevAddr([4]byte{2, 2, 2, 2}),
			AppSKey:   [16]byte{1, 1, 1, 1, 2, 2, 2, 2, 1, 1, 1, 1, 2, 2, 2, 2},
			NwkSKey:   [16]byte{4, 4, 4, 4, 3, 3, 3, 3, 4, 4, 4, 4, 3, 3, 3, 3},
		}

		done := sync.WaitGroup{}
		done.Add(2)

		go func() {
			defer done.Done()
			err := handler.HandleUp(dataIn, an1, adapter1)
			CheckErrors(t, nil, err)
			CheckAcks(t, true, an1.Acked)
			CheckSent(t, pktSent, adapter1.SentPkt)
		}()

		go func() {
			defer done.Done()
			<-time.After(2 * buffer_delay)
			err := handler.HandleUp(dataIn, an2, adapter2)
			CheckErrors(t, pointer.String(string(errors.Operational)), err)
			CheckAcks(t, false, an2.Acked)
			CheckSent(t, nil, adapter2.SentPkt)
		}()

		done.Wait()

		CheckPushed(t, nil, pktStorage.Pushed)
		CheckPersonalized(t, nil, devStorage.Personalized)
	}

}

// ----- TYPE utilities

//
// MOCK DEV STORAGE
//
type mockDevStorage struct {
	Failures     map[string]error
	LookupEntry  devEntry
	Personalized HRegistration
	Activated    HRegistration
}

func newMockDevStorage(failures ...string) *mockDevStorage {
	return &mockDevStorage{
		Failures: make(map[string]error),
	}
}

func (s mockDevStorage) Lookup(appEUI lorawan.EUI64, devEUI lorawan.EUI64) (devEntry, error) {
	if s.Failures["Lookup"] != nil {
		return devEntry{}, s.Failures["Lookup"]
	}
	return s.LookupEntry, nil
}

func (s *mockDevStorage) StorePersonalized(r HRegistration) error {
	s.Personalized = r
	return s.Failures["StorePersonalized"]
}

func (s *mockDevStorage) StoreActivated(r HRegistration) error {
	s.Activated = r
	return s.Failures["StoreActivated"]
}

func (s mockDevStorage) Close() error {
	return s.Failures["Close"]
}

//
// MOCK PKT STORAGE
//
type mockPktStorage struct {
	Failures  map[string]error
	PullEntry APacket
	Pushed    APacket
}

func newMockPktStorage() *mockPktStorage {
	return &mockPktStorage{
		Failures: make(map[string]error),
	}
}

func (s *mockPktStorage) Push(p APacket) error {
	s.Pushed = p
	return s.Failures["Push"]
}

func (s *mockPktStorage) Pull(appEUI lorawan.EUI64, devEUI lorawan.EUI64) (APacket, error) {
	if s.Failures["Pull"] != nil {
		return nil, s.Failures["Pull"]
	}
	return s.PullEntry, nil
}

func (s *mockPktStorage) Close() error {
	return s.Failures["Close"]
}

//
// MOCK ACK/NACKER
//
type mockAckNacker struct {
	Acked struct {
		Ack    *bool
		Packet Packet
	}
}

func newMockAckNacker() *mockAckNacker {
	return &mockAckNacker{}
}

func (an *mockAckNacker) Ack(p Packet) error {
	an.Acked = struct {
		Ack    *bool
		Packet Packet
	}{
		Ack:    pointer.Bool(true),
		Packet: p,
	}
	return nil
}

func (an *mockAckNacker) Nack() error {
	an.Acked = struct {
		Ack    *bool
		Packet Packet
	}{
		Ack:    pointer.Bool(false),
		Packet: nil,
	}
	return nil
}

//
// MOCK ADAPTER
//
type mockAdapter struct {
	Failures       map[string]error
	SentPkt        Packet
	SentRecipients []Recipient
	SendData       []byte
	Recipient      Recipient
	NextPacket     []byte
	NextAckNacker  AckNacker
	NextReg        Registration
}

func newMockAdapter() *mockAdapter {
	return &mockAdapter{
		Failures: make(map[string]error),
	}
}

func (a *mockAdapter) Send(p Packet, r ...Recipient) ([]byte, error) {
	a.SentPkt = p
	a.SentRecipients = r
	if a.Failures["Send"] != nil {
		return nil, a.Failures["Send"]
	}
	return a.SendData, nil
}

func (a *mockAdapter) GetRecipient(raw []byte) (Recipient, error) {
	if a.Failures["GetRecipient"] != nil {
		return nil, a.Failures["Send"]
	}
	return a.Recipient, nil
}

func (a *mockAdapter) Next() ([]byte, AckNacker, error) {
	if a.Failures["Next"] != nil {
		return nil, nil, a.Failures["Next"]
	}
	return a.NextPacket, a.NextAckNacker, nil
}

func (a *mockAdapter) NextRegistration() (Registration, AckNacker, error) {
	if a.Failures["NextRegistration"] != nil {
		return nil, nil, a.Failures["NextRegistration"]
	}
	return a.NextReg, a.NextAckNacker, nil
}

// ----- BUILD utilities
func newHPacket(appEUI [8]byte, devEUI [8]byte, payload string, metadata Metadata, fcnt uint32, appSKey [16]byte) HPacket {
	macPayload := lorawan.NewMACPayload(true)
	macPayload.FHDR = lorawan.FHDR{
		FCnt: fcnt,
	}
	macPayload.FPort = 1
	macPayload.FRMPayload = []lorawan.Payload{&lorawan.DataPayload{Bytes: []byte(payload)}}

	var key lorawan.AES128Key
	copy(key[:], appSKey[:])
	if err := macPayload.EncryptFRMPayload(key); err != nil {
		panic(err)
	}

	phyPayload := lorawan.NewPHYPayload(true)
	phyPayload.MHDR = lorawan.MHDR{
		MType: lorawan.UnconfirmedDataUp,
		Major: lorawan.LoRaWANR1,
	}
	phyPayload.MACPayload = macPayload

	var appEUIp lorawan.EUI64
	var devEUIp lorawan.EUI64
	copy(appEUIp[:], appEUI[:])
	copy(devEUIp[:], devEUI[:])

	packet, err := NewHPacket(appEUIp, devEUIp, phyPayload, metadata)
	if err != nil {
		panic(err)
	}
	return packet
}

func newBPacket(rawDevAddr [4]byte, payload string, metadata Metadata, fcnt uint32, appSKey [16]byte) BPacket {
	var devAddr lorawan.DevAddr
	copy(devAddr[:], rawDevAddr[:])

	macPayload := lorawan.NewMACPayload(false)
	macPayload.FHDR = lorawan.FHDR{
		DevAddr: devAddr,
		FCnt:    fcnt,
	}
	macPayload.FPort = 1
	macPayload.FRMPayload = []lorawan.Payload{&lorawan.DataPayload{Bytes: []byte(payload)}}

	var key lorawan.AES128Key
	copy(key[:], appSKey[:])
	if err := macPayload.EncryptFRMPayload(key); err != nil {
		panic(err)
	}

	phyPayload := lorawan.NewPHYPayload(false)
	phyPayload.MHDR = lorawan.MHDR{
		MType: lorawan.UnconfirmedDataDown,
		Major: lorawan.LoRaWANR1,
	}
	phyPayload.MACPayload = macPayload

	packet, err := NewBPacket(phyPayload, metadata)
	if err != nil {
		panic(err)
	}
	return packet

}

// ----- CHECK utilities
func check(t *testing.T, want, got interface{}, name string) {
	if !reflect.DeepEqual(want, got) {
		Ko(t, "%s don't match expectations.\nWant: %v\nGot:  %v", name, want, got)
	}
	Ok(t, fmt.Sprintf("Check %s", name))
}

func CheckPushed(t *testing.T, want APacket, got APacket) {
	check(t, want, got, "Pushed")
}

func CheckPersonalized(t *testing.T, want HRegistration, got HRegistration) {
	check(t, want, got, "Personalized")
}

func CheckAcks(t *testing.T, want interface{}, gotItf interface{}) {
	got := gotItf.(struct {
		Ack    *bool
		Packet Packet
	})

	if got.Ack == nil {
		Ko(t, "Invalid ack got: %+v", got)
	}

	switch want.(type) {
	case bool:
		check(t, want.(bool), *(got.Ack), "Acks")
	case Packet:
		check(t, want.(Packet), got.Packet, "Acks")
	default:
		panic("Unexpect ack wanted")
	}
}

func CheckRecipients(t *testing.T, want []Recipient, got []Recipient) {
	check(t, want, got, "Recipients")
}

func CheckSent(t *testing.T, want Packet, got Packet) {
	check(t, want, got, "Sent")
}

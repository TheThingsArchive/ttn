// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package components

import (
	"fmt"
	"testing"
	"time"

	"github.com/TheThingsNetwork/ttn/core"
	. "github.com/TheThingsNetwork/ttn/core/errors"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	. "github.com/TheThingsNetwork/ttn/utils/testing"
	"github.com/brocaar/lorawan"
)

func TestHandleUp(t *testing.T) {
	devices := []device{
		{
			DevAddr: [4]byte{1, 2, 3, 4},
			AppSKey: [16]byte{1, 2, 3, 4, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15},
			NwkSKey: [16]byte{1, 2, 3, 4, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15},
		},
		{
			DevAddr: [4]byte{0, 0, 0, 2},
			AppSKey: [16]byte{1, 2, 3, 4, 4, 5, 6, 7, 8, 9, 8, 8, 8, 8, 8, 8},
			NwkSKey: [16]byte{1, 2, 3, 4, 4, 5, 6, 7, 8, 9, 8, 8, 8, 8, 8, 8},
		},
		{
			DevAddr: [4]byte{0, 0, 0, 3},
			AppSKey: [16]byte{1, 2, 3, 14, 42, 5, 6, 7, 8, 9, 8, 8, 8, 8, 8, 8},
			NwkSKey: [16]byte{1, 2, 3, 42, 14, 5, 6, 7, 8, 9, 8, 8, 8, 8, 8, 8},
		},
		{
			DevAddr: [4]byte{1, 2, 3, 4},
			AppSKey: [16]byte{1, 2, 3, 4, 4, 5, 6, 7, 8, 9, 8, 8, 8, 8, 0xaa, 0xbb},
			NwkSKey: [16]byte{1, 2, 3, 4, 4, 5, 6, 7, 8, 9, 8, 8, 8, 8, 0xaa, 0xbb},
		},
		{
			DevAddr: [4]byte{1, 2, 3, 4},
			AppSKey: [16]byte{1, 2, 3, 4, 4, 5, 6, 37, 8, 9, 8, 8, 8, 8, 0xaa, 0xbb},
			NwkSKey: [16]byte{1, 2, 3, 4, 4, 12, 6, 7, 8, 9, 8, 8, 8, 8, 0xaa, 0xbb},
		},
	}

	applications := map[lorawan.EUI64]application{
		[8]byte{1, 2, 3, 4, 5, 6, 7, 8}: {
			Devices:    []device{devices[0], devices[1]},
			Registered: true,
		},
		[8]byte{0, 9, 8, 7, 6, 5, 4, 3}: {
			Devices:    []device{devices[2], devices[3]},
			Registered: true,
		},
		[8]byte{14, 14, 14, 14, 14, 14, 14, 14}: {
			Devices:    []device{devices[4]},
			Registered: false,
		},
	}

	packets := []packetShape{
		{
			Device: devices[0],
			Data:   "Packet 1 / Dev 1234 / App 12345678",
		},
		{
			Device: devices[0],
			Data:   "Packet 2 / Dev 1234 / App 12345678",
		},
		{
			Device: devices[1],
			Data:   "Packet 1 / Dev 0002 / App 12345678",
		},
		{
			Device: devices[2],
			Data:   "Packet 1 / Dev 0003 / App 09876543",
		},
		{
			Device: devices[3],
			Data:   "Packet 1 / Dev 1234 / App 09876543",
		},
		{
			Device: devices[4],
			Data:   "Packet 1 / Dev 1234 / App 1414141414141414",
		},
	}

	tests := []struct {
		Desc        string
		Schedule    []event
		WantNbAck   int
		WantNbNack  int
		WantPackets map[[12]byte][]string
		WantErrors  []string
	}{
		{
			Desc: "1 packet",
			Schedule: []event{
				event{time.Millisecond * 25, packets[0], nil},
			},
			WantNbAck: 1,
			WantPackets: map[[12]byte][]string{
				[12]byte{1, 2, 3, 4, 5, 6, 7, 8, 1, 2, 3, 4}: {"Packet 1 / Dev 1234 / App 12345678"},
			},
		},
		{
			Desc: "2 packets | same device | same payload | within time frame",
			Schedule: []event{
				event{time.Millisecond * 25, packets[0], nil},
				event{time.Millisecond * 100, packets[0], nil},
			},
			WantNbAck: 2,
			WantPackets: map[[12]byte][]string{
				[12]byte{1, 2, 3, 4, 5, 6, 7, 8, 1, 2, 3, 4}: []string{
					packets[0].Data,
				},
			},
		},
		{
			Desc: "2 packets | same device | same payload | in 2 time frames",
			Schedule: []event{
				event{time.Millisecond * 25, packets[0], nil},
				event{time.Millisecond * 750, packets[0], nil},
			},
			WantNbAck:  1,
			WantNbNack: 1,
			WantErrors: []string{ErrFailedOperation},
			WantPackets: map[[12]byte][]string{
				[12]byte{1, 2, 3, 4, 5, 6, 7, 8, 1, 2, 3, 4}: []string{
					packets[0].Data,
				},
			},
		},
		{
			Desc: "2 packets | same device | different payload | within time frame",
			Schedule: []event{
				event{time.Millisecond * 25, packets[0], nil},
				event{time.Millisecond * 100, packets[1], nil},
			},
			WantNbAck: 2,
			WantPackets: map[[12]byte][]string{
				[12]byte{1, 2, 3, 4, 5, 6, 7, 8, 1, 2, 3, 4}: []string{
					packets[0].Data,
					packets[1].Data,
				},
			},
		},
		{
			Desc: "3 packets | different device | same app | resp same payloads | within time frame",
			Schedule: []event{
				event{time.Millisecond * 25, packets[0], nil},
				event{time.Millisecond * 50, packets[0], nil},
				event{time.Millisecond * 100, packets[2], nil},
			},
			WantNbAck: 3,
			WantPackets: map[[12]byte][]string{
				[12]byte{1, 2, 3, 4, 5, 6, 7, 8, 1, 2, 3, 4}: []string{
					packets[0].Data,
				},
				[12]byte{1, 2, 3, 4, 5, 6, 7, 8, 0, 0, 0, 2}: []string{
					packets[2].Data,
				},
			},
		},
		{
			Desc: "3 packets | different device | different app | resp same payloads | within time frame",
			Schedule: []event{
				event{time.Millisecond * 25, packets[0], nil},
				event{time.Millisecond * 50, packets[2], nil},
				event{time.Millisecond * 100, packets[3], nil},
			},
			WantNbAck: 3,
			WantPackets: map[[12]byte][]string{
				[12]byte{1, 2, 3, 4, 5, 6, 7, 8, 1, 2, 3, 4}: []string{
					packets[0].Data,
				},
				[12]byte{1, 2, 3, 4, 5, 6, 7, 8, 0, 0, 0, 2}: []string{
					packets[2].Data,
				},
				[12]byte{0, 9, 8, 7, 6, 5, 4, 3, 0, 0, 0, 3}: []string{
					packets[3].Data,
				},
			},
		},
		{
			Desc: "3 packets | different device | different app | resp same payloads | within time frame | dev address conflict",
			Schedule: []event{
				event{time.Millisecond * 25, packets[0], nil},
				event{time.Millisecond * 50, packets[2], nil},
				event{time.Millisecond * 100, packets[4], nil},
			},
			WantNbAck: 3,
			WantPackets: map[[12]byte][]string{
				[12]byte{1, 2, 3, 4, 5, 6, 7, 8, 1, 2, 3, 4}: []string{
					packets[0].Data,
				},
				[12]byte{1, 2, 3, 4, 5, 6, 7, 8, 0, 0, 0, 2}: []string{
					packets[2].Data,
				},
				[12]byte{0, 9, 8, 7, 6, 5, 4, 3, 1, 2, 3, 4}: []string{
					packets[4].Data,
				},
			},
		},
		{
			Desc: "1 packet | unknown application",
			Schedule: []event{
				event{time.Millisecond * 25, packets[5], nil},
			},
			WantErrors:  []string{ErrNotFound},
			WantNbNack:  1,
			WantPackets: map[[12]byte][]string{},
		},
	}

	for _, test := range tests {
		// Describe
		Desc(t, test.Desc)

		// Build
		handler := genNewHandler(t, applications)
		genPacketsFromSchedule(&test.Schedule)
		chans := genComChannels("error", "ack", "nack", "packet")

		// Operate
		startSchedule(test.Schedule, handler, chans)

		// Check
		go func() {
			<-time.After(time.Second)
			for _, ch := range chans {
				close(ch)
			}
		}()
		checkChErrors(t, test.WantErrors, chans["error"])
		checkAcks(t, test.WantNbAck, chans["ack"], "ack")
		checkAcks(t, test.WantNbNack, chans["nack"], "nack")
		checkPackets(t, test.WantPackets, chans["packet"])

		if err := handler.db.Close(); err != nil {
			panic(err)
		}
	}
}

type event struct {
	Delay  time.Duration
	Shape  packetShape
	Packet *core.Packet
}

type device struct {
	DevAddr lorawan.DevAddr
	AppSKey lorawan.AES128Key
	NwkSKey lorawan.AES128Key
}

type packetShape struct {
	Device device
	Data   string
}

type application struct {
	Devices    []device
	Registered bool
}

func genPacketsFromSchedule(s *[]event) {
	for i, entry := range *s {
		entry.Packet = new(core.Packet)
		*entry.Packet = genPacketFromShape(entry.Shape)
		(*s)[i] = entry
	}
}

func genPacketFromShape(shape packetShape) core.Packet {
	// Build the macPayload
	macPayload := lorawan.NewMACPayload(true)
	macPayload.FHDR = lorawan.FHDR{DevAddr: shape.Device.DevAddr}
	macPayload.FRMPayload = []lorawan.Payload{&lorawan.DataPayload{
		Bytes: []byte(shape.Data),
	}}
	macPayload.FPort = uint8(1)
	if err := macPayload.EncryptFRMPayload(shape.Device.AppSKey); err != nil {
		panic(err)
	}

	// Build the physicalPayload
	phyPayload := lorawan.NewPHYPayload(true)
	phyPayload.MHDR = lorawan.MHDR{
		MType: lorawan.UnconfirmedDataUp,
		Major: lorawan.LoRaWANR1,
	}
	phyPayload.MACPayload = macPayload
	if err := phyPayload.SetMIC(shape.Device.NwkSKey); err != nil {
		panic(err)
	}
	return core.Packet{
		Payload:  phyPayload,
		Metadata: core.Metadata{},
	}
}

func genNewHandler(t *testing.T, applications map[lorawan.EUI64]application) *Handler {
	ctx := GetLogger(t, "Handler")

	db, err := NewHandlerStorage()
	if err != nil {
		panic(err)
	}

	if err := db.Reset(); err != nil {
		panic(err)
	}

	handler := NewHandler(db, ctx)
	if err != nil {
		panic(err)
	}

	for appEUI, app := range applications {
		if !app.Registered {
			continue
		}
		for _, device := range app.Devices {
			handler.Register(
				core.Registration{
					DevAddr: device.DevAddr,
					Recipient: core.Recipient{
						Address: device.DevAddr,
						Id:      appEUI,
					},
					Options: struct {
						AppSKey lorawan.AES128Key
						NwkSKey lorawan.AES128Key
					}{
						AppSKey: device.AppSKey,
						NwkSKey: device.NwkSKey,
					},
				},
				voidAckNacker{},
			)
		}
	}
	return handler
}

type voidAckNacker struct{}

func (v voidAckNacker) Ack(packets *core.Packet) error {
	return nil
}
func (v voidAckNacker) Nack() error {
	return nil
}

func genComChannels(names ...string) map[string]chan interface{} {
	chans := make(map[string]chan interface{})
	for _, name := range names {
		chans[name] = make(chan interface{}, 50)
	}
	return chans
}

func startSchedule(s []event, handler *Handler, chans map[string]chan interface{}) {
	mockAn := chanAckNacker{AckChan: chans["ack"], NackChan: chans["nack"]}
	mockAdapter := chanAdapter{PktChan: chans["packet"]}

	for _, ev := range s {
		<-time.After(ev.Delay)
		go func(ev event) {
			err := handler.HandleUp(*ev.Packet, mockAn, mockAdapter)
			if err != nil {
				chans["error"] <- err
			}
		}(ev)
	}
}

type chanAckNacker struct {
	AckChan  chan interface{}
	NackChan chan interface{}
}

func (an chanAckNacker) Ack(packets *core.Packet) error {
	an.AckChan <- true
	return nil
}

func (an chanAckNacker) Nack() error {
	an.NackChan <- true
	return nil
}

type chanAdapter struct {
	PktChan chan interface{}
}

func (a chanAdapter) Send(p core.Packet, r ...core.Recipient) (core.Packet, error) {
	a.PktChan <- struct {
		Packet    core.Packet
		Recipient []core.Recipient
	}{
		Packet:    p,
		Recipient: r,
	}
	return core.Packet{}, nil
}

func (a chanAdapter) Next() (core.Packet, core.AckNacker, error) {
	panic("Not Expected")
}

func (a chanAdapter) NextRegistration() (core.Registration, core.AckNacker, error) {
	panic("Not Expected")
}

func checkChErrors(t *testing.T, want []string, got chan interface{}) {
	nb := 0
outer:
	for gotErr := range got {
		for _, wantErr := range want {
			if got != nil && wantErr == gotErr.(errors.Failure).Nature {
				nb += 1
				continue outer
			}
		}
		Ko(t, "Got error [%v] but was only expecting: [%v]", gotErr, want)
		return
	}
	if nb != len(want) {
		Ko(t, "Expected %d error(s) but got only %d", len(want), nb)
		return
	}
	Ok(t, "Check errors")
}

func checkAcks(t *testing.T, want int, got chan interface{}, kind string) {
	nb := 0
	for {
		a, ok := <-got
		if !ok && a == nil {
			break
		}
		nb += 1
	}

	if nb != want {
		Ko(t, "Expected %d %s(s) but got %d", want, kind, nb)
		return
	}
	Ok(t, fmt.Sprintf("Check %s", kind))
}

func checkPackets(t *testing.T, want map[[12]byte][]string, got chan interface{}) {
	nb := 0
outer:
	for x := range got {
		msg := x.(struct {
			Packet    core.Packet
			Recipient []core.Recipient
		})

		if len(msg.Recipient) != 1 {
			Ko(t, "Expected exactly one recipient but got %d", len(msg.Recipient))
			return
		}

		appEUI := msg.Recipient[0].Id.(lorawan.EUI64)
		devAddr, err := msg.Packet.DevAddr()
		if err != nil {
			Ko(t, "Unexpected error: %v", err)
			return
		}

		var id [12]byte
		copy(id[:8], appEUI[:])
		copy(id[8:], devAddr[:])

		wantData, ok := want[id]
		if !ok {
			Ko(t, "Received unexpected packet for app %v and from node %v", appEUI, devAddr)
			return
		}

		macPayload := msg.Packet.Payload.MACPayload.(*lorawan.MACPayload)
		if len(macPayload.FRMPayload) != 1 {
			Ko(t, "Invalid macpayload in received packet from node %v", devAddr)
			return
		}

		gotData := string(macPayload.FRMPayload[0].(*lorawan.DataPayload).Bytes)

		for _, want := range wantData {
			if want == gotData {
				nb += 1
				continue outer
			}
		}

		Ko(t, "Received data don't match expectations.\nWant: %v\nGot:  %s", wantData, gotData)
		return
	}

	if nb != len(want) {
		Ko(t, "Handler sent %d packet(s) whereas %d were/was expected", nb, len(want))
		return
	}

	Ok(t, "Check packets")
}

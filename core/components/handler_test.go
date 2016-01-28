// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package components

import (
	"fmt"
	"testing"
	"time"

	"github.com/TheThingsNetwork/ttn/core"
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
	}

	applications := map[lorawan.EUI64]application{
		[8]byte{1, 2, 3, 4, 5, 6, 7, 8}: {
			Devices:    []device{devices[0]},
			Registered: true,
		},
	}

	packets := []packetShape{
		{
			Device: devices[0],
			Data:   "Packet 1 / Dev 1234 / App 12345678",
		},
	}

	tests := []struct {
		Desc        string
		Schedule    []event
		WantNbAck   int
		WantNbNack  int
		WantPackets map[[12]byte]string
		WantErrors  []error
	}{
		{
			Desc: "Easy - one packet",
			Schedule: []event{
				event{time.Millisecond * 25, packets[0], nil},
			},
			WantNbAck: 1,
			WantPackets: map[[12]byte]string{
				[12]byte{1, 2, 3, 4, 5, 6, 7, 8, 1, 2, 3, 4}: "Packet 1 / Dev 1234 / App 12345678",
			},
		},
		{
			Desc: "Two packets from the same device within the time frame",
			Schedule: []event{
				event{time.Millisecond * 25, packets[0], nil},
				event{time.Millisecond * 100, packets[0], nil},
			},
			WantNbAck: 2,
			WantPackets: map[[12]byte]string{
				[12]byte{1, 2, 3, 4, 5, 6, 7, 8, 1, 2, 3, 4}: "Packet 1 / Dev 1234 / App 12345678",
			},
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
		// Build the macPayload
		macPayload := lorawan.NewMACPayload(true)
		macPayload.FHDR = lorawan.FHDR{DevAddr: entry.Shape.Device.DevAddr}
		macPayload.FRMPayload = []lorawan.Payload{&lorawan.DataPayload{
			Bytes: []byte(entry.Shape.Data),
		}}
		macPayload.FPort = uint8(1)
		if err := macPayload.EncryptFRMPayload(entry.Shape.Device.AppSKey); err != nil {
			panic(err)
		}

		// Build the physicalPayload
		phyPayload := lorawan.NewPHYPayload(true)
		phyPayload.MHDR = lorawan.MHDR{
			MType: lorawan.UnconfirmedDataUp,
			Major: lorawan.LoRaWANR1,
		}
		phyPayload.MACPayload = macPayload
		if err := phyPayload.SetMIC(entry.Shape.Device.NwkSKey); err != nil {
			panic(err)
		}
		entry.Packet = &core.Packet{
			Payload:  phyPayload,
			Metadata: core.Metadata{},
		}
		(*s)[i] = entry
	}
}

func genNewHandler(t *testing.T, applications map[lorawan.EUI64]application) *Handler {
	ctx := GetLogger(t, "Handler")
	handler, err := NewHandler(newHandlerDB(), ctx)
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

func (v voidAckNacker) Ack(packets ...core.Packet) error {
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

func (an chanAckNacker) Ack(packets ...core.Packet) error {
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

func checkChErrors(t *testing.T, want []error, got chan interface{}) {
	nb := 0
outer:
	for gotErr := range got {
		for wantErr := range want {
			if wantErr == gotErr {
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

func checkPackets(t *testing.T, want map[[12]byte]string, got chan interface{}) {
	nb := 0
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
			Ko(t, "Received unexpected packet for app %x and from node %x", appEUI, devAddr)
			return
		}

		macPayload := msg.Packet.Payload.MACPayload.(*lorawan.MACPayload)
		if len(macPayload.FRMPayload) != 1 {
			Ko(t, "Invalid macpayload in received packet from node %x", devAddr)
			return
		}

		gotData := string(macPayload.FRMPayload[0].(*lorawan.DataPayload).Bytes)
		if wantData != gotData {
			Ko(t, "Received data don't match expectations.\nWant: %s\nGot:  %s", wantData, gotData)
			return
		}

		nb += 1
	}

	if nb != len(want) {
		Ko(t, "Handler sent %d packet(s) whereas %d were/was expected", nb, len(want))
		return
	}

	Ok(t, "Check packets")
}

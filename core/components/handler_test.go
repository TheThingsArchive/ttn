// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package components

import (
	//	"reflect"
	"testing"
	"time"

	"github.com/TheThingsNetwork/ttn/core"
	//	"github.com/TheThingsNetwork/ttn/utils/pointer"
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
			Devices:    []device{},
			Registered: true,
		},

		[8]byte{9, 10, 11, 12, 13, 14, 15, 16}: {
			Devices:    []device{},
			Registered: true,
		},

		[8]byte{1, 1, 2, 2, 3, 3, 4, 4}: {
			Devices:    []device{},
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
			Data:   "Packet 1 / Dev 1234 / App 12345678",
		},
	}

	tests := []struct {
		Schedule    schedule
		WantAck     map[[4]byte]bool
		WantPackets map[[12]byte]string
		WantError   error
	}{
		{
			Schedule: schedule{
				{time.Millisecond * 25, packets[0], nil},
			},
			WantAck: map[[4]byte]bool{
				[4]byte{1, 2, 3, 4}: true,
			},
			WantError: nil,
			WantPackets: map[[12]byte]string{
				[12]byte{1, 2, 3, 4, 5, 6, 7, 8, 1, 2, 3, 4}: "Packet 1 / Dev 1234 / App 12345678",
			},
		},
	}

	for _, test := range tests {
		// Describe

		// Build
		handler := genNewHandler(t, applications)
		genPacketsFromSchedule(&test.Schedule)
		chans := genComChannels("error", "ack", "nack", "packet")

		// Operate
		go startSchedule(test.Schedule, handler, chans)

		// Check
	}
}

type schedule []struct {
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

func genPacketsFromSchedule(s *schedule) {
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
		chans[name] = make(chan interface{})
	}
	return chans
}

func startSchedule(s schedule, handler *Handler, chans map[string]chan interface{}) {
	mockAn := chanAckNacker{AckChan: chans["ack"], NackChan: chans["nack"]}
	mockAdapter := chanAdapter{PktChan: chans["packet"]}

	for _, event := range s {
		<-time.After(event.Delay)
		go func() {
			err := handler.HandleUp(*event.Packet, mockAn, mockAdapter)
			if err != nil {
				chans["error"] <- err
			}
		}()
	}
}

type chanAckNacker struct {
	AckChan  chan interface{}
	NackChan chan interface{}
}

func (an chanAckNacker) Ack(packets ...core.Packet) error {
	if len(packets) == 0 {
		an.AckChan <- true
		return nil
	}

	for _, p := range packets {
		an.AckChan <- p
	}
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

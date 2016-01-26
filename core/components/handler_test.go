// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package components

import (
	//	"reflect"
	"testing"
	"time"

	//	"github.com/TheThingsNetwork/ttn/core"
	//	"github.com/TheThingsNetwork/ttn/utils/pointer"
	. "github.com/TheThingsNetwork/ttn/utils/testing"
	"github.com/brocaar/lorawan"
)

func TestHandleUp(t *testing.T) {
	applications := make(map[lorawan.EUI64]lorawan.AES128Key)
	applications[lorawan.EUI64([8]byte{1, 2, 3, 4, 5, 6, 7, 8})] = lorawan.AES128Key([16]byte{1, 2, 3, 4, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15})
	applications[lorawan.EUI64([8]byte{9, 10, 11, 12, 13, 14, 15, 16})] = lorawan.AES128Key([16]byte{15, 14, 13, 12, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1})
	applications[lorawan.EUI64([8]byte{1, 1, 2, 2, 3, 3, 4, 4})] = lorawan.AES128Key([16]byte{1, 2, 3, 4, 5, 6, 7, 8, 1, 2, 3, 4, 5, 6, 7, 8})

	packets := []plannedPacket{
		{
			AppEUI:  lorawan.EUI64([8]byte{1, 2, 3, 4, 5, 6, 7, 8}),
			DevAddr: lorawan.DevAddr([4]byte{1, 2, 3, 4}),
			Data:    "Packet 1 / Dev 1234 / App 12345678",
		},
		{
			AppEUI:  lorawan.EUI64([8]byte{1, 2, 3, 4, 5, 6, 7, 8}),
			DevAddr: lorawan.DevAddr([4]byte{1, 2, 3, 4}),
			Data:    "Packet 1 / Dev 1234 / App 12345678",
		},
	}

	tests := []struct {
		Schedule    []schedule
		WantAck     map[[4]byte]bool
		WantPackets map[[12]byte]string
		WantError   error
	}{
		{
			Schedule: []schedule{
				{time.Millisecond * 25, packets[0]},
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

		// Operate

		// Check
	}
}

type schedule struct {
	Delay  time.Duration
	Packet plannedPacket
}

type plannedPacket struct {
	AppEUI  lorawan.EUI64
	DevAddr lorawan.DevAddr
	Data    string
}

func genNewHandler(t *testing.T, applications map[lorawan.EUI64]lorawan.AES128Key) Handler {
	ctx := GetLogger(t, "Handler")
	handler, err := NewHandler(newHandlerDB(), ctx)
	if err != nil {
		panic(err)
	}
}

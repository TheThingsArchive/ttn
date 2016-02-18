// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package components

import (
	"reflect"
	"testing"
	"time"

	"github.com/TheThingsNetwork/ttn/core"
	. "github.com/TheThingsNetwork/ttn/utils/testing"
	"github.com/brocaar/lorawan"
)

func TestRouterHandleUp(t *testing.T) {
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
	}

	recipients := []core.Recipient{
		{Address: "Recipient1", Id: ""},
		{Address: "Recipient2", Id: ""},
	}

	tests := []struct {
		Desc            string
		KnownRecipients map[[4]byte]core.Recipient
		Packet          packetShape
		WantRecipients  []core.Recipient
		WantError       *string
	}{
		{
			Desc: "0 known | Send #0",
			Packet: packetShape{
				Device: devices[0],
				Data:   "MyData",
			},
			WantRecipients: nil,
			WantError:      nil,
		},
		{
			Desc: "Know #0 | Send #0",
			KnownRecipients: map[[4]byte]core.Recipient{
				devices[0].DevAddr: recipients[0],
			},
			Packet: packetShape{
				Device: devices[0],
				Data:   "MyData",
			},
			WantRecipients: []core.Recipient{recipients[0]},
			WantError:      nil,
		},
		{
			Desc: "Know #1 | Send #0",
			KnownRecipients: map[[4]byte]core.Recipient{
				devices[1].DevAddr: recipients[0],
			},
			Packet: packetShape{
				Device: devices[0],
				Data:   "MyData",
			},
			WantRecipients: nil,
			WantError:      nil,
		},
	}

	for _, test := range tests {
		// Describe
		Desc(t, test.Desc)

		// Build
		router := genNewRouter(t, test.KnownRecipients)
		packet := genPacketFromShape(test.Packet)

		// Operate
		recipients, err := handleRouterUp(router, packet)

		// Check
		checkErrors(t, test.WantError, err)
		checkRecipients(t, test.WantRecipients, recipients)

		if err := router.db.Close(); err != nil {
			panic(err)
		}
	}
}

type routerRecipient struct {
	Address string
}

// ----- BUILD utilities
func genNewRouter(t *testing.T, knownRecipients map[[4]byte]core.Recipient) *Router {
	ctx := GetLogger(t, "Router")

	db, err := NewRouterStorage(time.Hour * 8)
	if err != nil {
		panic(err)
	}

	if err := db.Reset(); err != nil {
		panic(err)
	}

	router := NewRouter(db, ctx)
	if err != nil {
		panic(err)
	}

	for devAddr, recipient := range knownRecipients {
		err := router.Register(core.Registration{
			DevAddr:   lorawan.DevAddr(devAddr),
			Recipient: recipient,
		}, voidAckNacker{})
		if err != nil {
			panic(err)
		}
	}

	return router
}

// ----- OPERATE utilities
func handleRouterUp(router core.Router, packet core.Packet) ([]core.Recipient, error) {
	adapter := &routerAdapter{}
	err := router.HandleUp(packet, voidAckNacker{}, adapter)
	return adapter.Recipients, err
}

type routerAdapter struct {
	Recipients []core.Recipient
}

func (a *routerAdapter) Send(p core.Packet, r ...core.Recipient) (core.Packet, error) {
	a.Recipients = r
	return core.Packet{}, nil
}

func (a *routerAdapter) Next() (core.Packet, core.AckNacker, error) {
	panic("Unexpected call to Next()")
}

func (a *routerAdapter) NextRegistration() (core.Registration, core.AckNacker, error) {
	panic("Unexpected call to NextRegistration")
}

// ----- Check utilities
func checkRecipients(t *testing.T, want []core.Recipient, got []core.Recipient) {
	if !reflect.DeepEqual(want, got) {
		Ko(t, "Contacted recipients don't match expectations.\nWant: %v\nGot:  %v", want, got)
		return
	}
	Ok(t, "Check recipients")
}

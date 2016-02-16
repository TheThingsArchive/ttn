// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package components

import (
	"testing"
	"time"

	"github.com/TheThingsNetwork/ttn/core"
	. "github.com/TheThingsNetwork/ttn/utils/testing"
)

func TestRouterHandleUp(t *testing.T) {
	tests := []struct {
		Desc           string
		Packet         packetShape
		WantRecipients []core.Recipient
		WantError      error
	}{}

	for _, test := range tests {
		// Describe
		Desc(t, test.Desc)

		// Build
		router := genNewRouter(t)
		packet := genPacketFromShape(test.Packet)

		// Operate
		recipients, err := handleRouterUp(router, packet)

		// Check
		checkErrors(t, test.WantError, err)
		checkRecipients(t, test.WantRecipients, recipients)
	}
}

type routerRecipient struct {
	Address string
}

// ----- BUILD utilities
func genNewRouter(t *testing.T) core.Router {
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

	return router
}

// ----- OPERATE utilities
func handleRouterUp(router core.Router, packet core.Packet) ([]core.Recipient, error) {
	an := voidAckNacker{}
	adapter := routerAdapter{}
	err := router.HandleUp(packet, an, adapter)
	return adapter.Recipients, err
}

type routerAdapter struct {
	Recipients []core.Recipient
}

func (a routerAdapter) Send(p core.Packet, r ...core.Recipient) (core.Packet, error) {
	a.Recipients = r
	return core.Packet{}, nil
}

func (a routerAdapter) Next() (core.Packet, core.AckNacker, error) {
	panic("Unexpected call to Next()")
}

func (a routerAdapter) NextRegistration() (core.Registration, core.AckNacker, error) {
	panic("Unexpected call to NextRegistration")
}

// ----- Check utilities
func checkRecipients(t *testing.T, want []core.Recipient, got []core.Recipient) {

}

// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package components

import (
	"github.com/thethingsnetwork/core"
)

type Broker struct{}

/* Scenario
broker := components.NewBroker()
upAdapter := adapters.brk_hdl_http
downAdapter := adapters.brk_rtr_http

downAdapter.Start("3000")
upAdapter.Start("8080")

// Handle registration coming from Handler
go func() {
	for {
		handler, devAddr, nwskey, err := upAdapter.NextRegistration()
		if err != nil {
			// Do some error handling
		}
		err = broker.Register(handler, devAddr, nwskey)
		if err != nil {
			// Do some error handling
		}
	}
}()

// Handle response packet coming from Handler
go func() {
	for {
		packet, an, error := upAdapter.Next()
		if err != nil {
			// Do some error handling
		}
		err = broker.Handle(packet, an)
		if err != nil {
			// Do some error handling
		}
	}
}

// Handle packet from router
go func() {
	for {
		packet, an, error := downAdapter.Next()
		if err != nil {
			// Do some error handling
		}
		err = broker.Handle(packet, an)
		if err != nil {
			// Do some error handling
		}
	}
}
*/

func NewBroker() {

}

func (b *Broker) NextUp() (*core.Packet, error) {
	return nil, nil
}

func (b *Broker) NextDown() (*core.Packet, error) {
	return nil, nil
}

func (b *Broker) Handle(p core.Packet, an core.AckNacker) error {
	return nil
}

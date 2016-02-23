// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package mqtt

import (
	MQTT "git.eclipse.org/gitroot/paho/org.eclipse.paho.mqtt.golang.git"

	//. "github.com/TheThingsNetwork/ttn/core/errors"
	core "github.com/TheThingsNetwork/ttn/refactor"
	//"github.com/TheThingsNetwork/ttn/utils/errors"
	//"github.com/TheThingsNetwork/ttn/utils/stats"
	"github.com/apex/log"
	//"github.com/brocaar/lorawan"
)

// Adapter type materializes an mqtt adapter which implements the basic mqtt protocol
type Adapter struct {
	ctx           log.Interface
	packets       chan PktReq // Channel used to "transforms" incoming request to something we can handle concurrently
	registrations chan RegReq // Incoming registrations
}

// Handler defines topic-specific handler.
type Handler interface {
	Topic() string
	Handle(client *MQTT.Client, chpkt chan<- PktReq, chreg chan<- RegReq, msg MQTT.Message)
}

// Message sent through the response channel of a pktReq or regReq
type MsgRes []byte // The response content.

// Message sent through the packets channel when an incoming request arrives
type PktReq struct {
	Packet []byte      // The actual packet that has been parsed
	Chresp chan MsgRes // A response channel waiting for an success or reject confirmation
}

// Message sent through the registration channel when an incoming registration arrives
type RegReq struct {
	Registration core.Registration
	Chresp       chan MsgRes
}

// NewAdapter constructs and allocates a new mqtt adapter
func NewAdapter(ctx log.Interface) (*Adapter, error) {
	adapter := &Adapter{
		ctx:           ctx,
		packets:       make(chan PktReq),
		registrations: make(chan RegReq),
	}

	return adapter, nil
}

// Send implements the core.Adapter interface
func (a *Adapter) Send(p core.Packet, recipients ...core.Recipient) ([]byte, error) {
	return nil, nil
}

// Next implements the core.Adapter interface
func (a *Adapter) Next() ([]byte, core.AckNacker, error) {
	return nil, nil, nil
}

// NextRegistration implements the core.Adapter interface. Not implemented for this adapters.
func (a *Adapter) NextRegistration() (core.Registration, core.AckNacker, error) {
	return nil, nil, nil
}

// Bind registers a handler to a specific endpoint
func (a *Adapter) Bind(h Handler) {
}

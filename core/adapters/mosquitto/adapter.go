// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package mosquitto

import (
	"github.com/TheThingsNetwork/ttn/core"
	"github.com/apex/log"
	"github.com/brocaar/lorawan"
)

type Adapter struct {
	ctx log.Interface
}

type PersonnalizedActivation struct {
	DevAddr lorawan.DevAddr
	NwkSKey lorawan.AES128Key
	AppSKey lorawan.AES128Key
}

const (
	TOPIC_ACTIVATIONS string = "activations"
	TOPIC_UPLINK             = "up"
	TOPIC_DOWNLINK           = "down"
	RESOURCE                 = "devices"
)

// NewAdapter constructs a new mqtt adapter
func NewAdapter(mqttBroker string, ctx log.Interface) (*Adapter, error) {
	return nil, nil
}

// Send implements the core.Adapter interface
func (a *Adapter) Send(p core.Packet, r ...core.Recipient) (core.Packet, error) {
	return core.Packet{}, nil
}

// Next implements the core.Adapter interface
func (a *Adapter) Next() (core.Packet, core.AckNacker, error) {
	return core.Packet{}, nil, nil
}

// NextRegistration implements the core.Adapter interface
func (a *Adapter) NextRegistration() (core.Registration, core.AckNacker, error) {
	return core.Registration{}, nil, nil
}

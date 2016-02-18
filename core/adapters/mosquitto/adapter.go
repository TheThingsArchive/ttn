// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package mosquitto

import (
	"github.com/TheThingsNetwork/ttn/core"
	"github.com/apex/log"
)

type Adapter struct {
	ctx log.Interface
}

// NewAdapter constructs a new mqtt adapter
func NewAdapter() (*Adapter, error) {
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

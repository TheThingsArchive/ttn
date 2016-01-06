// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package brk_hdl_http

import (
	"github.com/thethingsnetwork/core"
	"github.com/thethingsnetwork/core/lorawan"
)

type Adapter struct {
}

func NewAdapter(port string) (*Adapter, error) {
	return nil, nil
}

// Send implements the core.Adapter interface
func (a *Adapter) Send(p core.Packet, an core.AckNacker) error {
	return nil
}

// Next implements the core.Adapter inerface
func (a *Adapter) Next() (core.Packet, core.AckNacker, error) {
	return core.Packet{}, nil, nil
}

// NextRegistration implements the core.BrkHdlAdapter interface
func (a *Adapter) NextRegistration() (core.Recipient, lorawan.DevAddr, lorawan.AES128Key, error) {
	return core.Recipient{}, lorawan.DevAddr{}, lorawan.AES128Key{}, nil
}

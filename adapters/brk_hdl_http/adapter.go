// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package brk_hdl_http

import (
	"fmt"
	"github.com/thethingsnetwork/core"
	"github.com/thethingsnetwork/core/lorawan"
)

var ErrInvalidPort = fmt.Errorf("The given port is invalid")

type Adapter struct {
}

func NewAdapter(port uint) (*Adapter, error) {
	if port == 0 {
		return nil, ErrInvalidPort
	}

	a := Adapter{}
	return &a, nil
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

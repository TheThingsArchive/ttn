// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package brk_hdl_http

import (
	"fmt"
	"github.com/thethingsnetwork/core"
	"github.com/thethingsnetwork/core/utils/log"
)

var ErrInvalidPort = fmt.Errorf("The given port is invalid")
var ErrConnectionLost = fmt.Errorf("The connection has been lost")

type Adapter struct {
	loggers       []log.Logger // 0 to several loggers to get feedback from the Adapter.
	registrations chan regReq  // Communication dedicated to incoming registration from handlers
}

// NewAdapter constructs and allocate a new Broker <-> Handler http adapter
func NewAdapter(port uint, loggers ...log.Logger) (*Adapter, error) {
	if port == 0 {
		return nil, ErrInvalidPort
	}

	a := Adapter{
		registrations: make(chan regReq),
		loggers:       loggers,
	}

	go a.listenRegistration(port)

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
func (a *Adapter) NextRegistration() (core.Registration, core.AckNacker, error) {
	request := <-a.registrations
	return request.Registration, regAckNacker{response: request.response}, nil
}

func (a *Adapter) log(format string, i ...interface{}) {
	for _, logger := range a.loggers {
		logger.Log(format, i...)
	}
}

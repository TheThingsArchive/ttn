// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package router

import (
	. "github.com/TheThingsNetwork/ttn/refactor"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	"github.com/TheThingsNetwork/ttn/utils/stats"
	"github.com/apex/log"
)

type component struct {
	Storage
	ctx log.Interface
}

// Register implements the core.Component interface
func (r component) Register(reg Registration, an AckNacker) (err error) {
	defer ensureAckNack(an, nil, &err)
	stats.MarkMeter("router.registration.in")
	r.ctx.Debug("Handling registration")

	if err := r.Store(reg); err != nil {
		return errors.New(errors.Operational, err)
	}
	return nil
}

// HandleUp implements the core.Component interface
func (r component) HandleUp(p Packet, an AckNacker, up Adapter) error {
	return nil
}

// HandleDown implements the core.Component interface
func (r component) HandleDown(p Packet, an AckNacker, up Adapter) error {
	return errors.New(errors.Implementation, "Handle down not implemented on router")
}

func ensureAckNack(an AckNacker, ack *Packet, err *error) {
	if err != nil && *err != nil {
		an.Nack()
	} else {
		var p Packet
		if ack != nil {
			p = *ack
		}
		an.Ack(p)
	}
}

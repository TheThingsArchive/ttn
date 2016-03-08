// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package udp

import (
	"time"

	"github.com/TheThingsNetwork/ttn/core"
	"github.com/TheThingsNetwork/ttn/utils/errors"
)

// udpAckNacker represents an AckNacker for a udp adapter
type udpAckNacker struct {
	Chresp chan<- MsgRes
}

// Ack implements the core.Adapter interface
func (an udpAckNacker) Ack(p core.Packet) error {
	defer close(an.Chresp)
	if p == nil {
		return nil
	}
	data, err := p.MarshalBinary()
	if err != nil {
		return errors.New(errors.Structural, err)
	}
	select {
	case an.Chresp <- MsgRes(data):
		return nil
	case <-time.After(time.Millisecond * 50):
		return errors.New(errors.Operational, "Unable to send ack")
	}
}

// Ack implements the core.Adapter interface
func (an udpAckNacker) Nack(err error) error {
	defer close(an.Chresp)
	return nil
}

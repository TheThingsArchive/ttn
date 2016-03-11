// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package mqtt

import (
	"time"

	"github.com/TheThingsNetwork/ttn/core"
	"github.com/TheThingsNetwork/ttn/utils/errors"
)

// mqttAckNacker implements the core.AckNacker interface
type mqttAckNacker struct {
	Chresp chan<- MsgRes // A channel dedicated to send back a response
}

// Ack implements the core.AckNacker interface
func (an mqttAckNacker) Ack(p core.Packet) error {
	if an.Chresp == nil || p == nil {
		return nil
	}
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
		return errors.New(errors.Operational, "No response was given to the acknacker")
	}
}

// Nack implements the core.AckNacker interface
func (an mqttAckNacker) Nack(err error) error {
	if an.Chresp == nil {
		return nil
	}
	defer close(an.Chresp)
	return nil
}

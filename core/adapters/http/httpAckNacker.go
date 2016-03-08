// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package http

import (
	"net/http"
	"time"

	"github.com/TheThingsNetwork/ttn/core"
	"github.com/TheThingsNetwork/ttn/utils/errors"
)

// httpAckNacker implements the AckNacker interface
type httpAckNacker struct {
	Chresp chan<- MsgRes // A channel dedicated to send back a response
}

// Ack implements the core.AckNacker interface
func (an httpAckNacker) Ack(p core.Packet) error {
	if an.Chresp == nil {
		return nil
	}
	defer close(an.Chresp)

	var data []byte
	if p != nil {
		var err error
		data, err = p.MarshalBinary()
		if err != nil {
			return errors.New(errors.Structural, err)
		}
	}

	select {
	case an.Chresp <- MsgRes{
		StatusCode: http.StatusOK,
		Content:    data,
	}:
		return nil
	case <-time.After(time.Millisecond * 50):
		return errors.New(errors.Operational, "No response was given to the acknacker")
	}
}

// Nack implements the core.AckNacker interface
func (an httpAckNacker) Nack(err error) error {
	if an.Chresp == nil {
		return nil
	}
	defer close(an.Chresp)

	select {
	case an.Chresp <- MsgRes{
		StatusCode: http.StatusNotFound,
		Content:    []byte(errors.Structural),
	}:
		return nil
	case <-time.After(time.Millisecond * 50):
		return errors.New(errors.Operational, "No response was given to the acknacker")
	}
}

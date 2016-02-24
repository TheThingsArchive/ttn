// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package http

import (
	"net/http"
	"time"

	core "github.com/TheThingsNetwork/ttn/refactor"
	"github.com/TheThingsNetwork/ttn/utils/errors"
)

// An ackNacker for http registrations
type regAckNacker struct {
	Chresp chan<- MsgRes // A channel dedicated to send back a response
}

// Ack implements the core.Acker interface
func (r regAckNacker) Ack(p core.Packet) error {
	if r.Chresp == nil {
		return nil
	}
	defer close(r.Chresp)

	select {
	case r.Chresp <- MsgRes{StatusCode: http.StatusAccepted}:
		return nil
	case <-time.After(time.Millisecond * 50):
		return errors.New(errors.Operational, "No response was given to the acknacker")
	}
}

// Nack implements the core.Nacker interface
func (r regAckNacker) Nack() error {
	if r.Chresp == nil {
		return nil
	}
	select {
	case r.Chresp <- MsgRes{
		StatusCode: http.StatusConflict,
		Content:    []byte(errors.New(errors.Structural, "Unable to register device").Error()),
	}:
		return nil
	case <-time.After(time.Millisecond * 50):
		return errors.New(errors.Operational, "No response was given to the acknacker")
	}
}

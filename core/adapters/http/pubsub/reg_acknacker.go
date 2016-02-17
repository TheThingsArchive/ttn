// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package pubsub

import (
	"net/http"
	"time"

	"github.com/TheThingsNetwork/ttn/core"
	. "github.com/TheThingsNetwork/ttn/core/errors"
	"github.com/TheThingsNetwork/ttn/utils/errors"
)

type regAckNacker struct {
	response chan regRes // A channel dedicated to send back a response
}

// Ack implements the core.Acker interface
func (r regAckNacker) Ack(p *core.Packet) error {
	select {
	case r.response <- regRes{statusCode: http.StatusAccepted}:
		return nil
	case <-time.After(time.Millisecond * 50):
		return errors.NewFailure(ErrConnectionLost, "No response was given to the acknacker")
	}
}

// Nack implements the core.Nacker interface
func (r regAckNacker) Nack() error {
	select {
	case r.response <- regRes{
		statusCode: http.StatusConflict,
		content:    []byte("Unable to register the given device"),
	}:
		return nil
	case <-time.After(time.Millisecond * 50):
		return errors.NewFailure(ErrConnectionLost, "No response was given to the acknacker")
	}
}

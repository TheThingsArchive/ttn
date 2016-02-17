// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package http

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/TheThingsNetwork/ttn/core"
	. "github.com/TheThingsNetwork/ttn/core/errors"
	"github.com/TheThingsNetwork/ttn/utils/errors"
)

// packetAckNacker implements the AckNacker interface
type packetAckNacker struct {
	response chan pktRes // A channel dedicated to send back a response
}

// Ack implements the core.AckNacker interface
func (an packetAckNacker) Ack(p *core.Packet) error {
	defer close(an.response)
	if p == nil {
		an.response <- pktRes{statusCode: http.StatusOK}
		return nil
	}

	raw, err := json.Marshal(*p)
	if err != nil {
		return errors.NewFailure(ErrInvalidPacket, err)
	}

	select {
	case an.response <- pktRes{statusCode: http.StatusOK, content: raw}:
		return nil
	case <-time.After(time.Millisecond * 50):
		return errors.NewFailure(ErrConnectionLost, "No response was given to the acknacker")
	}
}

// Nack implements the core.AckNacker interface
func (an packetAckNacker) Nack() error {
	select {
	case an.response <- pktRes{
		statusCode: http.StatusNotFound,
		content:    []byte(`{"message":"Not in charge of the associated device"}`),
	}:
	case <-time.After(time.Millisecond * 50):
		return errors.NewFailure(ErrConnectionLost, "No response was given to the acknacker")
	}
	return nil
}

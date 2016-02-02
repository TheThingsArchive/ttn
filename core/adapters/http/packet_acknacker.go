// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/TheThingsNetwork/ttn/core"
)

var ErrConnectionLost = fmt.Errorf("Connection has been lost")
var ErrInvalidArguments = fmt.Errorf("Invalid arguments supplied")

// packetAckNacker implements the AckNacker interface
type packetAckNacker struct {
	response chan pktRes // A channel dedicated to send back a response
}

// Ack implements the core.AckNacker interface
func (an packetAckNacker) Ack(p ...core.Packet) error {
	if len(p) > 1 {
		return ErrInvalidArguments
	}
	var raw []byte
	if len(p) == 1 {
		var err error
		raw, err = json.Marshal(p[0])
		if err != nil {
			return err
		}
	}

	select {
	case an.response <- pktRes{statusCode: http.StatusOK, content: raw}:
		return nil
	case <-time.After(time.Millisecond * 50):
		return ErrConnectionLost
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
		return ErrConnectionLost
	}
	return nil
}

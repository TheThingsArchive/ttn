// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package parser

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/TheThingsNetwork/ttn/core"
	. "github.com/TheThingsNetwork/ttn/core/errors"
	"github.com/TheThingsNetwork/ttn/utils/errors"
)

// Parser gives a flexible way of parsing a request into a packet.
type PacketParser interface {
	// Parse transforms a given http request into a Packet. The error handling is not under its
	// responsibility. The parser can expect any query param, http method or header it needs.
	Parse(req *http.Request) (core.Packet, error)
}

// JSONPacket defines a parser for packet sent as JSON payload
type JSON struct{}

// Parse implements the PacketParser interface
func (p JSON) Parse(req *http.Request) (core.Packet, error) {
	// Check Content-type
	if req.Header.Get("Content-Type") != "application/json" {
		return core.Packet{}, errors.NewFailure(ErrInvalidRequest, "Received invalid content-type in request")
	}

	// Check configuration in body
	body := make([]byte, req.ContentLength)
	n, err := req.Body.Read(body)
	if err != nil && err != io.EOF {
		return core.Packet{}, errors.NewFailure(ErrInvalidRequest, err)
	}
	packet := new(core.Packet)
	if err := json.Unmarshal(body[:n], packet); err != nil {
		return core.Packet{}, errors.NewFailure(ErrInvalidRequest, err)
	}

	return *packet, nil
}

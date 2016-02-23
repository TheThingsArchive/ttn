// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package handlers

import (
	//"encoding/hex"
	//"encoding/json"
	//"io"
	"net/http"
	//"regexp"
	//"strings"

	//. "github.com/TheThingsNetwork/ttn/core/errors"
	core "github.com/TheThingsNetwork/ttn/refactor"
	. "github.com/TheThingsNetwork/ttn/refactor/adapters/http"
	//"github.com/TheThingsNetwork/ttn/utils/errors"
	//"github.com/brocaar/lorawan"
)

// Collect defines a handler for retrieving raw packets sent by a POST request.
//
// It listens to requests of the form: [POST] /packets/
//
// It expects an http header Content-Type = application/octet-stream
//
// The body is expected to a binary marshaling of the given packet
//
// This handler does not generate any registration.
type Collect struct{}

// Url implements the http.Handler interface
func (p Collect) Url() string {
	return "/packets/"
}

// Handle implements the http.Handler interface
func (p Collect) Handle(w http.ResponseWriter, chpkt chan<- PktReq, chreg chan<- RegReq, req *http.Request) {
}

// parse extracts params from the request and fails if the request is invalid.
func (p Collect) parse(req *http.Request) (core.Registration, error) {
	return nil, nil
}

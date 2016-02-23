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
	// core "github.com/TheThingsNetwork/ttn/refactor"
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
	// Check the http method
	if req.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("Unreckognized HTTP method. Please use [POST] to transfer a packet"))
		return
	}

	// Parse body and query params
	data, err := p.parse(req)
	if err != nil {
		BadRequest(w, err.Error())
		return
	}

	// Send the packet and wait for ack / nack
	chresp := make(chan MsgRes)
	chpkt <- PktReq{Packet: data, Chresp: chresp}
	r, ok := <-chresp
	if !ok {
		BadRequest(w, "Core server not responding")
		return
	}
	w.WriteHeader(r.StatusCode)
	w.Write(r.Content)
}

// parse extracts params from the request and fails if the request is invalid.
func (p Collect) parse(req *http.Request) ([]byte, error) {
	return nil, nil
}

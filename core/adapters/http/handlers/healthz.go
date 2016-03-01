// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package handlers

import (
	"net/http"

	. "github.com/TheThingsNetwork/ttn/core/adapters/http"
)

// Collect defines a handler to ping adapters via a GET request.
//
// It listens to requests of the form: [GET] /healthz/
//
//
// This handler does not generate any packet.
// This handler does not generate any registration.
type Healthz struct{}

// Url implements the http.Handler interface
func (p Healthz) Url() string {
	return "/healthz"
}

// Handle implements the http.Handler interface
func (p Healthz) Handle(w http.ResponseWriter, chpkt chan<- PktReq, chreg chan<- RegReq, req *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

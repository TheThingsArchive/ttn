// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package handlers

import (
	"net/http"

	. "github.com/TheThingsNetwork/ttn/core/adapters/http"
)

// Healthz defines a handler to ping adapters via a GET request.
//
// It listens to requests of the form: [GET] /healthz
//
//
// This handler does not generate any packet.
// This handler does not generate any registration.
type Healthz struct{}

// URL implements the http.Handler interface
func (p Healthz) URL() string {
	return "/healthz"
}

// Handle implements the http.Handler interface
func (p Healthz) Handle(w http.ResponseWriter, chpkt chan<- PktReq, chreg chan<- RegReq, req *http.Request) error {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
	return nil
}

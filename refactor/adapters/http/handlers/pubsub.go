// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package handlers

import (
	"net/http"

	core "github.com/TheThingsNetwork/ttn/refactor"
	. "github.com/TheThingsNetwork/ttn/refactor/adapters/http"
)

// Pubsub defines an handler to handle application | devEUI registration on a component.
//
// It listens to request of the form: [PUT] /end-devices/:devEUI
// where devEUI is a 8 bytes hex-encoded address.
//
// It also looks for params:
//
// - app_eui (8 bytes hex-encoded string)
// - app_url (http address as string)
// - nwks_key (16 bytes hex-encoded string)
//
// It fails with an http 400 Bad Request. if one of the parameter is missing or invalid
// It succeeds with an http 2xx if the request is valid (the response status is under the
// ackNacker responsibility.
// It can possibly fails with another status depending of the AckNacker response.
func PubSub(w http.ResponseWriter, reg chan<- RegReq, req *http.Request) {

}

// parse extracts params from the request and fails if the request is invalid.
func parse(req *http.Request) (core.Registration, error) {
	return nil, nil
}

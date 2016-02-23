// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package handlers

import (
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"regexp"
	"strings"

	. "github.com/TheThingsNetwork/ttn/core/errors"
	core "github.com/TheThingsNetwork/ttn/refactor"
	. "github.com/TheThingsNetwork/ttn/refactor/adapters/http"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	"github.com/brocaar/lorawan"
)

// Pubsub defines a handler to handle application | devEUI registration on a component.
//
// It listens to request of the form: [PUT] /end-devices/:devEUI
// where devEUI is a 8 bytes hex-encoded address.
//
// It expects a Content-Type = application/json
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
//
// The PubSub handler generates registration where:
// - AppEUI is available
// - DevEUI is available
// - NwkSKey is available
// - Recipient can be interpreted as an HttpRecipient (Url + Method)
type PubSub struct{}

// Url implements the http.Handler interface
func (p PubSub) Url() string {
	return "/end-devices/"
}

// Handle implements the http.Handler interface
func (p PubSub) Handle(w http.ResponseWriter, chreg chan<- RegReq, req *http.Request) {
	// Check the http method
	if req.Method != "PUT" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("Unreckognized HTTP method. Please use [PUT] to register a device"))
		return
	}

	// Parse body and query params
	registration, err := p.parse(req)
	if err != nil {
		BadRequest(w, err.Error())
		return
	}

	// Send the registration and wait for ack / nack
	chresp := make(chan MsgRes)
	chreg <- RegReq{Registration: registration, Chresp: chresp}
	r, ok := <-chresp
	if !ok {
		BadRequest(w, "Core server not responding")
		return
	}
	w.WriteHeader(r.StatusCode)
	w.Write(r.Content)
}

// parse extracts params from the request and fails if the request is invalid.
func (p PubSub) parse(req *http.Request) (core.Registration, error) {
	// Check Content-type
	if req.Header.Get("Content-Type") != "application/json" {
		return pubSubRegistration{}, errors.New(ErrInvalidStructure, "Received invalid content-type in request")
	}

	// Check the query parameter
	reg := regexp.MustCompile("end-devices/([a-fA-F0-9]{16})$") // 8-bytes, hex-encoded -> 16 chars
	query := reg.FindStringSubmatch(req.RequestURI)
	if len(query) < 2 {
		return pubSubRegistration{}, errors.New(ErrInvalidStructure, "Incorrect end-device address format")
	}
	devEUI, err := hex.DecodeString(query[1])
	if err != nil {
		return pubSubRegistration{}, errors.New(ErrInvalidStructure, err)
	}

	// Check configuration in body
	body := make([]byte, req.ContentLength)
	n, err := req.Body.Read(body)
	if err != nil && err != io.EOF {
		return pubSubRegistration{}, errors.New(ErrInvalidStructure, err)
	}
	params := &struct {
		AppEUI  string `json:"app_eui"`
		Url     string `json:"app_url"`
		NwkSKey string `json:"nwks_key"`
	}{}
	if err := json.Unmarshal(body[:n], params); err != nil {
		return pubSubRegistration{}, errors.New(ErrInvalidStructure, err)
	}

	// Verify each request parameter
	nwkSKey, err := hex.DecodeString(params.NwkSKey)
	if err != nil || len(nwkSKey) != 16 {
		return pubSubRegistration{}, errors.New(ErrInvalidStructure, "Incorrect network session key")
	}

	appEUI, err := hex.DecodeString(params.AppEUI)
	if err != nil || len(appEUI) != 8 {
		return pubSubRegistration{}, errors.New(ErrInvalidStructure, "Incorrect application eui")
	}

	params.Url = strings.Trim(params.Url, " ")
	if len(params.Url) <= 0 {
		return pubSubRegistration{}, errors.New(ErrInvalidStructure, "Incorrect application url")
	}

	// Create actual registration
	registration := pubSubRegistration{
		recipient: NewHttpRecipient(params.Url, "PUT"),
		appEUI:    lorawan.EUI64{},
		devEUI:    lorawan.EUI64{},
		nwkSKey:   lorawan.AES128Key{},
	}

	copy(registration.appEUI[:], appEUI[:])
	copy(registration.nwkSKey[:], nwkSKey[:])
	copy(registration.devEUI[:], devEUI[:])

	return registration, nil
}

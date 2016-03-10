// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package handlers

import (
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/TheThingsNetwork/ttn/core"
	. "github.com/TheThingsNetwork/ttn/core/adapters/http"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	"github.com/brocaar/lorawan"
)

// PubSub defines a handler to handle application | devEUI registration on a component.
//
// It listens to request of the form: [PUT] /end-devices
// where devEUI is a 8 bytes hex-encoded address.
//
// It expects a Content-Type = application/json
//
// It also looks for params:
//
// - dev_eui (8 bytes hex-encoded string)
// - app_eui (8 bytes hex-encoded string)
// - recipient {
// -	url (http address as string)
// -    method (http verb as string)
// - }
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

// URL implements the http.Handler interface
func (p PubSub) URL() string {
	return "/end-devices"
}

// Handle implements the http.Handler interface
func (p PubSub) Handle(w http.ResponseWriter, chpkt chan<- PktReq, chreg chan<- RegReq, req *http.Request) error {
	// Check the http method
	if req.Method != "PUT" {
		err := errors.New(errors.Structural, "Unreckognized HTTP method. Please use [PUT] to register a device")
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte(err.Error()))
		return err
	}

	// Parse body and query params
	registration, err := p.parse(req)
	if err != nil {
		BadRequest(w, err.Error())
		return err
	}

	// Send the registration and wait for ack / nack
	chresp := make(chan MsgRes)
	chreg <- RegReq{Registration: registration, Chresp: chresp}
	r, ok := <-chresp
	if !ok {
		err := errors.New(errors.Operational, "Core server not responding")
		BadRequest(w, err.Error())
		return err
	}
	w.WriteHeader(r.StatusCode)
	w.Write(r.Content)
	return nil
}

// parse extracts params from the request and fails if the request is invalid.
func (p PubSub) parse(req *http.Request) (core.Registration, error) {
	// Check Content-type
	if req.Header.Get("Content-Type") != "application/json" {
		return pubSubRegistration{}, errors.New(errors.Structural, "Received invalid content-type in request")
	}

	// Check configuration in body
	body := make([]byte, req.ContentLength)
	n, err := req.Body.Read(body)
	if err != nil && err != io.EOF {
		return pubSubRegistration{}, errors.New(errors.Structural, err)
	}
	defer req.Body.Close()
	params := new(struct {
		Recipient struct {
			URL    string `json:"url"`
			Method string `json:"method"`
		} `json:"recipient"`
		Registration struct {
			AppEUI  string `json:"app_eui"`
			DevEUI  string `json:"dev_eui"`
			NwkSKey string `json:"nwks_key"`
		} `json:"registration"`
	})
	if err := json.Unmarshal(body[:n], params); err != nil {
		return pubSubRegistration{}, errors.New(errors.Structural, "Unable to unmarshal the request body")
	}

	// Verify each request parameter
	nwkSKey, err := hex.DecodeString(params.Registration.NwkSKey)
	if err != nil || len(nwkSKey) != 16 {
		return pubSubRegistration{}, errors.New(errors.Structural, "Incorrect network session key")
	}

	appEUI, err := hex.DecodeString(params.Registration.AppEUI)
	if err != nil || len(appEUI) != 8 {
		return pubSubRegistration{}, errors.New(errors.Structural, "Incorrect application eui")
	}

	devEUI, err := hex.DecodeString(params.Registration.DevEUI)
	if err != nil || len(devEUI) != 8 {
		return pubSubRegistration{}, errors.New(errors.Structural, "Incorrect device eui")
	}

	params.Recipient.URL = strings.Trim(params.Recipient.URL, " ")
	if len(params.Recipient.URL) <= 0 {
		return pubSubRegistration{}, errors.New(errors.Structural, "Incorrect application url")
	}

	params.Recipient.Method = strings.Trim(params.Recipient.Method, " ")
	if len(params.Recipient.Method) <= 0 {
		return pubSubRegistration{}, errors.New(errors.Structural, "Incorrect application method")
	}

	// Create actual registration
	registration := pubSubRegistration{
		recipient: NewRecipient(params.Recipient.URL, params.Recipient.Method),
		appEUI:    lorawan.EUI64{},
		devEUI:    lorawan.EUI64{},
		nwkSKey:   lorawan.AES128Key{},
	}

	copy(registration.appEUI[:], appEUI[:])
	copy(registration.nwkSKey[:], nwkSKey[:])
	copy(registration.devEUI[:], devEUI[:])

	return registration, nil
}

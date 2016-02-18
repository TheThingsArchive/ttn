// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package parser

import (
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"regexp"
	"strings"

	"github.com/TheThingsNetwork/ttn/core"
	. "github.com/TheThingsNetwork/ttn/core/errors"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	"github.com/brocaar/lorawan"
)

// Parser gives a flexible way of parsing a request into a registration.
type RegistrationParser interface {
	// Parse transforms a given http request into a Registration. The error handling is not under its
	// responsibility. The parser can expect any query param, http method or header it needs.
	Parse(req *http.Request) (core.Registration, error)
}

// PubSub materializes a parser for pubsub requests.
//
// It expects requests to be of the following shapes:
//
//     Content-Type: application/json
//     Method: PUT
//     Path: end-devices/<devAddr> where <devAddr> is 8 bytes hex encoded
//     Params: app_id (string), app_url (string), nwks_key (16 bytes hex encoded)
type PubSub struct{}

// Parse implements the RegistrationParser interface
func (p PubSub) Parse(req *http.Request) (core.Registration, error) {
	// Check Content-type
	if req.Header.Get("Content-Type") != "application/json" {
		return core.Registration{}, errors.New(ErrInvalidStructure, "Received invalid content-type in request")
	}

	// Check the query parameter
	reg := regexp.MustCompile("end-devices/([a-fA-F0-9]{8})$")
	query := reg.FindStringSubmatch(req.RequestURI)
	if len(query) < 2 {
		return core.Registration{}, errors.New(ErrInvalidStructure, "Incorrect end-device address format")
	}
	devAddr, err := hex.DecodeString(query[1])
	if err != nil {
		return core.Registration{}, errors.New(ErrInvalidStructure, err)
	}

	// Check configuration in body
	body := make([]byte, req.ContentLength)
	n, err := req.Body.Read(body)
	if err != nil && err != io.EOF {
		return core.Registration{}, errors.New(ErrInvalidStructure, err)
	}
	params := &struct {
		Id      string `json:"app_id"`
		Url     string `json:"app_url"`
		NwkSKey string `json:"nwks_key"`
	}{}
	if err := json.Unmarshal(body[:n], params); err != nil {
		return core.Registration{}, errors.New(ErrInvalidStructure, err)
	}

	nwkSKey, err := hex.DecodeString(params.NwkSKey)
	if err != nil || len(nwkSKey) != 16 {
		return core.Registration{}, errors.New(ErrInvalidStructure, "Incorrect network session key")
	}

	params.Id = strings.Trim(params.Id, " ")
	params.Url = strings.Trim(params.Url, " ")
	if len(params.Id) <= 0 {
		return core.Registration{}, errors.New(ErrInvalidStructure, "Incorrect application id")
	}
	if len(params.Url) <= 0 {
		return core.Registration{}, errors.New(ErrInvalidStructure, "Incorrect application url")
	}

	// Create registration
	config := core.Registration{
		Recipient: core.Recipient{Id: params.Id, Address: params.Url},
	}
	options := lorawan.AES128Key{}
	copy(options[:], nwkSKey)
	config.Options = options
	copy(config.DevAddr[:], devAddr)

	return config, nil
}

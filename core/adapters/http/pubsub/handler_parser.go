// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package pubsub

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"

	"github.com/TheThingsNetwork/ttn/core"
	"github.com/TheThingsNetwork/ttn/lorawan"
)

type HandlerParser struct{}

func (p HandlerParser) Parse(req *http.Request) (core.Registration, error) {
	// Check Content-type
	if req.Header.Get("Content-Type") != "application/json" {
		return core.Registration{}, fmt.Errorf("Received invalid content-type in request")
	}

	// Check the query parameter
	reg := regexp.MustCompile("end-devices/([a-fA-F0-9]{8})$")
	query := reg.FindStringSubmatch(req.RequestURI)
	if len(query) < 2 {
		return core.Registration{}, fmt.Errorf("Incorrect end-device address format")
	}
	devAddr, err := hex.DecodeString(query[1])
	if err != nil {
		return core.Registration{}, err
	}

	// Check configuration in body
	body := make([]byte, req.ContentLength)
	n, err := req.Body.Read(body)
	if err != nil && err != io.EOF {
		return core.Registration{}, err
	}
	params := &struct {
		Id     string `json:"app_id"`
		Url    string `json:"app_url"`
		NwsKey string `json:"nws_key"`
	}{}
	if err := json.Unmarshal(body[:n], params); err != nil {
		return core.Registration{}, err
	}

	nwsKey, err := hex.DecodeString(params.NwsKey)
	if err != nil || len(nwsKey) != 16 {
		return core.Registration{}, fmt.Errorf("Incorrect network session key")
	}

	params.Id = strings.Trim(params.Id, " ")
	params.Url = strings.Trim(params.Url, " ")
	if len(params.Id) <= 0 {
		return core.Registration{}, fmt.Errorf("Incorrect application id")
	}
	if len(params.Url) <= 0 {
		return core.Registration{}, fmt.Errorf("Incorrect application url")
	}

	// Create registration
	config := core.Registration{
		Recipient: core.Recipient{Id: params.Id, Address: params.Url},
	}
	options := lorawan.AES128Key{}
	copy(options[:], nwsKey)
	config.Options = options
	copy(config.DevAddr[:], devAddr)

	return config, nil
}

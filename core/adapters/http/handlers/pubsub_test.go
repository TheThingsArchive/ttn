// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package handlers

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/TheThingsNetwork/ttn/core"
	. "github.com/TheThingsNetwork/ttn/core/adapters/http"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	. "github.com/TheThingsNetwork/ttn/utils/errors/checks"
	. "github.com/TheThingsNetwork/ttn/utils/testing"
	"github.com/brocaar/lorawan"
)

func TestPubSub(t *testing.T) {
	tests := []struct {
		Desc        string
		Payload     string
		ContentType string
		Method      string
		DevEUI      string
		ShouldAck   bool
		AckPacket   core.Packet

		WantContent      string
		WantStatusCode   int
		WantRegistration core.Registration
		WantError        *string
	}{
		{
			Desc:        "Invalid Payload. Valid ContentType. Valid Method. Valid DevEUI. Nack",
			Payload:     "TheThingsNetwork",
			ContentType: "application/json",
			Method:      "PUT",
			ShouldAck:   false,

			WantContent:      string(errors.Structural),
			WantStatusCode:   http.StatusBadRequest,
			WantRegistration: nil,
			WantError:        nil,
		},
		{
			Desc: "Valid Payload. Invalid ContentType. Valid Method. Valid DevEUI. Nack",
			Payload: `{
				"registration":{
					"app_eui":"0001020304050607",
					"nwks_key":"00010203040506070809000102030405",
					"dev_eui": "0000000001020304"
				},
				"recipient": {
					"url": "url",
					"method": "PUT"
				}
			}`,
			ContentType: "text/plain",
			Method:      "PUT",
			DevEUI:      "",
			ShouldAck:   false,

			WantContent:      string(errors.Structural),
			WantStatusCode:   http.StatusBadRequest,
			WantRegistration: nil,
			WantError:        nil,
		},
		{
			Desc: "Valid Payload. Valid ContentType. Invalid Method. Valid DevEUI. Nack",
			Payload: `{
				"registration":{
					"app_eui":"0001020304050607",
					"nwks_key":"00010203040506070809000102030405",
					"dev_eui": "0000000001020304"
				},
				"recipient": {
					"url": "url",
					"method": "PUT"
				}
			}`,
			ContentType: "application/json",
			Method:      "POST",
			DevEUI:      "0000000001020304",
			ShouldAck:   false,

			WantContent:      string(errors.Structural),
			WantStatusCode:   http.StatusMethodNotAllowed,
			WantRegistration: nil,
			WantError:        nil,
		},
		{
			Desc: "Valid Payload. Valid ContentType. Valid Method. Invalid DevEUI. Nack",
			Payload: `{
				"registration":{
					"app_eui":"0001020304050607",
					"nwks_key":"00010203040506070809000102030405",
					"dev_eui": "0000001144"
				},
				"recipient": {
					"url": "url",
					"method": "PUT"
				}
			}`,
			ContentType: "application/json",
			Method:      "PUT",
			DevEUI:      "12345678",
			ShouldAck:   false,

			WantContent:      string(errors.Structural),
			WantStatusCode:   http.StatusBadRequest,
			WantRegistration: nil,
			WantError:        nil,
		},
		{
			Desc: "Valid Payload. Valid ContentType. Valid Method. Valid DevEUI. Nack",
			Payload: `{
				"registration":{
					"app_eui":"0001020304050607",
					"nwks_key":"00010203040506070809000102030405",
					"dev_eui": "0000000001020304"
				},
				"recipient": {
					"url": "url",
					"method": "PUT"
				}
			}`,
			ContentType: "application/json",
			Method:      "PUT",
			DevEUI:      "0000000001020304",
			ShouldAck:   false,

			WantContent:    string(errors.Structural),
			WantStatusCode: http.StatusConflict,
			WantRegistration: pubSubRegistration{
				recipient: NewRecipient("url", "PUT"),
				appEUI:    lorawan.EUI64([8]byte{0, 1, 2, 3, 4, 5, 6, 7}),
				nwkSKey:   lorawan.AES128Key([16]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5}),
				devEUI:    lorawan.EUI64([8]byte{0, 0, 0, 0, 1, 2, 3, 4}),
			},
			WantError: nil,
		},
		{
			Desc: "Valid Payload. Valid ContentType. Valid Method. Valid DevEUI. Ack",
			Payload: `{
				"registration":{
					"app_eui":"0001020304050607",
					"nwks_key":"00010203040506070809000102030405",
					"dev_eui": "0000000001020304"
				},
				"recipient": {
					"url": "url",
					"method": "PUT"
				}
			}`,
			ContentType: "application/json",
			Method:      "PUT",
			DevEUI:      "0000000001020304",
			ShouldAck:   true,

			WantContent:    "",
			WantStatusCode: http.StatusAccepted,
			WantRegistration: pubSubRegistration{
				recipient: NewRecipient("url", "PUT"),
				appEUI:    lorawan.EUI64([8]byte{0, 1, 2, 3, 4, 5, 6, 7}),
				nwkSKey:   lorawan.AES128Key([16]byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 0, 1, 2, 3, 4, 5}),
				devEUI:    lorawan.EUI64([8]byte{0, 0, 0, 0, 1, 2, 3, 4}),
			},
			WantError: nil,
		},
	}

	var port uint = 4000
	for _, test := range tests {
		// Describe
		Desc(t, test.Desc)

		// Build
		adapter, url := createPubSubAdapter(t, port)
		port++
		client := testClient{}

		// Operate
		url = fmt.Sprintf("%s", url)
		chresp := client.Send(test.Payload, url, test.Method, test.ContentType)
		registration, err := tryNextRegistration(adapter, test.ShouldAck, test.AckPacket)
		var statusCode int
		var content []byte
		select {
		case resp := <-chresp:
			statusCode = resp.StatusCode
			content = resp.Content
		case <-time.After(time.Millisecond * 100):
		}

		// Check
		t.Log(string(content))
		CheckErrors(t, test.WantError, err)
		checkStatusCode(t, test.WantStatusCode, statusCode)
		checkContent(t, test.WantContent, content)
		checkRegistration(t, test.WantRegistration, registration)
	}
}

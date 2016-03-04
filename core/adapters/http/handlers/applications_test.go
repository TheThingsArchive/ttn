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

func TestApplications(t *testing.T) {
	tests := []struct {
		Desc        string
		Payload     string
		ContentType string
		Method      string
		ShouldAck   bool
		AckPacket   core.Packet

		WantContent      string
		WantStatusCode   int
		WantRegistration core.Registration
		WantError        *string
	}{
		{
			Desc:        "Invalid Payload. Valid ContentType. Valid Method. Nack",
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
			Desc:        "Valid Payload. Invalid ContentType. Valid Method. Nack",
			Payload:     `{"app_eui":"0000000011223344","app_url":"url"}`,
			ContentType: "text/plain",
			Method:      "PUT",
			ShouldAck:   false,

			WantContent:      string(errors.Structural),
			WantStatusCode:   http.StatusBadRequest,
			WantRegistration: nil,
			WantError:        nil,
		},
		{
			Desc:        "Valid Payload. Valid ContentType. Invalid Method. Nack",
			Payload:     `{"app_eui":"0000000011223344","app_url":"url"}`,
			ContentType: "application/json",
			Method:      "POST",
			ShouldAck:   false,

			WantContent:      string(errors.Structural),
			WantStatusCode:   http.StatusMethodNotAllowed,
			WantRegistration: nil,
			WantError:        nil,
		},
		{
			Desc:        "Valid Payload. Valid ContentType. Valid Method.  Nack",
			Payload:     `{"app_eui":"000011223344","app_url":"url"}`,
			ContentType: "application/json",
			Method:      "PUT",
			ShouldAck:   false,

			WantContent:      string(errors.Structural),
			WantStatusCode:   http.StatusBadRequest,
			WantRegistration: nil,
			WantError:        nil,
		},
		{
			Desc:        "Valid Payload. Valid ContentType. Valid Method. Nack",
			Payload:     `{"app_eui":"0000000001020304","app_url":"url"}`,
			ContentType: "application/json",
			Method:      "PUT",
			ShouldAck:   false,

			WantContent:    string(errors.Structural),
			WantStatusCode: http.StatusConflict,
			WantRegistration: applicationsRegistration{
				recipient: NewRecipient("url", "PUT"),
				appEUI:    lorawan.EUI64([8]byte{0, 0, 0, 0, 1, 2, 3, 4}),
			},
			WantError: nil,
		},
		{
			Desc:        "Valid Payload. Valid ContentType. Valid Method. Ack",
			Payload:     `{"app_eui":"0000000001020304","app_url":"url"}`,
			ContentType: "application/json",
			Method:      "PUT",
			ShouldAck:   true,

			WantContent:    "",
			WantStatusCode: http.StatusAccepted,
			WantRegistration: applicationsRegistration{
				recipient: NewRecipient("url", "PUT"),
				appEUI:    lorawan.EUI64([8]byte{0, 0, 0, 0, 1, 2, 3, 4}),
			},
			WantError: nil,
		},
	}

	var port uint = 4200
	for _, test := range tests {
		// Describe
		Desc(t, test.Desc)

		// Build
		adapter, url := createApplicationsAdapter(t, port)
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
		CheckErrors(t, test.WantError, err)
		checkStatusCode(t, test.WantStatusCode, statusCode)
		checkContent(t, test.WantContent, content)
		checkRegistration(t, test.WantRegistration, registration)
	}
}

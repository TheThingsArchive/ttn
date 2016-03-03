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
		AppEUI      string
		ShouldAck   bool
		AckPacket   core.Packet

		WantContent      string
		WantStatusCode   int
		WantRegistration core.Registration
		WantError        *string
	}{
		{
			Desc:        "Invalid Payload. Valid ContentType. Valid Method. Valid AppEUI. Nack",
			Payload:     "TheThingsNetwork",
			ContentType: "application/json",
			Method:      "PUT",
			AppEUI:      "0000000011223344",
			ShouldAck:   false,

			WantContent:      string(errors.Structural),
			WantStatusCode:   http.StatusBadRequest,
			WantRegistration: nil,
			WantError:        nil,
		},
		{
			Desc:        "Valid Payload. Invalid ContentType. Valid Method. Valid AppEUI. Nack",
			Payload:     `{"app_url":"url"}`,
			ContentType: "text/plain",
			Method:      "PUT",
			AppEUI:      "0000000011223344",
			ShouldAck:   false,

			WantContent:      string(errors.Structural),
			WantStatusCode:   http.StatusBadRequest,
			WantRegistration: nil,
			WantError:        nil,
		},
		{
			Desc:        "Valid Payload. Valid ContentType. Invalid Method. Valid AppEUI. Nack",
			Payload:     `{"app_url":"url"}`,
			ContentType: "application/json",
			Method:      "POST",
			AppEUI:      "0000000011223344",
			ShouldAck:   false,

			WantContent:      string(errors.Structural),
			WantStatusCode:   http.StatusMethodNotAllowed,
			WantRegistration: nil,
			WantError:        nil,
		},
		{
			Desc:        "Valid Payload. Valid ContentType. Valid Method. Invalid AppEUI. Nack",
			Payload:     `{"app_url":"url"}`,
			ContentType: "application/json",
			Method:      "PUT",
			AppEUI:      "12345678",
			ShouldAck:   false,

			WantContent:      string(errors.Structural),
			WantStatusCode:   http.StatusBadRequest,
			WantRegistration: nil,
			WantError:        nil,
		},
		{
			Desc:        "Valid Payload. Valid ContentType. Valid Method. Valid AppEUI. Nack",
			Payload:     `{"app_url":"url"}`,
			ContentType: "application/json",
			Method:      "PUT",
			AppEUI:      "0000000001020304",
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
			Desc:        "Valid Payload. Valid ContentType. Valid Method. Valid AppEUI. Ack",
			Payload:     `{"app_url":"url"}`,
			ContentType: "application/json",
			Method:      "PUT",
			AppEUI:      "0000000001020304",
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
		url = fmt.Sprintf("%s%s", url, test.AppEUI)
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

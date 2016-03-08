// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package handlers

import (
	"net/http"
	"testing"
	"time"

	"github.com/TheThingsNetwork/ttn/core"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	. "github.com/TheThingsNetwork/ttn/utils/errors/checks"
	. "github.com/TheThingsNetwork/ttn/utils/testing"
)

func TestCollect(t *testing.T) {
	tests := []struct {
		Desc        string
		Payload     string
		ContentType string
		Method      string
		ShouldAck   bool
		AckPacket   core.Packet

		WantContent    string
		WantStatusCode int
		WantPacket     []byte
		WantError      *string
	}{
		{
			Desc:        "Valid Payload. Invalid ContentType. Valid Method. Nack.",
			Payload:     "Patate",
			ContentType: "application/patate",
			Method:      "POST",
			ShouldAck:   false,

			WantContent:    string(errors.Structural),
			WantStatusCode: http.StatusBadRequest,
			WantPacket:     nil,
			WantError:      nil,
		},
		{
			Desc:        "Valid Payload. Valid ContentType. Invalid Method. Nack.",
			Payload:     "Patate",
			ContentType: "application/octet-stream",
			Method:      "PUT",
			ShouldAck:   false,

			WantContent:    string(errors.Structural),
			WantStatusCode: http.StatusMethodNotAllowed,
			WantPacket:     nil,
			WantError:      nil,
		},
		{
			Desc:        "Valid Payload. Valid ContentType. Valid Method. Nack.",
			Payload:     "Patate",
			ContentType: "application/octet-stream",
			Method:      "POST",
			ShouldAck:   false,

			WantContent:    "Unknown",
			WantStatusCode: http.StatusInternalServerError,
			WantPacket:     []byte("Patate"),
			WantError:      nil,
		},
		{
			Desc:        "Invalid Ack. Valid ContentType. Valid Method.",
			Payload:     "Patate",
			ContentType: "application/octet-stream",
			Method:      "POST",
			ShouldAck:   true,
			AckPacket:   testPacket{payload: ""},

			WantContent:    string(errors.Operational),
			WantStatusCode: http.StatusBadRequest,
			WantPacket:     []byte("Patate"),
			WantError:      nil,
		},
		{
			Desc:        "Valid Ack. Valid ContentType. Valid Method.",
			Payload:     "Patate",
			ContentType: "application/octet-stream",
			Method:      "POST",
			ShouldAck:   true,
			AckPacket:   testPacket{payload: "Response"},

			WantContent:    "Response",
			WantStatusCode: http.StatusOK,
			WantPacket:     []byte("Patate"),
			WantError:      nil,
		},
	}

	var port uint = 3000
	for _, test := range tests {
		// Describe
		Desc(t, test.Desc)

		// Build
		adapter, url := createCollectAdapter(t, port)
		client := testClient{}

		// Operate
		chresp := client.Send(test.Payload, url, test.Method, test.ContentType)
		packet, err := tryNext(adapter, test.ShouldAck, test.AckPacket)
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
		checkPacket(t, test.WantPacket, packet)
		port++
	}
}

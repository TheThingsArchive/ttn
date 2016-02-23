// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package handlers

import (
	"fmt"
	"testing"
	"time"

	core "github.com/TheThingsNetwork/ttn/refactor"
	. "github.com/TheThingsNetwork/ttn/utils/testing"
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
	}{}

	var port uint = 4000
	for _, test := range tests {
		// Describe
		Desc(t, test.Desc)

		// Build
		adapter, url := createPubSubAdapter(t, port)
		port += 1
		client := testClient{}

		// Operate
		url = fmt.Sprintf("%s/%s", url, test.DevEUI)
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
		checkErrors(t, test.WantError, err)
		checkStatusCode(t, test.WantStatusCode, statusCode)
		checkContent(t, test.WantContent, content)
		checkRegistration(t, test.WantRegistration, registration)
	}
}

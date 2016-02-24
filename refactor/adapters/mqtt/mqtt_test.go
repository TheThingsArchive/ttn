// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package mqtt

import (
	"reflect"
	"testing"

	"github.com/TheThingsNetwork/ttn/utils/errors"
	. "github.com/TheThingsNetwork/ttn/utils/testing"
	//"github.com/TheThingsNetwork/ttn/utils/pointer"
)

func TestMQTTSend(t *testing.T) {
	tests := []struct {
		Desc      string          // Test Description
		Packet    []byte          // Handy representation of the packet to send
		Recipient []testRecipient // List of recipient to send

		WantData     []byte  // Expected Data on the recipient
		WantResponse []byte  // Expected Response from the Send method
		WantError    *string // Expected error nature returned by the Send method
	}{}

	for _, test := range tests {
		// Describe
		Desc(t, test.Desc)

		// Build
		// Generate new adapter
		// Generate reception servers

		// Operate
		// Send data to recipients
		// Retrieve data from servers

		// Check
		// Check if data has been received
		// Check if response is valid
		// Check if error is valid
	}
}

// ----- TYPE utilities
type testRecipient struct {
	Response  []byte
	TopicUp   string
	TopicDown string
}

// ----- BUILD utilities

// ----- OPERATE utilities

// ----- CHECK utilities
func checkErrors(t *testing.T, want *string, got error) {
	if got == nil {
		if want == nil {
			Ok(t, "Check errors")
			return
		}
		Ko(t, "Expected error to be {%s} but got nothing", *want)
		return
	}

	if want == nil {
		Ko(t, "Expected no error but got {%v}", got)
		return
	}

	if got.(errors.Failure).Nature == *want {
		Ok(t, "Check errors")
		return
	}
	Ko(t, "Expected error to be {%s} but got {%v}", *want, got)
}

func checkResponses(t *testing.T, want []byte, got []byte) {
	if reflect.DeepEqual(want, got) {
		Ok(t, "Check responses")
		return
	}
	Ko(t, "Received response does not match expectations.\nWant: %s\nGot:  %s", string(want), string(got))
}

func checkData(t *testing.T, want []byte, got []byte) {
	if reflect.DeepEqual(want, got) {
		Ok(t, "Check data")
		return
	}
	Ko(t, "Received data does not match expectations.\nWant: %s\nGot:  %s", string(want), string(got))
}

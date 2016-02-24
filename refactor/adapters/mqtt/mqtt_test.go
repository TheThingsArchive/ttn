// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package mqtt

import (
	"testing"

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

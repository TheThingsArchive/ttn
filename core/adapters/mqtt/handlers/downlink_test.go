// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package handlers

import (
	"testing"
	"time"

	"github.com/TheThingsNetwork/ttn/core"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	. "github.com/TheThingsNetwork/ttn/utils/errors/checks"
	"github.com/TheThingsNetwork/ttn/utils/pointer"
	. "github.com/TheThingsNetwork/ttn/utils/testing"
	"github.com/brocaar/lorawan"
)

func TestDownlinkTopic(t *testing.T) {
	wantTopic := "+/devices/+/down"

	// Describe
	Desc(t, "Topic should equal: %s", wantTopic)

	// Build
	handler := Downlink{}

	// Operate
	topic := handler.Topic()

	// Check
	checkTopics(t, wantTopic, topic)
}

func TestDownlinkHandle(t *testing.T) {
	packet, _ := core.NewAPacket(
		lorawan.EUI64([8]byte{1, 1, 1, 1, 1, 1, 1, 1}),
		lorawan.EUI64([8]byte{2, 2, 2, 2, 2, 2, 2, 2}),
		[]byte{1, 2, 3, 4},
		nil,
	)

	tests := []struct {
		Desc    string      // The test's description
		Client  *MockClient // An mqtt client to mock (or not) the behavior
		Topic   string      // The topic to which the message is addressed
		Payload []byte      // The message's payload

		WantError        *string            // The expected error from the handler
		WantSubscription *string            // The topic to which a subscription is expected
		WantRegistration core.HRegistration // The expected registration towards the adapter
		WantPacket       core.Packet        // The expected packet towards the adapter
	}{

		{
			Desc:    "Valid Payload | Valid Topic",
			Client:  NewMockClient(),
			Payload: []byte{1, 2, 3, 4},
			Topic:   "0101010101010101/devices/0202020202020202/down",

			WantError:        nil,
			WantPacket:       packet,
			WantSubscription: nil,
			WantRegistration: nil,
		},
		{
			Desc:    "Valid Payload | Invalid Topic #2",
			Client:  NewMockClient(),
			Payload: []byte{1, 2, 3, 4},
			Topic:   "0101010101010101/devices/0202020202020202/down/again",

			WantError:        pointer.String(string(errors.Structural)),
			WantPacket:       nil,
			WantSubscription: nil,
			WantRegistration: nil,
		},
		{
			Desc:    "Valid Payload | Invalid Topic",
			Client:  NewMockClient(),
			Payload: []byte{1, 2, 3, 4},
			Topic:   "0101010101010101/devices/0202020202020202",

			WantError:        pointer.String(string(errors.Structural)),
			WantPacket:       nil,
			WantSubscription: nil,
			WantRegistration: nil,
		},
		{
			Desc:    "Valid Payload | Invalid AppEUI",
			Client:  NewMockClient(),
			Payload: []byte{1, 2, 3, 4},
			Topic:   "010101/devices/0202020202020202/down",

			WantError:        pointer.String(string(errors.Structural)),
			WantPacket:       nil,
			WantSubscription: nil,
			WantRegistration: nil,
		},
		{
			Desc:    "Valid Payload | Invalid DevEUI",
			Client:  NewMockClient(),
			Payload: []byte{1, 2, 3, 4},
			Topic:   "0101010101010101/devices/020202/down",

			WantError:        pointer.String(string(errors.Structural)),
			WantPacket:       nil,
			WantSubscription: nil,
			WantRegistration: nil,
		},
		{
			Desc:    "Invalid Payload | Valid Topic",
			Client:  NewMockClient(),
			Payload: []byte{},
			Topic:   "0101010101010101/devices/0202020202020202/down",

			WantError:        pointer.String(string(errors.Structural)),
			WantPacket:       nil,
			WantSubscription: nil,
			WantRegistration: nil,
		},
	}

	for i, test := range tests {
		// Describe
		Desc(t, "#%d: %s", i, test.Desc)

		// Build
		consumer, chpkt, chreg := newTestConsumer()
		handler := Downlink{}

		// Operate
		err := handler.Handle(test.Client, chpkt, chreg, MockMessage{
			payload: test.Payload,
			topic:   test.Topic,
		})
		<-time.After(time.Millisecond * 100)

		// Check
		CheckErrors(t, test.WantError, err)
		checkSubscriptions(t, test.WantSubscription, test.Client.InSubscribe)
		checkRegistrations(t, test.WantRegistration, consumer.Registration)
		checkPackets(t, test.WantPacket, consumer.Packet)
	}
}

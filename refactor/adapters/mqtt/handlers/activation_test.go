// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package handlers

import (
	"testing"

	core "github.com/TheThingsNetwork/ttn/refactor"
	. "github.com/TheThingsNetwork/ttn/utils/testing"
)

// func (a Activation) Handle(client Client, chpkt chan<- PktReq, chreg chan<- RegReq, msg MQTT.Message) {

func TestActivation(t *testing.T) {
	tests := []struct {
		Desc    string      // The test's description
		Client  *testClient // An mqtt client to mock (or not) the behavior
		Topic   string      // The topic to which the message is addressed
		Payload []byte      // The message's payload

		WantError        *string           // The expected error from the handler
		WantSubscription *string           // The topic to which a subscription is expected
		WantRegistration core.Registration // The expected registration towards the adapter
		WantPacket       []byte            // The expected packet towards the adapter
	}{}

	for i, test := range tests {
		// Describe
		Desc(t, "#%d: %s", i, test.Desc)

		// Build
		consumer, chpkt, chreg := newTestConsumer()
		handler := Activation{}

		// Operate
		err := handler.Handle(test.Client, chpkt, chreg, testMessage{
			payload: test.Payload,
			topic:   test.Topic,
		})

		// Check
		checkErrors(t, test.WantError, err)
		checkSubscriptions(t, test.WantSubscription, test.Client.Subscription)
		checkRegistrations(t, test.WantRegistration, consumer.Registration)
		checkPackets(t, test.WantPacket, consumer.Packet)
	}
}

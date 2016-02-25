// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package handlers

import (
	"reflect"
	"testing"

	core "github.com/TheThingsNetwork/ttn/refactor"
	. "github.com/TheThingsNetwork/ttn/refactor/adapters/mqtt"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	. "github.com/TheThingsNetwork/ttn/utils/testing"
)

// func (a Activation) Handle(client Client, chpkt chan<- PktReq, chreg chan<- RegReq, msg MQTT.Message) {

func TestActivation(t *testing.T) {
	tests := []struct {
		Desc    string // The test's description
		Client  Client // An mqtt client to mock (or not) the behavior
		Topic   string // The topic to which the message is addressed
		Payload []byte // The message's payload

		WantError        *string           // The expected error from the handler
		WantSubscription *string           // The topic to which a subscription is expected
		WantRegistration core.Registration // The expected registration towards the adapter
		WantPacket       []byte            // The expected packet towards the adapter
	}{}

	for i, test := range tests {
		// Describe
		Desc(t, "#%d: %s", i, test.Desc)

		// Build
		// Generate a mock Client
		// Generate a mock Consumer
		// Generate a handler

		// Operate
		// err := Handle message, client and channels to the handler
		// Retrieve subscriptions
		// Retrieve registrations

		// Check
		// Check errors
		// Check subscriptions
		// Check registrations
		// Check Packets
	}
}

// ----- BUILD utilities

// ----- OPERATE utilities

// ----- CHECK utilities
func checkErrors(t *testing.T, want *string, got error) {
	if got == nil {
		if want == nil {
			Ok(t, "Check Errors")
			return
		}
		Ko(t, "Expected error to be {%s} but got nothing", *want)
		return
	}

	if want == nil {
		Ko(t, "Expected no error but got {%v}", got)
		return
	}

	if got.(errors.Failure).Nature == errors.Nature(*want) {
		Ok(t, "Check Errors")
		return
	}
	Ko(t, "Expected error to be {%s} but got {%v}", *want, got)
}

func checkRegistrations(t *testing.T, want core.Registration, got core.Registration) {
	// Check if interfaces are nil
	if got == nil {
		if want == nil {
			Ok(t, "Check Registrations")
			return
		}
		Ko(t, "Expected registration to be %v but got nothing", want)
		return
	}
	if want == nil {
		Ko(t, "Expected no registration but got %v", got)
		return
	}

	// Check recipient topicUp
	rWant, ok := want.Recipient().(MqttRecipient)
	if !ok {
		panic("Expected test to be made with MQTTRecipient")
	}
	rGot, ok := got.Recipient().(MqttRecipient)
	if !ok {
		Ko(t, "Recipient isn't MqttRecipient: %v", got.Recipient())
		return
	}
	if rWant.TopicUp() != rGot.TopicUp() {
		Ko(t, "Recipients got different topics up.\nWant: %s\nGot:  %s", rWant.TopicUp(), rGot.TopicUp())
		return
	}

	// Check DevEUIs
	deWant, err := want.DevEUI()
	if err != nil {
		panic("Expected devEUI to be accessible in test registration")
	}
	deGot, err := got.DevEUI()
	if err != nil || !reflect.DeepEqual(deWant, deGot) {
		Ko(t, "Registrations' DevEUI are different.\nWant: %v\nGot:  %v", deWant, deGot)
		return
	}

	// Check AppEUIs
	aeWant, err := want.AppEUI()
	if err != nil {
		panic("Expected appEUI to be accessible in test registration")
	}
	aeGot, err := got.AppEUI()
	if err != nil || !reflect.DeepEqual(aeWant, aeGot) {
		Ko(t, "Registrations' appEUI are different.\nWant: %v\nGot:  %v", aeWant, aeGot)
		return
	}

	// Check Network Session Keys
	nkWant, err := want.NwkSKey()
	if err != nil {
		panic("Expected nwkSKey to be accessible in test registration")
	}
	nkGot, err := got.NwkSKey()
	if err != nil || !reflect.DeepEqual(nkWant, nkGot) {
		Ko(t, "Registrations' nwkSKey are different.\nWant: %v\nGot:  %v", nkWant, nkGot)
		return
	}

	// Check Application Session Keys
	akWant, err := want.AppSKey()
	if err != nil {
		panic("Expected nwkSKey to be accessible in test registration")
	}
	akGot, err := got.AppSKey()
	if err != nil || !reflect.DeepEqual(akWant, akGot) {
		Ko(t, "Registrations' nwkSKey are different.\nWant: %v\nGot:  %v", akWant, akGot)
		return
	}

	// Pheeew
	Ok(t, "Check Registrations")
}

func checkSubscriptions(t *testing.T, want *string, got *string) {
	if got == nil {
		if want == nil {
			Ok(t, "Check Subscriptions")
			return
		}
		Ko(t, "Expected subscription to be %s but got nothing", *want)
		return
	}
	if want == nil {
		Ko(t, "Expected no subscription but got %v", *got)
		return
	}

	if *want != *got {
		Ko(t, "Expected subscription to be %s but got %s", *want, *got)
		return
	}

	Ok(t, "Check Subscriptions")
}

func checkPackets(t *testing.T, want []byte, got []byte) {
	if reflect.DeepEqual(want, got) {
		Ok(t, "Check Packets")
		return
	}
	Ko(t, "Received packet does not match expectations.\nWant: %s\nGot:  %s", string(want), string(got))
}

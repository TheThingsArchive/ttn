// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package handlers

import (
	"reflect"
	"testing"

	"github.com/TheThingsNetwork/ttn/core"
	. "github.com/TheThingsNetwork/ttn/core/adapters/mqtt"
	. "github.com/TheThingsNetwork/ttn/utils/testing"
)

// testConsumer generates a component which consumes messages from two channels and make the last
// result available

type testConsumer struct {
	Packet       []byte
	Registration core.Registration
}

func newTestConsumer() (*testConsumer, chan PktReq, chan RegReq) {
	chpkt := make(chan PktReq)
	chreg := make(chan RegReq)
	consumer := testConsumer{}

	go func() {
		for msg := range chpkt {
			consumer.Packet = msg.Packet
		}
	}()

	go func() {
		for msg := range chreg {
			consumer.Registration = msg.Registration
		}
	}()

	return &consumer, chpkt, chreg
}

// ----- CHECK utilities
func checkTopics(t *testing.T, want string, got string) {
	if want == got {
		Ok(t, "Check Topics")
		return
	}

	Ko(t, "Topic does not match expectation.\nWant: %s\nGot:  %s", want, got)
}

func checkRegistrations(t *testing.T, want core.HRegistration, got core.Registration) {
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
	rWant, ok := want.Recipient().(Recipient)
	if !ok {
		panic("Expected test to be made with MQTTRecipient")
	}
	rGot, ok := got.Recipient().(Recipient)
	if !ok {
		Ko(t, "Recipient isn't MqttRecipient: %v", got.Recipient())
		return
	}
	if rWant.TopicUp() != rGot.TopicUp() {
		Ko(t, "Recipients got different topics up.\nWant: %s\nGot:  %s", rWant.TopicUp(), rGot.TopicUp())
		return
	}

	rgot, ok := got.(core.HRegistration)
	if !ok {
		Ko(t, "Expected to receive an HRegistration but got %+v", got)
		return
	}

	// Check DevEUIs
	if !reflect.DeepEqual(want.DevEUI(), rgot.DevEUI()) {
		Ko(t, "Registrations' DevEUI are different.\nWant: %v\nGot:  %v", want.DevEUI(), rgot.DevEUI())
		return
	}

	// Check AppEUIs
	if !reflect.DeepEqual(want.AppEUI(), rgot.AppEUI()) {
		Ko(t, "Registrations' appEUI are different.\nWant: %v\nGot:  %v", want.AppEUI(), rgot.AppEUI())
		return
	}

	// Check Network Session Keys
	if !reflect.DeepEqual(want.NwkSKey(), rgot.NwkSKey()) {
		Ko(t, "Registrations' nwkSKey are different.\nWant: %v\nGot:  %v", want.NwkSKey(), rgot.NwkSKey())
		return
	}

	// Check Application Session Keys
	if !reflect.DeepEqual(want.AppSKey(), rgot.AppSKey()) {
		Ko(t, "Registrations' appSKey are different.\nWant: %v\nGot:  %v", want.AppSKey(), rgot.AppSKey())
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

func checkPackets(t *testing.T, want core.Packet, got []byte) {
	if want == nil {
		if got == nil {
			Ok(t, "Check Packets")
			return
		}
		Ko(t, "Expected no packet but got %v", got)
		return
	}

	data, err := want.MarshalBinary()
	if err != nil {
		panic(err)
	}
	if reflect.DeepEqual(data, got) {
		Ok(t, "Check Packets")
		return
	}
	Ko(t, "Received packet does not match expectations.\nWant: %v\nGot:  %v", data, got)
}

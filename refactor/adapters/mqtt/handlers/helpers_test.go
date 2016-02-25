// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package handlers

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	MQTT "git.eclipse.org/gitroot/paho/org.eclipse.paho.mqtt.golang.git"
	core "github.com/TheThingsNetwork/ttn/refactor"
	. "github.com/TheThingsNetwork/ttn/refactor/adapters/mqtt"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	"github.com/TheThingsNetwork/ttn/utils/pointer"
	. "github.com/TheThingsNetwork/ttn/utils/testing"
)

// ----- TYPE utilities

// testToken gives a fake implementation of MQTT.Token
//
// Provide a failure if you need to simulate an Error() result.
type testToken struct {
	MQTT.Token
	Failure *string
}

func (t testToken) Wait() bool {
	return true
}

func (t testToken) WaitTimeout(d time.Duration) bool {
	<-time.After(d)
	return true
}

func (t testToken) Error() error {
	if t.Failure == nil {
		return nil
	}
	return fmt.Errorf(*t.Failure)
}

// testMessage gives a fake implementation of MQTT.Message
//
// provide payload and topic. Other methods are constants.
type testMessage struct {
	payload interface{}
	topic   string
}

func (m testMessage) Duplicate() bool {
	return false
}
func (m testMessage) Qos() byte {
	return 2
}
func (m testMessage) Retained() bool {
	return false
}
func (m testMessage) Topic() string {
	return m.topic
}
func (m testMessage) MessageID() uint16 {
	return 0
}
func (m testMessage) Payload() []byte {
	return m.payload.([]byte)
}

// testClient gives a fake implementation of a MQTT.ClientInt
//
// It saves the last subscription, unsubscriptions and publication call made
//
// It can also fails on demand (use the newTestClient method to define which methods should fail)
type testClient struct {
	Subscription    *string
	Unsubscriptions []string
	Publication     MQTT.Message

	failures  map[string]*string
	connected bool
}

func newTestClient(failures ...string) *testClient {
	client := testClient{failures: make(map[string]*string), connected: true}

	isFailure := func(x string) bool {
		for _, f := range failures {
			if f == x {
				return true
			}
		}
		return false
	}

	if isFailure("Connect") {
		client.failures["Connect"] = pointer.String("MockError -> Failed to connect")
	}

	if isFailure("Publish") {
		client.failures["Publish"] = pointer.String("MockError -> Failed to publish")
	}

	if isFailure("Subscribe") {
		client.failures["Subscribe"] = pointer.String("MockError -> Failed to subscribe")
	}

	if isFailure("Unsubscribe") {
		client.failures["Unsubscribe"] = pointer.String("MockError -> Failed to unsubscribe")
	}

	return &client
}

func (c *testClient) Connect() MQTT.Token {
	c.connected = true
	return testToken{Failure: c.failures["Connect"]}
}

func (c *testClient) Disconnect(quiesce uint) {
	<-time.After(time.Duration(quiesce))
	c.connected = false
	return
}

func (c testClient) IsConnected() bool {
	return c.connected
}

func (c *testClient) Publish(topic string, qos byte, retained bool, payload interface{}) MQTT.Token {
	c.Publication = testMessage{payload: payload, topic: topic}
	return testToken{Failure: c.failures["Publish"]}
}

func (c *testClient) Subscribe(topic string, qos byte, callback func(c Client, m MQTT.Message)) MQTT.Token {
	c.Subscription = &topic
	return testToken{Failure: c.failures["Subscribe"]}
}

func (c *testClient) Unsubscribe(topics ...string) MQTT.Token {
	c.Unsubscriptions = topics
	return testToken{Failure: c.failures["Unsubscribe"]}
}

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

// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package mqtt

import (
	"fmt"
	"testing"
	"time"

	MQTT "git.eclipse.org/gitroot/paho/org.eclipse.paho.mqtt.golang.git"
	"github.com/TheThingsNetwork/ttn/utils/pointer"
	. "github.com/TheThingsNetwork/ttn/utils/testing"
)

// ----- TYPE utilities

// MockHandler provides a fake implementation of a mqtt.Handler
type MockHandler struct {
	Failures  map[string]error
	OutTopic  string
	InMessage MQTT.Message
}

func NewMockHandler() *MockHandler {
	return &MockHandler{
		Failures: make(map[string]error),
		OutTopic: "MockTopic",
	}
}

func (h *MockHandler) Topic() string {
	return h.OutTopic
}

func (h *MockHandler) Handle(client Client, chpkt chan<- PktReq, chreg chan<- RegReq, msg MQTT.Message) error {
	h.InMessage = msg
	return h.Failures["Handle"]
}

// MockToken gives a fake implementation of MQTT.Token
//
// Provide a failure if you need to simulate an Error() result.
type MockToken struct {
	MQTT.Token
	Failure *string
}

func (t MockToken) Wait() bool {
	return true
}

func (t MockToken) WaitTimeout(d time.Duration) bool {
	<-time.After(d)
	return true
}

func (t MockToken) Error() error {
	if t.Failure == nil {
		return nil
	}
	return fmt.Errorf(*t.Failure)
}

// MockMessage gives a fake implementation of MQTT.Message
//
// provide payload and topic. Other methods are constants.
type MockMessage struct {
	payload interface{}
	topic   string
}

func (m MockMessage) Duplicate() bool {
	return false
}
func (m MockMessage) Qos() byte {
	return 2
}
func (m MockMessage) Retained() bool {
	return false
}
func (m MockMessage) Topic() string {
	return m.topic
}
func (m MockMessage) MessageID() uint16 {
	return 0
}
func (m MockMessage) Payload() []byte {
	switch m.payload.(type) {
	case []byte:
		return m.payload.([]byte)
	default:
		return nil
	}
}

// MockClient gives a fake implementation of a MQTT.ClientInt
//
// It saves the last subscription, unsubscriptions and publication call made
//
// It can also fails on demand (use the newMockClient method to define which methods should fail)
type MockClient struct {
	InSubscribe         *string
	InPublish           MQTT.Message
	InUnsubscribe       []string
	InSubscribeCallBack func(c Client, m MQTT.Message)

	Failures  map[string]*string
	connected bool
}

func NewMockClient(failures ...string) *MockClient {
	client := MockClient{Failures: make(map[string]*string), connected: true, InPublish: MockMessage{}}

	isFailure := func(x string) bool {
		for _, f := range failures {
			if f == x {
				return true
			}
		}
		return false
	}

	if isFailure("Connect") {
		client.Failures["Connect"] = pointer.String("MockError -> Failed to connect")
	}

	if isFailure("Publish") {
		client.Failures["Publish"] = pointer.String("MockError -> Failed to publish")
	}

	if isFailure("Subscribe") {
		client.Failures["Subscribe"] = pointer.String("MockError -> Failed to subscribe")
	}

	if isFailure("Unsubscribe") {
		client.Failures["Unsubscribe"] = pointer.String("MockError -> Failed to unsubscribe")
	}

	return &client
}

func (c *MockClient) Connect() MQTT.Token {
	c.connected = true
	return MockToken{Failure: c.Failures["Connect"]}
}

func (c *MockClient) Disconnect(quiesce uint) {
	<-time.After(time.Duration(quiesce))
	c.connected = false
	return
}

func (c MockClient) IsConnected() bool {
	return c.connected
}

func (c *MockClient) Publish(topic string, qos byte, retained bool, payload interface{}) MQTT.Token {
	c.InPublish = MockMessage{payload: payload, topic: topic}
	return MockToken{Failure: c.Failures["Publish"]}
}

func (c *MockClient) Subscribe(topic string, qos byte, callback func(c Client, m MQTT.Message)) MQTT.Token {
	c.InSubscribe = &topic
	c.InSubscribeCallBack = callback
	return MockToken{Failure: c.Failures["Subscribe"]}
}

func (c *MockClient) Unsubscribe(topics ...string) MQTT.Token {
	c.InUnsubscribe = topics
	return MockToken{Failure: c.Failures["Unsubscribe"]}
}

// ----- CHECK utilities

func checkSubscriptions(t *testing.T, want *string, got *string) {
	if got == nil {
		if want == nil {
			Ok(t, "Check Subscriptions")
			return
		}
		Ko(t, "Expected subscription to be %s but got nothing", *want)
	}
	if want == nil {
		Ko(t, "Expected no subscription but got %v", *got)
	}

	if *want != *got {
		Ko(t, "Expected subscription to be %s but got %s", *want, *got)
	}

	Ok(t, "Check Subscriptions")
}

// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package handlers

import (
	"fmt"
	"time"

	MQTT "github.com/KtorZ/paho.mqtt.golang"
	"github.com/TheThingsNetwork/ttn/utils/pointer"
)

// ----- TYPE utilities

// testToken gives a fake implementation of MQTT.Token
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
	MQTT.Client
	InSubscribe   *string
	InPublish     MQTT.Message
	InUnsubscribe []string

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

func (c *MockClient) Subscribe(topic string, qos byte, callback MQTT.MessageHandler) MQTT.Token {
	c.InSubscribe = &topic
	return MockToken{Failure: c.Failures["Subscribe"]}
}

func (c *MockClient) Unsubscribe(topics ...string) MQTT.Token {
	c.InUnsubscribe = topics
	return MockToken{Failure: c.Failures["Unsubscribe"]}
}

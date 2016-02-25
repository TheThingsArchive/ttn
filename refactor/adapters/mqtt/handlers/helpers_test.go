// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package handlers

import (
	"fmt"
	"time"

	MQTT "git.eclipse.org/gitroot/paho/org.eclipse.paho.mqtt.golang.git"
	core "github.com/TheThingsNetwork/ttn/refactor"
	. "github.com/TheThingsNetwork/ttn/refactor/adapters/mqtt"
	"github.com/TheThingsNetwork/ttn/utils/pointer"
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

func newTestClient(failConnect bool, failPublish bool, failSubscribe bool, failUnsubscribe bool) Client {
	client := testClient{failures: make(map[string]*string), connected: true}

	if failConnect {
		client.failures["Connect"] = pointer.String("MockError -> Failed to connect")
	}

	if failPublish {
		client.failures["Publish"] = pointer.String("MockError -> Failed to publish")
	}

	if failSubscribe {
		client.failures["Subscribe"] = pointer.String("MockError -> Failed to subscribe")
	}

	if failUnsubscribe {
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

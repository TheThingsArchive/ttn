// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package mqtt

import (
	"fmt"
	"time"

	MQTT "git.eclipse.org/gitroot/paho/org.eclipse.paho.mqtt.golang.git"
	"github.com/TheThingsNetwork/ttn/utils/errors"
)

// Client abstracts the interface of the paho client to allow an easier testing
type Client interface {
	Connect() MQTT.Token
	Disconnect(quiesce uint)
	IsConnected() bool
	Publish(topic string, qos byte, retained bool, payload interface{}) MQTT.Token
	Subscribe(topic string, qos byte, callback func(c Client, m MQTT.Message)) MQTT.Token
	Unsubscribe(topics ...string) MQTT.Token
}

// NewClient generates a new paho MQTT client from an id and a broker url
//
// The broker url is expected to contain a port if needed such as mybroker.com:87354
//
// The scheme has to be the same as the one used by the broker: tcp, tls or web socket
func NewClient(id string, broker string, scheme Scheme) (Client, error) {
	opts := MQTT.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("%s://%s", scheme, broker))
	opts.SetClientID(id)
	c := MQTT.NewClient(opts)
	if token := c.Connect(); token.Wait() && token.Error() != nil {
		return nil, errors.New(errors.Operational, token.Error())
	}
	return client{c}, nil
}

// Type client abstract the MQTT client to provide a valid interface.
// This is needed because of the signature of the Subscribe() method which is using a plain MQTT
// type instead of an interface.
type client struct {
	*MQTT.Client
}

// Subscribe just implements the interface and forward the call to the actual MQTT client
func (c client) Subscribe(topic string, qos byte, callback func(c Client, m MQTT.Message)) MQTT.Token {
	return c.Client.Subscribe(topic, qos, func(c *MQTT.Client, m MQTT.Message) {
		callback(client{c}, m)
	})
}

// Disconnect implements the interface and ensure that the disconnection won't panic if already
// closed
func (c client) Disconnect(quiesce uint) {
	go func() {
		defer func() { recover() }()
		c.Client.Disconnect(quiesce)
	}()
	<-time.After(time.Millisecond * time.Duration(quiesce))
}

// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package amqp

import (
	"fmt"
	"time"

	AMQP "github.com/streadway/amqp"
)

// Publisher connects to the AMQP server and can publish on uplink and activations from devices
type Publisher interface {
	Connect() error
	Disconnect()
	IsConnected() bool

	PublishUplink(payload UplinkMessage) error
}

// DefaultPublisher is the default AMQP client for The Things Network
type DefaultPublisher struct {
	url      string
	ctx      Logger
	conn     *AMQP.Connection
	channel  *AMQP.Channel
	exchange string
}

var (
	// ConnectRetries says how many times the client should retry a failed connection
	ConnectRetries = 10
	// ConnectRetryDelay says how long the client should wait between retries
	ConnectRetryDelay = time.Second
)

// NewPublisher creates a new DefaultPublisher
func NewPublisher(ctx Logger, url, exchange string) Publisher {
	if ctx == nil {
		ctx = &noopLogger{}
	}
	return &DefaultPublisher{
		ctx:      ctx,
		url:      url,
		exchange: exchange,
	}
}

// Connect to the MQTT broker. It will retry for ConnectRetries times with a delay of ConnectRetryDelay between retries
func (c *DefaultPublisher) Connect() error {
	if c.IsConnected() {
		return nil
	}
	var err error
	var conn *AMQP.Connection
	for retries := 0; retries < ConnectRetries; retries++ {
		conn, err = AMQP.Dial(c.url)
		if err == nil {
			break
		}
		c.ctx.Warnf("Could not connect to AMQP Broker (%s). Retrying...", err.Error())
		<-time.After(ConnectRetryDelay)
	}
	if err != nil {
		return fmt.Errorf("Could not connect to AMQP Broker (%s).", err)
	}
	channel, err := conn.Channel()
	if err != nil {
		conn.Close()
		return fmt.Errorf("Could not get AMQP channel (%s).", err)
	}
	if err = channel.ExchangeDeclare(c.exchange, "topic", true, false, false, false, nil); err != nil {
		channel.Close()
		conn.Close()
		return fmt.Errorf("Could not AMQP exchange (%s).", err)
	}
	c.conn = conn
	c.channel = channel
	return nil
}

func (c *DefaultPublisher) publish(key string, msg []byte, timestamp time.Time) error {
	return c.channel.Publish(c.exchange, key, false, false, AMQP.Publishing{
		ContentType: "application/json",
		Timestamp:   timestamp,
		Body:        msg,
	})
}

// Disconnect from the AMQP broker
func (c *DefaultPublisher) Disconnect() {
	if !c.IsConnected() {
		return
	}
	c.ctx.Debug("Disconnecting from AMQP")
	c.channel.Close()
	c.channel = nil
	c.conn.Close()
	c.conn = nil
}

// IsConnected returns true if there is a connection to the AMQP broker
func (c *DefaultPublisher) IsConnected() bool {
	return c.conn != nil
}

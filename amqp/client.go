// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package amqp

import (
	"fmt"
	"time"

	AMQP "github.com/streadway/amqp"
)

// Client connects to an AMQP server
type Client interface {
	Connect() error
	Disconnect()
	IsConnected() bool
	NewPublisher(exchange string) Publisher
}

// DefaultClient is the default AMQP client for The Things Network
type DefaultClient struct {
	url  string
	ctx  Logger
	conn *AMQP.Connection
}

var (
	// ConnectRetries says how many times the client should retry a failed connection
	ConnectRetries = 10
	// ConnectRetryDelay says how long the client should wait between retries
	ConnectRetryDelay = time.Second
)

// NewClient creates a new DefaultClient
func NewClient(ctx Logger, username, password, host string) Client {
	if ctx == nil {
		ctx = &noopLogger{}
	}
	credentials := "guest:guest"
	if username != "" {
		if password != "" {
			credentials = fmt.Sprintf("%s:%s", username, password)
		} else {
			credentials = username
		}
	}
	return &DefaultClient{
		ctx: ctx,
		url: fmt.Sprintf("amqp://%s@%s", credentials, host),
	}
}

// Connect to the AMQP server. It will retry for ConnectRetries times with a delay of ConnectRetryDelay between retries
func (c *DefaultClient) Connect() error {
	var err error
	var conn *AMQP.Connection
	for retries := 0; retries < ConnectRetries; retries++ {
		conn, err = AMQP.Dial(c.url)
		if err == nil {
			break
		}
		c.ctx.Warnf("Could not connect to AMQP server (%s). Retrying...", err.Error())
		<-time.After(ConnectRetryDelay)
	}
	if err != nil {
		return fmt.Errorf("Could not connect to AMQP server (%s).", err)
	}

	c.conn = conn
	return nil
}

// Disconnect from the AMQP server
func (c *DefaultClient) Disconnect() {
	if !c.IsConnected() {
		return
	}
	c.ctx.Debug("Disconnecting from AMQP")
	if err := c.conn.Close(); err != nil {
		c.ctx.Warnf("Could not close AMQP connection (%s)", err)
	}
	c.conn = nil
}

// IsConnected returns true if there is a connection to the AMQP server.
func (c *DefaultClient) IsConnected() bool {
	return c.conn != nil
}

// NewPublisher returns a new publisher
func (c *DefaultClient) NewPublisher(exchange string) Publisher {
	return &DefaultPublisher{
		ctx:      c.ctx,
		conn:     c.conn,
		exchange: exchange,
	}
}

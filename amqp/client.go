// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package amqp

import (
	"fmt"
	"io"
	"sync"
	"time"

	AMQP "github.com/streadway/amqp"
)

// ChannelUser represents a user of an AMQP channel, for example a Publisher
type ChannelUser interface {
	Open() error
	io.Closer
}

// Client connects to an AMQP server
type Client interface {
	Connect() error
	Disconnect()
	IsConnected() bool

	NewTopicPublisher(exchange string) Publisher
}

// DefaultClient is the default AMQP client for The Things Network
type DefaultClient struct {
	url          string
	ctx          Logger
	conn         *AMQP.Connection
	mutex        *sync.Mutex
	closed       chan *AMQP.Error
	channels     map[ChannelUser]*AMQP.Channel
	reconnecting bool
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
		ctx:      ctx,
		url:      fmt.Sprintf("amqp://%s@%s", credentials, host),
		mutex:    &sync.Mutex{},
		channels: make(map[ChannelUser]*AMQP.Channel),
	}
}

// Connect to the AMQP server. It will retry for ConnectRetries times with a delay of ConnectRetryDelay between retries
func (c *DefaultClient) Connect() error {
	var err error
	var conn *AMQP.Connection
	for retries := 0; c.reconnecting || retries < ConnectRetries; retries++ {
		conn, err = AMQP.Dial(c.url)
		if err == nil {
			break
		}
		c.ctx.Warnf("Could not connect to AMQP server (%s). Retrying attempt %d, reconnect is %v...", err.Error(), retries+1, c.reconnecting)
		<-time.After(ConnectRetryDelay)
	}
	if err != nil {
		return fmt.Errorf("Could not connect to AMQP server (%s)", err)
	}

	c.closed = make(chan *AMQP.Error)
	conn.NotifyClose(c.closed)
	go func(errc chan *AMQP.Error) {
		err := <-errc
		if err != nil {
			c.ctx.Warnf("Connection closed (%s). Reconnecting...", err)
			c.reconnecting = true
			c.Connect()
		} else {
			c.ctx.Info("Connection closed")
		}
	}(c.closed)

	c.conn = conn
	c.reconnecting = false

	c.mutex.Lock()
	defer c.mutex.Unlock()
	for user, channel := range c.channels {
		channel.Close()
		go user.Open()
	}

	return nil
}

// Disconnect from the AMQP server
func (c *DefaultClient) Disconnect() {
	if !c.IsConnected() {
		return
	}

	c.ctx.Debug("Disconnecting from AMQP")

	c.mutex.Lock()
	defer c.mutex.Unlock()
	for user, channel := range c.channels {
		channel.Close()
		delete(c.channels, user)
	}

	c.reconnecting = false
	c.conn.Close()
	c.conn = nil
}

// IsConnected returns true if there is a connection to the AMQP server.
func (c *DefaultClient) IsConnected() bool {
	return c.conn != nil
}

func (c *DefaultClient) openChannel(u ChannelUser) (*AMQP.Channel, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	channel, err := c.conn.Channel()
	if err != nil {
		return nil, err
	}
	c.channels[u] = channel

	return channel, nil
}

func (c *DefaultClient) closeChannel(u ChannelUser) error {
	channel, ok := c.channels[u]
	if !ok {
		return nil
	}
	err := channel.Close()

	c.mutex.Lock()
	defer c.mutex.Unlock()
	delete(c.channels, u)

	return err
}

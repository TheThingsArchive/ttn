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

// Client connects to an AMQP server
type Client interface {
	Connect() error
	Disconnect()
	IsConnected() bool

	NewPublisher(exchange, exchangeType string) Publisher
	NewSubscriber(exchange, exchangeType, name string, durable, autoDelete bool) Subscriber
}

// DefaultClient is the default AMQP client for The Things Network
type DefaultClient struct {
	url      string
	ctx      Logger
	conn     *AMQP.Connection
	mutex    *sync.Mutex
	channels map[*DefaultChannelClient]*AMQP.Channel
}

// ChannelClient represents a AMQP channel client
type ChannelClient interface {
	Open() error
	io.Closer
}

// DefaultChannelClient represents the default client of an AMQP channel
type DefaultChannelClient struct {
	ctx          Logger
	client       *DefaultClient
	channel      *AMQP.Channel
	exchange     string
	exchangeType string
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
		channels: make(map[*DefaultChannelClient]*AMQP.Channel),
	}
}

func (c *DefaultClient) connect(reconnect bool) (chan *AMQP.Error, error) {
	var err error
	var conn *AMQP.Connection
	for retries := 0; reconnect || retries < ConnectRetries; retries++ {
		conn, err = AMQP.Dial(c.url)
		if err == nil {
			break
		}
		c.ctx.Warnf("Could not connect to AMQP server (%s). Retry attempt %d, reconnect is %v...", err.Error(), retries+1, reconnect)
		<-time.After(ConnectRetryDelay)
	}
	if err != nil {
		return nil, fmt.Errorf("Could not connect to AMQP server (%s)", err)
	}

	closed := make(chan *AMQP.Error)
	conn.NotifyClose(closed)
	go func() {
		err := <-closed
		if err != nil {
			c.ctx.Warnf("Connection closed (%s). Reconnecting...", err)
			c.connect(true)
		} else {
			c.ctx.Info("Connection closed")
		}
	}()

	c.conn = conn
	c.ctx.Info("Connected to AMQP")

	if reconnect {
		c.mutex.Lock()
		defer c.mutex.Unlock()
		for user, channel := range c.channels {
			channel.Close()
			go user.Open()
		}
	}

	return closed, nil
}

// Connect to the AMQP server. It will retry for ConnectRetries times with a delay of ConnectRetryDelay between retries
func (c *DefaultClient) Connect() error {
	_, err := c.connect(false)
	return err
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

	c.conn.Close()
	c.conn = nil
}

// IsConnected returns true if there is a connection to the AMQP server.
func (c *DefaultClient) IsConnected() bool {
	return c.conn != nil
}

func (c *DefaultClient) openChannel(u *DefaultChannelClient) (*AMQP.Channel, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	channel, err := c.conn.Channel()
	if err != nil {
		return nil, err
	}
	c.channels[u] = channel

	return channel, nil
}

func (c *DefaultClient) closeChannel(u *DefaultChannelClient) error {
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

// Open opens a new channel and declares the exchange
func (p *DefaultChannelClient) Open() error {
	channel, err := p.client.openChannel(p)
	if err != nil {
		return fmt.Errorf("Could not open AMQP channel (%s)", err)
	}

	if p.exchange != "" {
		if err := channel.ExchangeDeclare(p.exchange, p.exchangeType, true, false, false, false, nil); err != nil {
			return fmt.Errorf("Could not declare AMQP exchange (%s)", err)
		}
	}

	p.channel = channel
	return nil
}

// Close closes the channel
func (p *DefaultChannelClient) Close() error {
	return p.client.closeChannel(p)
}

// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package amqp

import (
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/TheThingsNetwork/go-utils/log"
	AMQP "github.com/streadway/amqp"
)

// Client connects to an AMQP server
type Client interface {
	Connect() error
	Disconnect()
	IsConnected() bool

	NewPublisher(exchange string) Publisher
	NewSubscriber(exchange, name string, durable, autoDelete bool) Subscriber
}

// DefaultClient is the default AMQP client for The Things Network
type DefaultClient struct {
	url      string
	ctx      log.Interface
	conn     *AMQP.Connection
	mutex    sync.Mutex
	channels map[*DefaultChannelClient]*AMQP.Channel
}

// ChannelClient represents an AMQP channel client
type ChannelClient interface {
	Open() error
	io.Closer
}

// ChannelClientUser is a user of a channel, e.g. a publisher or consumer
type channelClientUser interface {
	use(*AMQP.Channel) error
	close()
}

// DefaultChannelClient represents the default client of an AMQP channel
type DefaultChannelClient struct {
	ctx          log.Interface
	client       *DefaultClient
	channel      *AMQP.Channel
	usersMutex   sync.RWMutex
	users        []channelClientUser
	name         string
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
func NewClient(ctx log.Interface, username, password, host string) Client {
	if ctx == nil {
		ctx = log.Get()
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
			channel, err = c.conn.Channel()
			if err != nil {
				c.ctx.Warnf("Failed to reopen channel %s for %s (%s)", user.name, user.exchange, err)
				continue
			}
			c.ctx.Infof("Reopened channel %s for %s", user.name, user.exchange)
			user.channel = channel
			user.usersMutex.RLock()
			defer user.usersMutex.RUnlock()
			for _, channelUser := range user.users {
				if err := channelUser.use(channel); err != nil {
					c.ctx.WithError(err).Warnf("Failed to use channel (%s)", err)
				}
			}
			c.channels[user] = channel
		}
	}

	return closed, nil
}

// GetChannel gets a new AMQP channel
func (c *DefaultClient) GetChannel() (*AMQP.Channel, error) {
	return c.conn.Channel()
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

	p.channel = channel
	return nil
}

// Close closes the channel
func (p *DefaultChannelClient) Close() error {
	p.usersMutex.RLock()
	defer p.usersMutex.RUnlock()
	for _, user := range p.users {
		user.close()
	}
	return p.client.closeChannel(p)
}

func (p *DefaultChannelClient) addUser(u channelClientUser) {
	p.usersMutex.Lock()
	defer p.usersMutex.Unlock()
	p.users = append(p.users, u)
}

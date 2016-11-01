// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package amqp

import (
	"os"
	"testing"
	"time"

	. "github.com/TheThingsNetwork/ttn/utils/testing"
	. "github.com/smartystreets/assertions"
	AMQP "github.com/streadway/amqp"
)

var host string

func init() {
	host = os.Getenv("AMQP_ADDRESS")
	if host == "" {
		host = "localhost:5672"
	}
}

func TestNewClient(t *testing.T) {
	a := New(t)
	c := NewClient(GetLogger(t, "TestNewClient"), "guest", "guest", host)
	a.So(c, ShouldNotBeNil)
}

func TestConnect(t *testing.T) {
	a := New(t)
	c := NewClient(GetLogger(t, "TestConnect"), "guest", "guest", host)
	err := c.Connect()
	defer c.Disconnect()
	a.So(err, ShouldBeNil)

	// Connecting while already connected should not change anything
	err = c.Connect()
	defer c.Disconnect()
	a.So(err, ShouldBeNil)
}

func TestConnectInvalidAddress(t *testing.T) {
	a := New(t)
	ConnectRetries = 2
	ConnectRetryDelay = 50 * time.Millisecond
	c := NewClient(GetLogger(t, "TestConnectInvalidAddress"), "guest", "guest", "localhost:56720")
	err := c.Connect()
	defer c.Disconnect()
	a.So(err, ShouldNotBeNil)
}

func TestIsConnected(t *testing.T) {
	a := New(t)
	c := NewClient(GetLogger(t, "TestIsConnected"), "guest", "guest", host)

	a.So(c.IsConnected(), ShouldBeFalse)

	c.Connect()
	defer c.Disconnect()

	a.So(c.IsConnected(), ShouldBeTrue)
}

func TestDisconnect(t *testing.T) {
	a := New(t)
	c := NewClient(GetLogger(t, "TestDisconnect"), "guest", "guest", host)

	// Disconnecting when not connected should not change anything
	c.Disconnect()
	a.So(c.IsConnected(), ShouldBeFalse)

	c.Connect()
	defer c.Disconnect()
	c.Disconnect()

	a.So(c.IsConnected(), ShouldBeFalse)
}

func TestReopenChannelClient(t *testing.T) {
	a := New(t)
	ctx := GetLogger(t, "TestReopenChannelClient")
	c := NewClient(ctx, "guest", "guest", host).(*DefaultClient)
	closed, err := c.connect(false)
	a.So(err, ShouldBeNil)
	defer c.Disconnect()

	p := &DefaultChannelClient{
		ctx:    ctx,
		client: c,
	}
	err = p.Open()
	a.So(err, ShouldBeNil)
	defer p.Close()

	test := func() error {
		return p.channel.Publish("", "test", false, false, AMQP.Publishing{
			Body: []byte("test"),
		})
	}

	// First attempt should be OK
	err = test()
	a.So(err, ShouldBeNil)

	// Make sure that the old channel is closed
	p.channel.Close()

	// Simulate a connection close so a new channel should be opened
	closed <- AMQP.ErrClosed

	// Give the reconnect some time
	time.Sleep(100 * time.Millisecond)

	// Second attempt should be OK as well and will only work on a new channel
	err = test()
	a.So(err, ShouldBeNil)
}

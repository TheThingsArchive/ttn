// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package amqp

import (
	"os"
	"testing"
	"time"

	"github.com/TheThingsNetwork/ttn/core/types"
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
	c := NewClient(getLogger(t, "TestNewClient"), "guest", "guest", host)
	a.So(c, ShouldNotBeNil)
}

func TestConnect(t *testing.T) {
	a := New(t)
	c := NewClient(getLogger(t, "TestConnect"), "guest", "guest", host)
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
	c := NewClient(getLogger(t, "TestConnectInvalidAddress"), "guest", "guest", "localhost:56720")
	err := c.Connect()
	defer c.Disconnect()
	a.So(err, ShouldNotBeNil)
}

func TestIsConnected(t *testing.T) {
	a := New(t)
	c := NewClient(getLogger(t, "TestIsConnected"), "guest", "guest", host)

	a.So(c.IsConnected(), ShouldBeFalse)

	c.Connect()
	defer c.Disconnect()

	a.So(c.IsConnected(), ShouldBeTrue)
}

func TestDisconnect(t *testing.T) {
	a := New(t)
	c := NewClient(getLogger(t, "TestDisconnect"), "guest", "guest", host)

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
	ctx := getLogger(t, "TestReopenChannelClient")
	c := NewClient(ctx, "guest", "guest", host).(*DefaultClient)
	closed, err := c.connect(false)
	a.So(err, ShouldBeNil)
	defer c.Disconnect()

	publisher := c.NewPublisher("amq.topic")
	err = publisher.Open()
	a.So(err, ShouldBeNil)
	defer publisher.Close()

	subscriber := c.NewSubscriber("amq.topic", "", false, false)
	err = subscriber.Open()
	a.So(err, ShouldBeNil)
	defer subscriber.Close()

	downs := make(chan types.DownlinkMessage, 1)
	err = subscriber.SubscribeDownlink(func(_ Subscriber, appID string, _ string, msg types.DownlinkMessage) {
		a.So(appID, ShouldEqual, "app")
		ctx.Debugf("Got downlink message")
		downs <- msg
	})
	a.So(err, ShouldBeNil)

	test := func() {
		ctx.Debug("Testing publish")
		err := publisher.PublishDownlink(types.DownlinkMessage{
			AppID: "app",
		})
		a.So(err, ShouldBeNil)
		select {
		case <-downs:
		case <-time.After(100 * time.Millisecond):
			panic("Published message didn't come in in time")
		}
		return
	}

	// First attempt should be OK
	test()

	// Make sure that the old channel is closed
	publisher.(*DefaultPublisher).channel.Close()
	subscriber.(*DefaultSubscriber).channel.Close()

	// Simulate a connection close so a new channel should be opened
	closed <- AMQP.ErrClosed

	// Give the reconnect some time
	time.Sleep(100 * time.Millisecond)

	// Second attempt should be OK as well and will only work on a new channel
	test()
	a.So(err, ShouldBeNil)
}

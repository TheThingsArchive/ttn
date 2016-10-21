// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package amqp

import (
	"testing"
	"time"

	"github.com/TheThingsNetwork/ttn/core/types"
	. "github.com/TheThingsNetwork/ttn/utils/testing"
	. "github.com/smartystreets/assertions"
	AMQP "github.com/streadway/amqp"
)

func TestOpenPublisher(t *testing.T) {
	a := New(t)
	c := NewClient(GetLogger(t, "TestOpenPublisher"), "guest", "guest", host)
	err := c.Connect()
	a.So(err, ShouldBeNil)
	defer c.Disconnect()

	p := c.NewTopicPublisher("test")
	err = p.Open()
	a.So(err, ShouldBeNil)
	defer p.Close()
}

func TestReopenPublisher(t *testing.T) {
	a := New(t)
	c := NewClient(GetLogger(t, "TestReopenPublisher"), "guest", "guest", host)
	err := c.Connect()
	a.So(err, ShouldBeNil)
	defer c.Disconnect()

	p := c.NewTopicPublisher("test")
	err = p.Open()
	a.So(err, ShouldBeNil)
	defer p.Close()

	// First attempt should be OK
	err = p.PublishUplink(types.UplinkMessage{})
	a.So(err, ShouldBeNil)

	// Make sure that the publisher's old channel is closed
	p.(*DefaultPublisher).channel.Close()

	// Simulate a connection close so a new channel should be opened
	c.(*DefaultClient).closed <- AMQP.ErrClosed

	// Give the reconnect some time
	time.Sleep(100 * time.Millisecond)

	// Second attempt should be OK as well and will only work on a new channel
	err = p.PublishUplink(types.UplinkMessage{})
	a.So(err, ShouldBeNil)
}

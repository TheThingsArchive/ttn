// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package amqp

import (
	"os"
	"testing"
	"time"

	. "github.com/TheThingsNetwork/ttn/utils/testing"
	. "github.com/smartystreets/assertions"
)

var host string

func init() {
	host = os.Getenv("AMQP_ADDR")
	if host == "" {
		host = "localhost:5672"
	}
}

func TestNewPublisher(t *testing.T) {
	a := New(t)
	c := NewPublisher(GetLogger(t, "TestNewPublisher"), "guest", "guest", host, "test")
	a.So(c, ShouldNotBeNil)
}

func TestConnect(t *testing.T) {
	a := New(t)
	c := NewPublisher(GetLogger(t, "TestConnect"), "guest", "guest", host, "test")
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
	c := NewPublisher(GetLogger(t, "TestConnectInvalidAddress"), "guest", "guest", "localhost:56720", "test")
	err := c.Connect()
	defer c.Disconnect()
	a.So(err, ShouldNotBeNil)
}

func TestIsConnected(t *testing.T) {
	a := New(t)
	c := NewPublisher(GetLogger(t, "TestIsConnected"), "guest", "guest", host, "test")

	a.So(c.IsConnected(), ShouldBeFalse)

	c.Connect()
	defer c.Disconnect()

	a.So(c.IsConnected(), ShouldBeTrue)
}

func TestDisconnect(t *testing.T) {
	a := New(t)
	c := NewPublisher(GetLogger(t, "TestDisconnect"), "guest", "guest", host, "test")

	// Disconnecting when not connected should not change anything
	c.Disconnect()
	a.So(c.IsConnected(), ShouldBeFalse)

	c.Connect()
	defer c.Disconnect()
	c.Disconnect()

	a.So(c.IsConnected(), ShouldBeFalse)
}

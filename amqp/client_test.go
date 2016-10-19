// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package amqp

import (
	"fmt"
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
		host = "localhost"
	}
}

func TestNewPublisher(t *testing.T) {
	a := New(t)
	c := NewPublisher(GetLogger(t, "TestNewPublisher"), fmt.Sprintf("amqp://guest:guest@%s:5672/", host), "test")
	a.So(c, ShouldNotBeNil)
}

func TestConnect(t *testing.T) {
	a := New(t)
	c := NewPublisher(GetLogger(t, "TestConnect"), fmt.Sprintf("amqp://guest:guest@%s:5672/", host), "test")
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
	c := NewPublisher(GetLogger(t, "TestConnectInvalidAddress"), fmt.Sprintf("amqp://guest:guest@%s:56720/", host), "test")
	err := c.Connect()
	defer c.Disconnect()
	a.So(err, ShouldNotBeNil)
}

func TestIsConnected(t *testing.T) {
	a := New(t)
	c := NewPublisher(GetLogger(t, "TestIsConnected"), fmt.Sprintf("amqp://guest:guest@%s:5672/", host), "test")

	a.So(c.IsConnected(), ShouldBeFalse)

	c.Connect()
	defer c.Disconnect()

	a.So(c.IsConnected(), ShouldBeTrue)
}

func TestDisconnect(t *testing.T) {
	a := New(t)
	c := NewPublisher(GetLogger(t, "TestDisconnect"), fmt.Sprintf("amqp://guest:guest@%s:5672/", host), "test")

	// Disconnecting when not connected should not change anything
	c.Disconnect()
	a.So(c.IsConnected(), ShouldBeFalse)

	c.Connect()
	defer c.Disconnect()
	c.Disconnect()

	a.So(c.IsConnected(), ShouldBeFalse)
}

// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package amqp

import (
	"testing"

	. "github.com/smartystreets/assertions"
)

func TestOpenPublisher(t *testing.T) {
	a := New(t)
	c := NewClient(getLogger(t, "TestOpenPublisher"), "guest", "guest", host)
	err := c.Connect()
	a.So(err, ShouldBeNil)
	defer c.Disconnect()

	p := c.NewPublisher("amq.topic")
	err = p.Open()
	a.So(err, ShouldBeNil)
	defer p.Close()
}

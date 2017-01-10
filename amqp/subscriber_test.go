// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package amqp

import (
	"testing"

	. "github.com/smartystreets/assertions"
)

func TestOpenSubscriber(t *testing.T) {
	a := New(t)
	c := NewClient(getLogger(t, "TestOpenSubscriber"), "guest", "guest", host)
	err := c.Connect()
	a.So(err, ShouldBeNil)
	defer c.Disconnect()

	s := c.NewSubscriber("amq.topic", "", false, true)
	err = s.Open()
	a.So(err, ShouldBeNil)
	defer s.Close()
}

// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package amqp

import (
	"sync"
	"testing"

	"github.com/TheThingsNetwork/ttn/core/types"
	. "github.com/smartystreets/assertions"
)

func TestPublishDownlink(t *testing.T) {
	a := New(t)
	c := NewClient(getLogger(t, "TestPublishDownlink"), "guest", "guest", host)
	err := c.Connect()
	a.So(err, ShouldBeNil)
	defer c.Disconnect()

	p := c.NewPublisher("")
	err = p.Open()
	a.So(err, ShouldBeNil)
	defer p.Close()

	err = p.PublishDownlink(types.DownlinkMessage{
		AppID:      "app",
		DevID:      "test",
		PayloadRaw: []byte{0x01, 0x08},
	})
	a.So(err, ShouldBeNil)
}

func TestSubscribeDownlink(t *testing.T) {
	a := New(t)
	c := NewClient(getLogger(t, "TestSubscribeDownlink"), "guest", "guest", host)
	err := c.Connect()
	a.So(err, ShouldBeNil)
	defer c.Disconnect()

	p := c.NewPublisher("amq.topic")
	err = p.Open()
	a.So(err, ShouldBeNil)
	defer p.Close()

	s := c.NewSubscriber("amq.topic", "", false, true)
	err = s.Open()
	a.So(err, ShouldBeNil)
	defer s.Close()

	wg := &sync.WaitGroup{}
	wg.Add(1)
	err = s.SubscribeDownlink(func(_ Subscriber, appID, devID string, req types.DownlinkMessage) {
		a.So(appID, ShouldEqual, "app")
		a.So(devID, ShouldEqual, "test")
		a.So(req.PayloadRaw, ShouldResemble, []byte{0x01, 0x08})
		wg.Done()
	})
	a.So(err, ShouldBeNil)

	err = p.PublishDownlink(types.DownlinkMessage{
		AppID:      "app",
		DevID:      "test",
		PayloadRaw: []byte{0x01, 0x08},
	})
	a.So(err, ShouldBeNil)

	wg.Wait()
}

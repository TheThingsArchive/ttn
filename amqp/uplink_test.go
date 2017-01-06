// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package amqp

import (
	"sync"
	"testing"

	"github.com/TheThingsNetwork/ttn/core/types"
	. "github.com/smartystreets/assertions"
)

func TestPublishUplink(t *testing.T) {
	a := New(t)
	c := NewClient(getLogger(t, "TestPublishUplink"), "guest", "guest", host)
	err := c.Connect()
	a.So(err, ShouldBeNil)
	defer c.Disconnect()

	p := c.NewPublisher("")
	err = p.Open()
	a.So(err, ShouldBeNil)
	defer p.Close()

	err = p.PublishUplink(types.UplinkMessage{
		AppID:      "app",
		DevID:      "test",
		PayloadRaw: []byte{0x01, 0x08},
	})
	a.So(err, ShouldBeNil)
}

func TestSubscribeUplink(t *testing.T) {
	a := New(t)
	c := NewClient(getLogger(t, "TestSubscribeUplink"), "guest", "guest", host)
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
	err = s.SubscribeUplink(func(_ Subscriber, appID, devID string, req types.UplinkMessage) {
		a.So(appID, ShouldEqual, "app")
		a.So(devID, ShouldEqual, "test")
		a.So(req.PayloadRaw, ShouldResemble, []byte{0x01, 0x08})
		wg.Done()
	})
	a.So(err, ShouldBeNil)

	err = p.PublishUplink(types.UplinkMessage{
		AppID:      "app",
		DevID:      "test",
		PayloadRaw: []byte{0x01, 0x08},
	})
	a.So(err, ShouldBeNil)

	wg.Wait()
}

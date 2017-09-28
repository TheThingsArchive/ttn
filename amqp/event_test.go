// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package amqp

import (
	"testing"
	"time"

	"github.com/TheThingsNetwork/ttn/core/types"
	. "github.com/TheThingsNetwork/ttn/utils/testing"
	. "github.com/smartystreets/assertions"
)

func TestPublishSubscribeAppEvents(t *testing.T) {
	a := New(t)
	c := NewClient(getLogger(t, "TestPublishSubscribeDeviceEvents"), "guest", "guest", host)
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

	var wg WaitGroup
	wg.Add(1)
	err = s.SubscribeAppEvents("app-id", "some-event",
		func(_ Subscriber, appID string, eventType types.EventType, payload []byte) {
			a.So(appID, ShouldEqual, "app-id")
			a.So(eventType, ShouldEqual, "some-event")
			a.So(string(payload), ShouldEqual, `"payload"`)
			wg.Done()
		})
	a.So(err, ShouldBeNil)
	p.PublishAppEvent("app-id", "some-event", "payload")
	err = wg.WaitFor(time.Millisecond * 100)
	a.So(err, ShouldBeNil)
}

func TestPublishSubscribeDeviceEvents(t *testing.T) {
	a := New(t)
	c := NewClient(getLogger(t, "TestPublishSubscribeDeviceEvents"), "guest", "guest", host)
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

	var wg WaitGroup
	wg.Add(1)
	err = s.SubscribeDeviceEvents("app-id", "dev-id", "some-event",
		func(subscriber Subscriber, appID string, devID string, event types.DeviceEvent) {
			a.So(appID, ShouldEqual, "app-id")
			a.So(devID, ShouldEqual, "dev-id")
			a.So(event.Event, ShouldEqual, "some-event")
			a.So(event.Data.(string), ShouldEqual, "payload")
			wg.Done()
		})
	a.So(err, ShouldBeNil)
	p.PublishDeviceEvent(types.DeviceEvent{
		AppID: "app-id",
		DevID: "dev-id",
		Event: "some-event",
		Data:  "payload",
	})
	err = wg.WaitFor(time.Millisecond * 200)
	a.So(err, ShouldBeNil)
}

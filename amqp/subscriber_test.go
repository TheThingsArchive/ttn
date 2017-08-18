// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package amqp

import (
	"sync"
	"testing"

	"github.com/TheThingsNetwork/ttn/core/types"
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

func TestQueueOps(t *testing.T) {
	a := New(t)
	c := NewClient(getLogger(t, "TestQueueOps"), "guest", "guest", host)
	err := c.Connect()
	a.So(err, ShouldBeNil)
	defer c.Disconnect()

	s := c.NewSubscriber("amq.topic", "", false, true)
	err = s.Open()
	a.So(err, ShouldBeNil)
	defer s.Close()

	name, err := s.QueueDeclare()
	a.So(err, ShouldBeNil)
	a.So(name, ShouldNotBeEmpty)

	key := "*"
	err = s.QueueBind(name, key)
	a.So(err, ShouldBeNil)
	defer s.QueueUnbind(name, key)
}

func TestSubscriberConsume(t *testing.T) {
	a := New(t)
	c := NewClient(getLogger(t, "TestSubscriberConsume"), "guest", "guest", host)
	err := c.Connect()
	a.So(err, ShouldBeNil)
	defer c.Disconnect()

	s := c.NewSubscriber("amq.topic", "", false, true)
	err = s.Open()
	a.So(err, ShouldBeNil)
	defer s.Close()

	name, err := s.QueueDeclare()
	a.So(err, ShouldBeNil)
	a.So(name, ShouldNotBeEmpty)

	key := DeviceKey{
		DevID: SimpleWildcard,
		AppID: "TestSubscriberConsume",
		Type:  Wildcard,
	}.String()
	err = s.QueueBind(name, key)
	a.So(err, ShouldBeNil)
	defer s.QueueUnbind(name, key)

	wg := sync.WaitGroup{}
	wg.Add(2)
	s.ConsumeMessages(name, func(subscriber Subscriber, appID string, devID string, req types.UplinkMessage) {
		a.So(appID, ShouldEqual, "TestSubscriberConsume")
		a.So(req.PayloadRaw, ShouldResemble, []byte("TestRaw"))
		wg.Done()
	}, func(subscriber Subscriber, appID string, devID string, req types.DeviceEvent) {
		a.So(devID, ShouldEqual, "TestSubscriberConsumeDevice")
		a.So(req.Event, ShouldEqual, types.ActivationEvent)
		wg.Done()
	})
	p := c.NewPublisher("amq.topic")
	err = p.Open()
	a.So(err, ShouldBeNil)
	defer p.Close()

	err = p.PublishUplink(types.UplinkMessage{
		AppID:      "TestSubscriberConsume",
		DevID:      "TestSubscriberConsumeDevice",
		PayloadRaw: []byte("TestRaw"),
	})
	a.So(err, ShouldBeNil)

	err = p.PublishDeviceEvent(types.DeviceEvent{
		AppID: "TestSubscriberConsume",
		DevID: "TestSubscriberConsumeDevice",
		Event: types.ActivationEvent,
	})
	a.So(err, ShouldBeNil)
	wg.Wait()
}

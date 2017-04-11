// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package handler

import (
	"os"
	"testing"
	"time"

	"github.com/TheThingsNetwork/ttn/amqp"
	"github.com/TheThingsNetwork/ttn/core/component"
	"github.com/TheThingsNetwork/ttn/core/handler/device"
	"github.com/TheThingsNetwork/ttn/core/types"
	. "github.com/TheThingsNetwork/ttn/utils/testing"
	. "github.com/smartystreets/assertions"
)

func TestHandleAMQP(t *testing.T) {
	host := os.Getenv("AMQP_ADDRESS")
	if host == "" {
		host = "localhost:5672"
	}

	a := New(t)
	var wg WaitGroup
	c := amqp.NewClient(GetLogger(t, "TestHandleAMQP"), "guest", "guest", host)
	err := c.Connect()
	a.So(err, ShouldBeNil)
	defer c.Disconnect()

	appID := "handler-amqp-app1"
	devID := "handler-amqp-dev1"
	h := &handler{
		Component: &component.Component{Ctx: GetLogger(t, "TestHandleAMQP")},
		devices:   device.NewRedisDeviceStore(GetRedisClient(), "handler-test-handle-amqp"),
	}
	h.WithAMQP("guest", "guest", host, "amq.topic")
	h.devices.Set(&device.Device{
		AppID: appID,
		DevID: devID,
	})
	defer func() {
		h.devices.Delete(appID, devID)
	}()
	err = h.HandleAMQP("guest", "guest", host, "amq.topic", "")
	a.So(err, ShouldBeNil)

	p := c.NewPublisher("amq.topic")
	err = p.Open()
	a.So(err, ShouldBeNil)
	defer p.Close()
	err = p.PublishDownlink(types.DownlinkMessage{
		AppID:      appID,
		DevID:      devID,
		PayloadRaw: []byte{0xAA, 0xBC},
	})
	a.So(err, ShouldBeNil)
	<-time.After(50 * time.Millisecond)
	q, _ := h.devices.DownlinkQueue(appID, devID)
	downlink, _ := q.Next()
	a.So(downlink, ShouldNotBeNil)

	{
		wg.Add(1)
		s := c.NewSubscriber("amq.topic", "", false, true)
		err = s.Open()
		a.So(err, ShouldBeNil)
		defer s.Close()
		err = s.SubscribeDeviceUplink(appID, devID, func(_ amqp.Subscriber, r_appID string, r_devID string, req types.UplinkMessage) {
			a.So(r_appID, ShouldEqual, appID)
			a.So(r_devID, ShouldEqual, devID)
			a.So(req.PayloadRaw, ShouldResemble, []byte{0xAA, 0xBC})
			wg.Done()
		})
		a.So(err, ShouldBeNil)

		h.amqpUp <- &types.UplinkMessage{
			DevID:      devID,
			AppID:      appID,
			PayloadRaw: []byte{0xAA, 0xBC},
			PayloadFields: map[string]interface{}{
				"field": "value",
			},
		}
	}

	// App Events
	{
		wg.Add(1)
		s := c.NewSubscriber("amq.topic", "", false, true)
		err = s.Open()
		a.So(err, ShouldBeNil)
		defer s.Close()
		err = s.SubscribeDeviceEvents(appID, devID, "", func(_ amqp.Subscriber, r_appID string, r_devID string, r_event types.EventType, evt []byte) {
			a.So(r_appID, ShouldEqual, appID)
			a.So(r_devID, ShouldEqual, devID)
			wg.Done()
		})
		a.So(err, ShouldBeNil)

		h.amqpEvent <- &types.DeviceEvent{
			DevID: devID,
			AppID: appID,
			Event: types.UplinkErrorEvent,
		}
	}

	// Device Events
	{
		wg.Add(1)
		s := c.NewSubscriber("amq.topic", "", false, true)
		err = s.Open()
		a.So(err, ShouldBeNil)
		defer s.Close()
		err = s.SubscribeAppEvents(appID, "", func(_ amqp.Subscriber, r_appID string, r_event types.EventType, evt []byte) {
			a.So(r_appID, ShouldEqual, appID)
			wg.Done()
		})
		a.So(err, ShouldBeNil)

		h.amqpEvent <- &types.DeviceEvent{
			AppID: appID,
			Event: types.UplinkErrorEvent,
		}
	}

	a.So(wg.WaitFor(200*time.Millisecond), ShouldBeNil)
}

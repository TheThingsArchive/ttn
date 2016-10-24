// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package mqtt

import (
	"fmt"
	"testing"
	"time"

	"github.com/TheThingsNetwork/ttn/core/types"
	. "github.com/TheThingsNetwork/ttn/utils/testing"
	. "github.com/smartystreets/assertions"
)

func TestPublishSubscribeAppEvents(t *testing.T) {
	a := New(t)
	c := NewClient(GetLogger(t, "Test"), "test", "", "", fmt.Sprintf("tcp://%s:1883", host))
	c.Connect()
	defer c.Disconnect()
	var wg WaitGroup
	wg.Add(1)
	subToken := c.SubscribeAppEvents("app-id", "", func(_ Client, appID string, eventType types.EventType, payload []byte) {
		a.So(appID, ShouldEqual, "app-id")
		a.So(eventType, ShouldEqual, "some-event")
		a.So(string(payload), ShouldEqual, `"payload"`)
		wg.Done()
	})
	waitForOK(subToken, a)
	pubToken := c.PublishAppEvent("app-id", "some-event", "payload")
	waitForOK(pubToken, a)
	unsubToken := c.UnsubscribeAppEvents("app-id", "")
	waitForOK(unsubToken, a)
	wg.WaitFor(100 * time.Millisecond)
}

func TestPublishSubscribeDeviceEvents(t *testing.T) {
	a := New(t)
	c := NewClient(GetLogger(t, "Test"), "test", "", "", fmt.Sprintf("tcp://%s:1883", host))
	c.Connect()
	defer c.Disconnect()
	var wg WaitGroup
	wg.Add(1)
	subToken := c.SubscribeDeviceEvents("app-id", "dev-id", "", func(_ Client, appID string, devID string, eventType types.EventType, payload []byte) {
		a.So(appID, ShouldEqual, "app-id")
		a.So(devID, ShouldEqual, "dev-id")
		a.So(eventType, ShouldEqual, "some-event")
		a.So(string(payload), ShouldEqual, `"payload"`)
		wg.Done()
	})
	waitForOK(subToken, a)
	pubToken := c.PublishDeviceEvent("app-id", "dev-id", "some-event", "payload")
	waitForOK(pubToken, a)
	unsubToken := c.UnsubscribeDeviceEvents("app-id", "dev-id", "")
	waitForOK(unsubToken, a)
	wg.WaitFor(100 * time.Millisecond)
}

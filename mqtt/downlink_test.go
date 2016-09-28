// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package mqtt

import (
	"fmt"
	"testing"
	"time"

	. "github.com/TheThingsNetwork/ttn/utils/testing"
	. "github.com/smartystreets/assertions"
)

// Downlink pub/sub

func TestPublishDownlink(t *testing.T) {
	a := New(t)
	c := NewClient(GetLogger(t, "Test"), "test", "", "", fmt.Sprintf("tcp://%s:1883", host))
	c.Connect()
	defer c.Disconnect()

	dataDown := DownlinkMessage{
		AppID:      "someid",
		DevID:      "someid",
		PayloadRaw: []byte{0x01, 0x02, 0x03, 0x04},
	}

	token := c.PublishDownlink(dataDown)
	waitForOK(token, a)

	a.So(token.Error(), ShouldBeNil)
}

func TestSubscribeDeviceDownlink(t *testing.T) {
	a := New(t)
	c := NewClient(GetLogger(t, "Test"), "test", "", "", fmt.Sprintf("tcp://%s:1883", host))
	c.Connect()
	defer c.Disconnect()

	token := c.SubscribeDeviceDownlink("someid", "someid", func(client Client, appID string, devID string, req DownlinkMessage) {

	})
	waitForOK(token, a)
	a.So(token.Error(), ShouldBeNil)

	token = c.UnsubscribeDeviceDownlink("someid", "someid")
	waitForOK(token, a)
	a.So(token.Error(), ShouldBeNil)
}

func TestSubscribeAppDownlink(t *testing.T) {
	a := New(t)
	c := NewClient(GetLogger(t, "Test"), "test", "", "", fmt.Sprintf("tcp://%s:1883", host))
	c.Connect()
	defer c.Disconnect()

	token := c.SubscribeAppDownlink("someid", func(client Client, appID string, devID string, req DownlinkMessage) {

	})
	waitForOK(token, a)
	a.So(token.Error(), ShouldBeNil)

	token = c.UnsubscribeAppDownlink("someid")
	waitForOK(token, a)
	a.So(token.Error(), ShouldBeNil)
}

func TestSubscribeDownlink(t *testing.T) {
	a := New(t)
	c := NewClient(GetLogger(t, "Test"), "test", "", "", fmt.Sprintf("tcp://%s:1883", host))
	c.Connect()
	defer c.Disconnect()

	token := c.SubscribeDownlink(func(client Client, appID string, devID string, req DownlinkMessage) {

	})
	waitForOK(token, a)
	a.So(token.Error(), ShouldBeNil)

	token = c.UnsubscribeDownlink()
	waitForOK(token, a)
	a.So(token.Error(), ShouldBeNil)
}

func TestPubSubDownlink(t *testing.T) {
	a := New(t)
	c := NewClient(GetLogger(t, "Test"), "test", "", "", fmt.Sprintf("tcp://%s:1883", host))
	c.Connect()
	defer c.Disconnect()

	var wg WaitGroup

	wg.Add(1)

	subToken := c.SubscribeDeviceDownlink("app3", "dev3", func(client Client, appID string, devID string, req DownlinkMessage) {
		a.So(appID, ShouldResemble, "app3")
		a.So(devID, ShouldResemble, "dev3")

		wg.Done()
	})
	waitForOK(subToken, a)

	pubToken := c.PublishDownlink(DownlinkMessage{
		AppID:      "app3",
		DevID:      "dev3",
		PayloadRaw: []byte{0x01, 0x02, 0x03, 0x04},
	})
	waitForOK(pubToken, a)

	a.So(wg.WaitFor(200*time.Millisecond), ShouldBeNil)

	unsubToken := c.UnsubscribeDeviceDownlink("app3", "dev3")
	waitForOK(unsubToken, a)
}

func TestPubSubAppDownlink(t *testing.T) {
	a := New(t)
	c := NewClient(GetLogger(t, "Test"), "test", "", "", fmt.Sprintf("tcp://%s:1883", host))
	c.Connect()
	defer c.Disconnect()

	var wg WaitGroup

	wg.Add(2)

	subToken := c.SubscribeAppDownlink("app4", func(client Client, appID string, devID string, req DownlinkMessage) {
		a.So(appID, ShouldResemble, "app4")
		a.So(req.PayloadRaw, ShouldResemble, []byte{0x01, 0x02, 0x03, 0x04})
		wg.Done()
	})
	waitForOK(subToken, a)

	pubToken := c.PublishDownlink(DownlinkMessage{
		AppID:      "app4",
		DevID:      "dev1",
		PayloadRaw: []byte{0x01, 0x02, 0x03, 0x04},
	})
	waitForOK(pubToken, a)
	pubToken = c.PublishDownlink(DownlinkMessage{
		AppID:      "app4",
		DevID:      "dev2",
		PayloadRaw: []byte{0x01, 0x02, 0x03, 0x04},
	})
	waitForOK(pubToken, a)

	a.So(wg.WaitFor(200*time.Millisecond), ShouldBeNil)

	unsubToken := c.UnsubscribeAppDownlink("app3")
	waitForOK(unsubToken, a)
}

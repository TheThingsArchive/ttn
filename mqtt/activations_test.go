// Copyright © 2017 The Things Network
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

// Activations pub/sub

func TestPublishActivations(t *testing.T) {
	a := New(t)
	c := NewClient(getLogger(t, "Test"), "test", "", "", fmt.Sprintf("tcp://%s", host))
	c.Connect()
	defer c.Disconnect()

	dataActivations := types.Activation{
		AppID:    "someid",
		DevID:    "someid",
		Metadata: types.Metadata{DataRate: "SF7BW125"},
	}

	token := c.PublishActivation(dataActivations)
	waitForOK(token, a)

	a.So(token.Error(), ShouldBeNil)
}

func TestSubscribeDeviceActivations(t *testing.T) {
	a := New(t)
	c := NewClient(getLogger(t, "Test"), "test", "", "", fmt.Sprintf("tcp://%s", host))
	c.Connect()
	defer c.Disconnect()

	token := c.SubscribeDeviceActivations("someid", "someid", func(client Client, appID string, devID string, req types.Activation) {

	})
	waitForOK(token, a)
	a.So(token.Error(), ShouldBeNil)

	token = c.UnsubscribeDeviceActivations("someid", "someid")
	waitForOK(token, a)
	a.So(token.Error(), ShouldBeNil)
}

func TestSubscribeAppActivations(t *testing.T) {
	a := New(t)
	c := NewClient(getLogger(t, "Test"), "test", "", "", fmt.Sprintf("tcp://%s", host))
	c.Connect()
	defer c.Disconnect()

	token := c.SubscribeAppActivations("someid", func(client Client, appID string, devID string, req types.Activation) {

	})
	waitForOK(token, a)
	a.So(token.Error(), ShouldBeNil)

	token = c.UnsubscribeAppActivations("someid")
	waitForOK(token, a)
	a.So(token.Error(), ShouldBeNil)
}

func TestSubscribeActivations(t *testing.T) {
	a := New(t)
	c := NewClient(getLogger(t, "Test"), "test", "", "", fmt.Sprintf("tcp://%s", host))
	c.Connect()
	defer c.Disconnect()

	token := c.SubscribeActivations(func(client Client, appID string, devID string, req types.Activation) {

	})
	waitForOK(token, a)
	a.So(token.Error(), ShouldBeNil)

	token = c.UnsubscribeActivations()
	waitForOK(token, a)
	a.So(token.Error(), ShouldBeNil)
}

func TestPubSubActivations(t *testing.T) {
	a := New(t)
	c := NewClient(getLogger(t, "Test"), "test", "", "", fmt.Sprintf("tcp://%s", host))
	c.Connect()
	defer c.Disconnect()

	var wg WaitGroup

	wg.Add(1)

	subToken := c.SubscribeDeviceActivations("app5", "dev1", func(client Client, appID string, devID string, req types.Activation) {
		a.So(appID, ShouldResemble, "app5")
		a.So(devID, ShouldResemble, "dev1")

		wg.Done()
	})
	waitForOK(subToken, a)

	pubToken := c.PublishActivation(types.Activation{
		AppID:    "app5",
		DevID:    "dev1",
		Metadata: types.Metadata{DataRate: "SF7BW125"},
	})
	waitForOK(pubToken, a)

	a.So(wg.WaitFor(200*time.Millisecond), ShouldBeNil)

	unsubToken := c.UnsubscribeDeviceActivations("app5", "dev1")
	waitForOK(unsubToken, a)
}

func TestPubSubAppActivations(t *testing.T) {
	a := New(t)
	c := NewClient(getLogger(t, "Test"), "test", "", "", fmt.Sprintf("tcp://%s", host))
	c.Connect()
	defer c.Disconnect()

	var wg WaitGroup

	wg.Add(2)

	subToken := c.SubscribeAppActivations("app6", func(client Client, appID string, devID string, req types.Activation) {
		a.So(appID, ShouldResemble, "app6")
		a.So(req.Metadata.DataRate, ShouldEqual, "SF7BW125")
		wg.Done()
	})
	waitForOK(subToken, a)

	pubToken := c.PublishActivation(types.Activation{
		AppID:    "app6",
		DevID:    "dev1",
		Metadata: types.Metadata{DataRate: "SF7BW125"},
	})
	waitForOK(pubToken, a)
	pubToken = c.PublishActivation(types.Activation{
		AppID:    "app6",
		DevID:    "dev2",
		Metadata: types.Metadata{DataRate: "SF7BW125"},
	})
	waitForOK(pubToken, a)

	a.So(wg.WaitFor(200*time.Millisecond), ShouldBeNil)

	unsubToken := c.UnsubscribeAppActivations("app6")
	waitForOK(unsubToken, a)
}

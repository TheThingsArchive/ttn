// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package mqtt

import (
	"fmt"
	"sync"
	"testing"
	"time"

	. "github.com/TheThingsNetwork/ttn/utils/testing"
	. "github.com/smartystreets/assertions"
)

func TestToken(t *testing.T) {
	a := New(t)

	okToken := simpleToken{}
	okToken.Wait()
	a.So(okToken.Error(), ShouldBeNil)

	failToken := simpleToken{fmt.Errorf("Err")}
	failToken.Wait()
	a.So(failToken.Error(), ShouldNotBeNil)
}

func TestNewClient(t *testing.T) {
	a := New(t)
	c := NewClient(GetLogger(t, "Test"), "test", "", "", "tcp://localhost:1883")
	a.So(c.(*defaultClient).mqtt, ShouldNotBeNil)
}

func TestConnect(t *testing.T) {
	a := New(t)
	c := NewClient(GetLogger(t, "Test"), "test", "", "", "tcp://localhost:1883")
	err := c.Connect()
	a.So(err, ShouldBeNil)

	// Connecting while already connected should not change anything
	err = c.Connect()
	a.So(err, ShouldBeNil)
}

func TestConnectInvalidAddress(t *testing.T) {
	a := New(t)
	ConnectRetries = 2
	ConnectRetryDelay = 50 * time.Millisecond
	c := NewClient(GetLogger(t, "Test"), "test", "", "", "tcp://localhost:18830") // No MQTT on 18830
	err := c.Connect()
	a.So(err, ShouldNotBeNil)
}

func TestConnectInvalidCredentials(t *testing.T) {
	t.Skipf("Need authenticated MQTT for TestConnectInvalidCredentials - Skipping")
}

func TestIsConnected(t *testing.T) {
	a := New(t)
	c := NewClient(GetLogger(t, "Test"), "test", "", "", "tcp://localhost:1883")

	a.So(c.IsConnected(), ShouldBeFalse)

	c.Connect()

	a.So(c.IsConnected(), ShouldBeTrue)
}

func TestDisconnect(t *testing.T) {
	a := New(t)
	c := NewClient(GetLogger(t, "Test"), "test", "", "", "tcp://localhost:1883")

	// Disconnecting when not connected should not change anything
	c.Disconnect()
	a.So(c.IsConnected(), ShouldBeFalse)

	c.Connect()
	c.Disconnect()

	a.So(c.IsConnected(), ShouldBeFalse)
}

func TestRandomTopicPublish(t *testing.T) {
	ctx := GetLogger(t, "TestRandomTopicPublish")

	c := NewClient(ctx, "test", "", "", "tcp://localhost:1883")
	c.Connect()

	c.(*defaultClient).mqtt.Subscribe("randomtopic", QoS, nil).Wait()
	c.(*defaultClient).mqtt.Publish("randomtopic", QoS, false, []byte{0x00}).Wait()

	<-time.After(50 * time.Millisecond)

	ctx.Info("This test should have printed one message.")
}

// Uplink pub/sub

func TestPublishUplink(t *testing.T) {
	a := New(t)
	c := NewClient(GetLogger(t, "Test"), "test", "", "", "tcp://localhost:1883")
	c.Connect()

	dataUp := UplinkMessage{
		Payload: []byte{0x01, 0x02, 0x03, 0x04},
	}

	token := c.PublishUplink("someid", "someid", dataUp)
	token.Wait()

	a.So(token.Error(), ShouldBeNil)
}

func TestSubscribeDeviceUplink(t *testing.T) {
	a := New(t)
	c := NewClient(GetLogger(t, "Test"), "test", "", "", "tcp://localhost:1883")
	c.Connect()

	token := c.SubscribeDeviceUplink("someid", "someid", func(client Client, appID string, devID string, req UplinkMessage) {

	})
	token.Wait()

	a.So(token.Error(), ShouldBeNil)
}

func TestSubscribeAppUplink(t *testing.T) {
	a := New(t)
	c := NewClient(GetLogger(t, "Test"), "test", "", "", "tcp://localhost:1883")
	c.Connect()

	token := c.SubscribeAppUplink("someid", func(client Client, appID string, devID string, req UplinkMessage) {

	})
	token.Wait()

	a.So(token.Error(), ShouldBeNil)
}

func TestSubscribeUplink(t *testing.T) {
	a := New(t)
	c := NewClient(GetLogger(t, "Test"), "test", "", "", "tcp://localhost:1883")
	c.Connect()

	token := c.SubscribeUplink(func(client Client, appID string, devID string, req UplinkMessage) {

	})
	token.Wait()

	a.So(token.Error(), ShouldBeNil)
}

func TestPubSubUplink(t *testing.T) {
	a := New(t)
	c := NewClient(GetLogger(t, "Test"), "test", "", "", "tcp://localhost:1883")
	c.Connect()

	var wg sync.WaitGroup

	wg.Add(1)

	c.SubscribeDeviceUplink("app1", "dev1", func(client Client, appID string, devID string, req UplinkMessage) {
		a.So(appID, ShouldResemble, "app1")
		a.So(devID, ShouldResemble, "dev1")

		wg.Done()
	}).Wait()

	c.PublishUplink("app1", "dev1", UplinkMessage{Payload: []byte{0x01, 0x02, 0x03, 0x04}}).Wait()

	wg.Wait()
}

func TestPubSubAppUplink(t *testing.T) {
	a := New(t)
	c := NewClient(GetLogger(t, "Test"), "test", "", "", "tcp://localhost:1883")
	c.Connect()

	var wg sync.WaitGroup

	wg.Add(2)

	c.SubscribeAppUplink("app2", func(client Client, appID string, devID string, req UplinkMessage) {
		a.So(appID, ShouldResemble, "app2")
		a.So(req.Payload, ShouldResemble, []byte{0x01, 0x02, 0x03, 0x04})
		wg.Done()
	}).Wait()

	c.PublishUplink("app2", "dev1", UplinkMessage{Payload: []byte{0x01, 0x02, 0x03, 0x04}}).Wait()
	c.PublishUplink("app2", "dev2", UplinkMessage{Payload: []byte{0x01, 0x02, 0x03, 0x04}}).Wait()

	wg.Wait()
}

// Downlink pub/sub

func TestPublishDownlink(t *testing.T) {
	a := New(t)
	c := NewClient(GetLogger(t, "Test"), "test", "", "", "tcp://localhost:1883")
	c.Connect()

	dataDown := DownlinkMessage{
		Payload: []byte{0x01, 0x02, 0x03, 0x04},
	}

	token := c.PublishDownlink("someid", "someid", dataDown)
	token.Wait()

	a.So(token.Error(), ShouldBeNil)
}

func TestSubscribeDeviceDownlink(t *testing.T) {
	a := New(t)
	c := NewClient(GetLogger(t, "Test"), "test", "", "", "tcp://localhost:1883")
	c.Connect()

	token := c.SubscribeDeviceDownlink("someid", "someid", func(client Client, appID string, devID string, req DownlinkMessage) {

	})
	token.Wait()

	a.So(token.Error(), ShouldBeNil)
}

func TestSubscribeAppDownlink(t *testing.T) {
	a := New(t)
	c := NewClient(GetLogger(t, "Test"), "test", "", "", "tcp://localhost:1883")
	c.Connect()

	token := c.SubscribeAppDownlink("someid", func(client Client, appID string, devID string, req DownlinkMessage) {

	})
	token.Wait()

	a.So(token.Error(), ShouldBeNil)
}

func TestSubscribeDownlink(t *testing.T) {
	a := New(t)
	c := NewClient(GetLogger(t, "Test"), "test", "", "", "tcp://localhost:1883")
	c.Connect()

	token := c.SubscribeDownlink(func(client Client, appID string, devID string, req DownlinkMessage) {

	})
	token.Wait()

	a.So(token.Error(), ShouldBeNil)
}

func TestPubSubDownlink(t *testing.T) {
	a := New(t)
	c := NewClient(GetLogger(t, "Test"), "test", "", "", "tcp://localhost:1883")
	c.Connect()

	var wg sync.WaitGroup

	wg.Add(1)

	c.SubscribeDeviceDownlink("app3", "dev3", func(client Client, appID string, devID string, req DownlinkMessage) {
		a.So(appID, ShouldResemble, "app3")
		a.So(devID, ShouldResemble, "dev3")

		wg.Done()
	}).Wait()

	c.PublishDownlink("app3", "dev3", DownlinkMessage{Payload: []byte{0x01, 0x02, 0x03, 0x04}}).Wait()

	wg.Wait()
}

func TestPubSubAppDownlink(t *testing.T) {
	a := New(t)
	c := NewClient(GetLogger(t, "Test"), "test", "", "", "tcp://localhost:1883")
	c.Connect()

	var wg sync.WaitGroup

	wg.Add(2)

	c.SubscribeAppDownlink("app4", func(client Client, appID string, devID string, req DownlinkMessage) {
		a.So(appID, ShouldResemble, "app4")
		a.So(req.Payload, ShouldResemble, []byte{0x01, 0x02, 0x03, 0x04})
		wg.Done()
	}).Wait()

	c.PublishDownlink("app4", "dev1", DownlinkMessage{Payload: []byte{0x01, 0x02, 0x03, 0x04}}).Wait()
	c.PublishDownlink("app4", "dev2", DownlinkMessage{Payload: []byte{0x01, 0x02, 0x03, 0x04}}).Wait()

	wg.Wait()
}

// Activations pub/sub

func TestPublishActivations(t *testing.T) {
	a := New(t)
	c := NewClient(GetLogger(t, "Test"), "test", "", "", "tcp://localhost:1883")
	c.Connect()

	dataActivations := Activation{Metadata: []Metadata{Metadata{DataRate: "SF7BW125"}}}

	token := c.PublishActivation("someid", "someid", dataActivations)
	token.Wait()

	a.So(token.Error(), ShouldBeNil)
}

func TestSubscribeDeviceActivations(t *testing.T) {
	a := New(t)
	c := NewClient(GetLogger(t, "Test"), "test", "", "", "tcp://localhost:1883")
	c.Connect()

	token := c.SubscribeDeviceActivations("someid", "someid", func(client Client, appID string, devID string, req Activation) {

	})
	token.Wait()

	a.So(token.Error(), ShouldBeNil)
}

func TestSubscribeAppActivations(t *testing.T) {
	a := New(t)
	c := NewClient(GetLogger(t, "Test"), "test", "", "", "tcp://localhost:1883")
	c.Connect()

	token := c.SubscribeAppActivations("someid", func(client Client, appID string, devID string, req Activation) {

	})
	token.Wait()

	a.So(token.Error(), ShouldBeNil)
}

func TestSubscribeActivations(t *testing.T) {
	a := New(t)
	c := NewClient(GetLogger(t, "Test"), "test", "", "", "tcp://localhost:1883")
	c.Connect()

	token := c.SubscribeActivations(func(client Client, appID string, devID string, req Activation) {

	})
	token.Wait()

	a.So(token.Error(), ShouldBeNil)
}

func TestPubSubActivations(t *testing.T) {
	a := New(t)
	c := NewClient(GetLogger(t, "Test"), "test", "", "", "tcp://localhost:1883")
	c.Connect()

	var wg sync.WaitGroup

	wg.Add(1)

	c.SubscribeDeviceActivations("app5", "dev1", func(client Client, appID string, devID string, req Activation) {
		a.So(appID, ShouldResemble, "app5")
		a.So(devID, ShouldResemble, "dev1")

		wg.Done()
	}).Wait()

	c.PublishActivation("app5", "dev1", Activation{Metadata: []Metadata{Metadata{DataRate: "SF7BW125"}}}).Wait()

	wg.Wait()
}

func TestPubSubAppActivations(t *testing.T) {
	a := New(t)
	c := NewClient(GetLogger(t, "Test"), "test", "", "", "tcp://localhost:1883")
	c.Connect()

	var wg sync.WaitGroup

	wg.Add(2)

	c.SubscribeAppActivations("app6", func(client Client, appID string, devID string, req Activation) {
		a.So(appID, ShouldResemble, "app6")
		a.So(req.Metadata[0].DataRate, ShouldEqual, "SF7BW125")
		wg.Done()
	}).Wait()

	c.PublishActivation("app6", "dev1", Activation{Metadata: []Metadata{Metadata{DataRate: "SF7BW125"}}}).Wait()
	c.PublishActivation("app6", "dev2", Activation{Metadata: []Metadata{Metadata{DataRate: "SF7BW125"}}}).Wait()

	wg.Wait()
}

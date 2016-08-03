// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package mqtt

import (
	"fmt"
	"os"
	"testing"
	"time"

	. "github.com/TheThingsNetwork/ttn/utils/testing"
	"github.com/apex/log"
	. "github.com/smartystreets/assertions"
)

var host string

func init() {
	host = os.Getenv("MQTT_HOST")
	if host == "" {
		host = "localhost"
	}
}

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
	c := NewClient(GetLogger(t, "Test"), "test", "", "", fmt.Sprintf("tcp://%s:1883", host))
	a.So(c.(*DefaultClient).mqtt, ShouldNotBeNil)
}

func TestConnect(t *testing.T) {
	a := New(t)
	c := NewClient(GetLogger(t, "Test"), "test", "", "", fmt.Sprintf("tcp://%s:1883", host))
	err := c.Connect()
	defer c.Disconnect()
	a.So(err, ShouldBeNil)

	// Connecting while already connected should not change anything
	err = c.Connect()
	defer c.Disconnect()
	a.So(err, ShouldBeNil)
}

func TestConnectInvalidAddress(t *testing.T) {
	a := New(t)
	ConnectRetries = 2
	ConnectRetryDelay = 50 * time.Millisecond
	c := NewClient(GetLogger(t, "Test"), "test", "", "", "tcp://localhost:18830") // No MQTT on 18830
	err := c.Connect()
	defer c.Disconnect()
	a.So(err, ShouldNotBeNil)
}

func TestConnectInvalidCredentials(t *testing.T) {
	t.Skipf("Need authenticated MQTT for TestConnectInvalidCredentials - Skipping")
}

func TestIsConnected(t *testing.T) {
	a := New(t)
	c := NewClient(GetLogger(t, "Test"), "test", "", "", fmt.Sprintf("tcp://%s:1883", host))

	a.So(c.IsConnected(), ShouldBeFalse)

	c.Connect()
	defer c.Disconnect()

	a.So(c.IsConnected(), ShouldBeTrue)
}

func TestDisconnect(t *testing.T) {
	a := New(t)
	c := NewClient(GetLogger(t, "Test"), "test", "", "", fmt.Sprintf("tcp://%s:1883", host))

	// Disconnecting when not connected should not change anything
	c.Disconnect()
	a.So(c.IsConnected(), ShouldBeFalse)

	c.Connect()
	defer c.Disconnect()
	c.Disconnect()

	a.So(c.IsConnected(), ShouldBeFalse)
}

func TestRandomTopicPublish(t *testing.T) {
	ctx := GetLogger(t, "TestRandomTopicPublish")

	c := NewClient(ctx, "test", "", "", fmt.Sprintf("tcp://%s:1883", host))
	c.Connect()
	defer c.Disconnect()

	c.(*DefaultClient).mqtt.Subscribe("randomtopic", QoS, nil).Wait()
	c.(*DefaultClient).mqtt.Publish("randomtopic", QoS, false, []byte{0x00}).Wait()

	<-time.After(50 * time.Millisecond)

	ctx.Info("This test should have printed one message.")
}

// Uplink pub/sub

func TestPublishUplink(t *testing.T) {
	a := New(t)
	c := NewClient(GetLogger(t, "Test"), "test", "", "", fmt.Sprintf("tcp://%s:1883", host))
	c.Connect()
	defer c.Disconnect()

	dataUp := UplinkMessage{
		AppID:   "someid",
		DevID:   "someid",
		Payload: []byte{0x01, 0x02, 0x03, 0x04},
	}

	token := c.PublishUplink(dataUp)
	token.Wait()

	a.So(token.Error(), ShouldBeNil)
}

func TestSubscribeDeviceUplink(t *testing.T) {
	a := New(t)
	c := NewClient(GetLogger(t, "Test"), "test", "", "", fmt.Sprintf("tcp://%s:1883", host))
	c.Connect()
	defer c.Disconnect()

	token := c.SubscribeDeviceUplink("someid", "someid", func(client Client, appID string, devID string, req UplinkMessage) {

	})
	token.Wait()
	a.So(token.Error(), ShouldBeNil)

	token = c.UnsubscribeDeviceUplink("someid", "someid")
	token.Wait()
	a.So(token.Error(), ShouldBeNil)
}

func TestSubscribeAppUplink(t *testing.T) {
	a := New(t)
	c := NewClient(GetLogger(t, "Test"), "test", "", "", fmt.Sprintf("tcp://%s:1883", host))
	c.Connect()
	defer c.Disconnect()

	token := c.SubscribeAppUplink("someid", func(client Client, appID string, devID string, req UplinkMessage) {

	})
	token.Wait()
	a.So(token.Error(), ShouldBeNil)

	token = c.UnsubscribeAppUplink("someid")
	token.Wait()
	a.So(token.Error(), ShouldBeNil)
}

func TestSubscribeUplink(t *testing.T) {
	a := New(t)
	c := NewClient(GetLogger(t, "Test"), "test", "", "", fmt.Sprintf("tcp://%s:1883", host))
	c.Connect()
	defer c.Disconnect()

	token := c.SubscribeUplink(func(client Client, appID string, devID string, req UplinkMessage) {

	})
	token.Wait()
	a.So(token.Error(), ShouldBeNil)

	token = c.UnsubscribeUplink()
	token.Wait()
	a.So(token.Error(), ShouldBeNil)
}

func TestPubSubUplink(t *testing.T) {
	a := New(t)
	c := NewClient(GetLogger(t, "Test"), "test", "", "", fmt.Sprintf("tcp://%s:1883", host))
	c.Connect()
	defer c.Disconnect()

	waitChan := make(chan bool, 1)

	c.SubscribeDeviceUplink("app1", "dev1", func(client Client, appID string, devID string, req UplinkMessage) {
		a.So(appID, ShouldResemble, "app1")
		a.So(devID, ShouldResemble, "dev1")

		waitChan <- true
	}).Wait()

	c.PublishUplink(UplinkMessage{
		Payload: []byte{0x01, 0x02, 0x03, 0x04},
		AppID:   "app1",
		DevID:   "dev1",
	}).Wait()

	select {
	case <-waitChan:
	case <-time.After(1 * time.Second):
		panic("Did not receive Uplink")
	}

	c.UnsubscribeDeviceUplink("app1", "dev1").Wait()
}

func TestPubSubAppUplink(t *testing.T) {
	a := New(t)
	c := NewClient(GetLogger(t, "Test"), "test", "", "", fmt.Sprintf("tcp://%s:1883", host))
	c.Connect()
	defer c.Disconnect()

	var wg WaitGroup

	wg.Add(2)

	c.SubscribeAppUplink("app2", func(client Client, appID string, devID string, req UplinkMessage) {
		a.So(appID, ShouldResemble, "app2")
		a.So(req.Payload, ShouldResemble, []byte{0x01, 0x02, 0x03, 0x04})
		wg.Done()
	}).Wait()

	c.PublishUplink(UplinkMessage{
		AppID:   "app2",
		DevID:   "dev1",
		Payload: []byte{0x01, 0x02, 0x03, 0x04},
	}).Wait()
	c.PublishUplink(UplinkMessage{
		AppID:   "app2",
		DevID:   "dev2",
		Payload: []byte{0x01, 0x02, 0x03, 0x04},
	}).Wait()

	a.So(wg.WaitFor(200*time.Millisecond), ShouldBeNil)

	c.UnsubscribeAppUplink("app1").Wait()
}

// Downlink pub/sub

func TestPublishDownlink(t *testing.T) {
	a := New(t)
	c := NewClient(GetLogger(t, "Test"), "test", "", "", fmt.Sprintf("tcp://%s:1883", host))
	c.Connect()
	defer c.Disconnect()

	dataDown := DownlinkMessage{
		AppID:   "someid",
		DevID:   "someid",
		Payload: []byte{0x01, 0x02, 0x03, 0x04},
	}

	token := c.PublishDownlink(dataDown)
	token.Wait()

	a.So(token.Error(), ShouldBeNil)
}

func TestSubscribeDeviceDownlink(t *testing.T) {
	a := New(t)
	c := NewClient(GetLogger(t, "Test"), "test", "", "", fmt.Sprintf("tcp://%s:1883", host))
	c.Connect()
	defer c.Disconnect()

	token := c.SubscribeDeviceDownlink("someid", "someid", func(client Client, appID string, devID string, req DownlinkMessage) {

	})
	token.Wait()
	a.So(token.Error(), ShouldBeNil)

	token = c.UnsubscribeDeviceDownlink("someid", "someid")
	token.Wait()
	a.So(token.Error(), ShouldBeNil)
}

func TestSubscribeAppDownlink(t *testing.T) {
	a := New(t)
	c := NewClient(GetLogger(t, "Test"), "test", "", "", fmt.Sprintf("tcp://%s:1883", host))
	c.Connect()
	defer c.Disconnect()

	token := c.SubscribeAppDownlink("someid", func(client Client, appID string, devID string, req DownlinkMessage) {

	})
	token.Wait()
	a.So(token.Error(), ShouldBeNil)

	token = c.UnsubscribeAppDownlink("someid")
	token.Wait()
	a.So(token.Error(), ShouldBeNil)
}

func TestSubscribeDownlink(t *testing.T) {
	a := New(t)
	c := NewClient(GetLogger(t, "Test"), "test", "", "", fmt.Sprintf("tcp://%s:1883", host))
	c.Connect()
	defer c.Disconnect()

	token := c.SubscribeDownlink(func(client Client, appID string, devID string, req DownlinkMessage) {

	})
	token.Wait()
	a.So(token.Error(), ShouldBeNil)

	token = c.UnsubscribeDownlink()
	token.Wait()
	a.So(token.Error(), ShouldBeNil)
}

func TestPubSubDownlink(t *testing.T) {
	a := New(t)
	c := NewClient(GetLogger(t, "Test"), "test", "", "", fmt.Sprintf("tcp://%s:1883", host))
	c.Connect()
	defer c.Disconnect()

	var wg WaitGroup

	wg.Add(1)

	c.SubscribeDeviceDownlink("app3", "dev3", func(client Client, appID string, devID string, req DownlinkMessage) {
		a.So(appID, ShouldResemble, "app3")
		a.So(devID, ShouldResemble, "dev3")

		wg.Done()
	}).Wait()

	c.PublishDownlink(DownlinkMessage{
		AppID:   "app3",
		DevID:   "dev3",
		Payload: []byte{0x01, 0x02, 0x03, 0x04},
	}).Wait()

	a.So(wg.WaitFor(200*time.Millisecond), ShouldBeNil)

	c.UnsubscribeDeviceDownlink("app3", "dev3").Wait()
}

func TestPubSubAppDownlink(t *testing.T) {
	a := New(t)
	c := NewClient(GetLogger(t, "Test"), "test", "", "", fmt.Sprintf("tcp://%s:1883", host))
	c.Connect()
	defer c.Disconnect()

	var wg WaitGroup

	wg.Add(2)

	c.SubscribeAppDownlink("app4", func(client Client, appID string, devID string, req DownlinkMessage) {
		a.So(appID, ShouldResemble, "app4")
		a.So(req.Payload, ShouldResemble, []byte{0x01, 0x02, 0x03, 0x04})
		wg.Done()
	}).Wait()

	c.PublishDownlink(DownlinkMessage{
		AppID:   "app4",
		DevID:   "dev1",
		Payload: []byte{0x01, 0x02, 0x03, 0x04},
	}).Wait()
	c.PublishDownlink(DownlinkMessage{
		AppID:   "app4",
		DevID:   "dev2",
		Payload: []byte{0x01, 0x02, 0x03, 0x04},
	}).Wait()

	a.So(wg.WaitFor(200*time.Millisecond), ShouldBeNil)

	c.UnsubscribeAppDownlink("app3").Wait()
}

// Activations pub/sub

func TestPublishActivations(t *testing.T) {
	a := New(t)
	c := NewClient(GetLogger(t, "Test"), "test", "", "", fmt.Sprintf("tcp://%s:1883", host))
	c.Connect()
	defer c.Disconnect()

	dataActivations := Activation{
		AppID:    "someid",
		DevID:    "someid",
		Metadata: Metadata{DataRate: "SF7BW125"},
	}

	token := c.PublishActivation(dataActivations)
	token.Wait()

	a.So(token.Error(), ShouldBeNil)
}

func TestSubscribeDeviceActivations(t *testing.T) {
	a := New(t)
	c := NewClient(GetLogger(t, "Test"), "test", "", "", fmt.Sprintf("tcp://%s:1883", host))
	c.Connect()
	defer c.Disconnect()

	token := c.SubscribeDeviceActivations("someid", "someid", func(client Client, appID string, devID string, req Activation) {

	})
	token.Wait()
	a.So(token.Error(), ShouldBeNil)

	token = c.UnsubscribeDeviceActivations("someid", "someid")
	token.Wait()
	a.So(token.Error(), ShouldBeNil)
}

func TestSubscribeAppActivations(t *testing.T) {
	a := New(t)
	c := NewClient(GetLogger(t, "Test"), "test", "", "", fmt.Sprintf("tcp://%s:1883", host))
	c.Connect()
	defer c.Disconnect()

	token := c.SubscribeAppActivations("someid", func(client Client, appID string, devID string, req Activation) {

	})
	token.Wait()
	a.So(token.Error(), ShouldBeNil)

	token = c.UnsubscribeAppActivations("someid")
	token.Wait()
	a.So(token.Error(), ShouldBeNil)
}

func TestSubscribeActivations(t *testing.T) {
	a := New(t)
	c := NewClient(GetLogger(t, "Test"), "test", "", "", fmt.Sprintf("tcp://%s:1883", host))
	c.Connect()
	defer c.Disconnect()

	token := c.SubscribeActivations(func(client Client, appID string, devID string, req Activation) {

	})
	token.Wait()
	a.So(token.Error(), ShouldBeNil)

	token = c.UnsubscribeActivations()
	token.Wait()
	a.So(token.Error(), ShouldBeNil)
}

func TestPubSubActivations(t *testing.T) {
	a := New(t)
	c := NewClient(GetLogger(t, "Test"), "test", "", "", fmt.Sprintf("tcp://%s:1883", host))
	c.Connect()
	defer c.Disconnect()

	var wg WaitGroup

	wg.Add(1)

	c.SubscribeDeviceActivations("app5", "dev1", func(client Client, appID string, devID string, req Activation) {
		a.So(appID, ShouldResemble, "app5")
		a.So(devID, ShouldResemble, "dev1")

		wg.Done()
	}).Wait()

	c.PublishActivation(Activation{
		AppID:    "app5",
		DevID:    "dev1",
		Metadata: Metadata{DataRate: "SF7BW125"},
	}).Wait()

	a.So(wg.WaitFor(200*time.Millisecond), ShouldBeNil)

	c.UnsubscribeDeviceActivations("app5", "dev1")
}

func TestPubSubAppActivations(t *testing.T) {
	a := New(t)
	c := NewClient(GetLogger(t, "Test"), "test", "", "", fmt.Sprintf("tcp://%s:1883", host))
	c.Connect()
	defer c.Disconnect()

	var wg WaitGroup

	wg.Add(2)

	c.SubscribeAppActivations("app6", func(client Client, appID string, devID string, req Activation) {
		a.So(appID, ShouldResemble, "app6")
		a.So(req.Metadata.DataRate, ShouldEqual, "SF7BW125")
		wg.Done()
	}).Wait()

	c.PublishActivation(Activation{
		AppID:    "app6",
		DevID:    "dev1",
		Metadata: Metadata{DataRate: "SF7BW125"},
	}).Wait()
	c.PublishActivation(Activation{
		AppID:    "app6",
		DevID:    "dev2",
		Metadata: Metadata{DataRate: "SF7BW125"},
	}).Wait()

	a.So(wg.WaitFor(200*time.Millisecond), ShouldBeNil)

	c.UnsubscribeAppActivations("app6")
}

func ExampleNewClient() {
	ctx := log.WithField("Example", "NewClient")
	exampleClient := NewClient(ctx, "ttnctl", "my-app-id", "my-access-key", "staging.thethingsnetwork.org:1883")
	err := exampleClient.Connect()
	if err != nil {
		ctx.WithError(err).Fatal("Could not connect")
	}
}

var exampleClient Client

func ExampleDefaultClient_SubscribeDeviceUplink() {
	token := exampleClient.SubscribeDeviceUplink("my-app-id", "my-dev-id", func(client Client, appID string, devID string, req UplinkMessage) {
		// Do something with the message
	})
	token.Wait()
	if err := token.Error(); err != nil {
		panic(err)
	}
}

func ExampleDefaultClient_PublishDownlink() {
	token := exampleClient.PublishDownlink(DownlinkMessage{
		AppID:   "my-app-id",
		DevID:   "my-dev-id",
		FPort:   1,
		Payload: []byte{0x01, 0x02, 0x03, 0x04},
	})
	token.Wait()
	if err := token.Error(); err != nil {
		panic(err)
	}
}

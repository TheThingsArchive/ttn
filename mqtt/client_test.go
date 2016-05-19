// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package mqtt

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/TheThingsNetwork/ttn/core"
	"github.com/TheThingsNetwork/ttn/core/types"
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

	eui := types.EUI64{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}
	dataUp := core.DataUpAppReq{
		Payload: []byte{0x01, 0x02, 0x03, 0x04},
	}

	token := c.PublishUplink(types.AppEUI(eui), types.DevEUI(eui), dataUp)
	token.Wait()

	a.So(token.Error(), ShouldBeNil)
}

func TestSubscribeDeviceUplink(t *testing.T) {
	a := New(t)
	c := NewClient(GetLogger(t, "Test"), "test", "", "", "tcp://localhost:1883")
	c.Connect()

	eui := types.EUI64{0x02, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}

	token := c.SubscribeDeviceUplink(types.AppEUI(eui), types.DevEUI(eui), func(client Client, appEUI types.AppEUI, devEUI types.DevEUI, req core.DataUpAppReq) {

	})
	token.Wait()

	a.So(token.Error(), ShouldBeNil)
}

func TestSubscribeAppUplink(t *testing.T) {
	a := New(t)
	c := NewClient(GetLogger(t, "Test"), "test", "", "", "tcp://localhost:1883")
	c.Connect()

	eui := types.AppEUI{0x03, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}

	token := c.SubscribeAppUplink(eui, func(client Client, appEUI types.AppEUI, devEUI types.DevEUI, req core.DataUpAppReq) {

	})
	token.Wait()

	a.So(token.Error(), ShouldBeNil)
}

func TestSubscribeUplink(t *testing.T) {
	a := New(t)
	c := NewClient(GetLogger(t, "Test"), "test", "", "", "tcp://localhost:1883")
	c.Connect()

	token := c.SubscribeUplink(func(client Client, appEUI types.AppEUI, devEUI types.DevEUI, req core.DataUpAppReq) {

	})
	token.Wait()

	a.So(token.Error(), ShouldBeNil)
}

func TestPubSubUplink(t *testing.T) {
	a := New(t)
	c := NewClient(GetLogger(t, "Test"), "test", "", "", "tcp://localhost:1883")
	c.Connect()

	appEUI := types.AppEUI{0x04, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}
	devEUI := types.DevEUI{0x08, 0x07, 0x06, 0x05, 0x04, 0x03, 0x02, 0x01}

	var wg sync.WaitGroup

	wg.Add(1)

	c.SubscribeDeviceUplink(appEUI, devEUI, func(client Client, appEUI types.AppEUI, devEUI types.DevEUI, req core.DataUpAppReq) {
		a.So(appEUI, ShouldResemble, appEUI)
		a.So(devEUI, ShouldResemble, devEUI)

		wg.Done()
	}).Wait()

	c.PublishUplink(appEUI, devEUI, core.DataUpAppReq{Payload: []byte{0x01, 0x02, 0x03, 0x04}}).Wait()

	wg.Wait()
}

func TestPubSubAppUplink(t *testing.T) {
	a := New(t)
	c := NewClient(GetLogger(t, "Test"), "test", "", "", "tcp://localhost:1883")
	c.Connect()

	appEUI := types.AppEUI{0x05, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}
	devEUI1 := types.DevEUI{0x08, 0x07, 0x06, 0x05, 0x04, 0x03, 0x02, 0x01}
	devEUI2 := types.DevEUI{0x08, 0x07, 0x06, 0x05, 0x04, 0x03, 0x02, 0x02}

	var wg sync.WaitGroup

	wg.Add(2)

	c.SubscribeAppUplink(appEUI, func(client Client, appEUI types.AppEUI, devEUI types.DevEUI, req core.DataUpAppReq) {
		a.So(appEUI, ShouldResemble, appEUI)
		a.So(req.Payload, ShouldResemble, []byte{0x01, 0x02, 0x03, 0x04})
		wg.Done()
	}).Wait()

	c.PublishUplink(appEUI, devEUI1, core.DataUpAppReq{Payload: []byte{0x01, 0x02, 0x03, 0x04}}).Wait()
	c.PublishUplink(appEUI, devEUI2, core.DataUpAppReq{Payload: []byte{0x01, 0x02, 0x03, 0x04}}).Wait()

	wg.Wait()
}

func TestInvalidUplink(t *testing.T) {
	ctx := GetLogger(t, "TestInvalidUplink")

	c := NewClient(ctx, "test", "", "", "tcp://localhost:1883")
	c.Connect()

	eui := types.AppEUI{0x06, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}

	c.SubscribeAppUplink(eui, func(client Client, appEUI types.AppEUI, devEUI types.DevEUI, req core.DataUpAppReq) {
		Ko(t, "Did not expect any message")
	}).Wait()

	// Invalid Topic
	c.(*defaultClient).mqtt.Publish("0602030405060708/devices/x/up", QoS, false, []byte{0x00}).Wait()

	// Invalid Payload
	c.(*defaultClient).mqtt.Publish("0602030405060708/devices/0602030405060708/up", QoS, false, []byte{0x00}).Wait()

	<-time.After(50 * time.Millisecond)

	ctx.Info("This test should have printed two warnings.")
}

// Downlink pub/sub

func TestPublishDownlink(t *testing.T) {
	a := New(t)
	c := NewClient(GetLogger(t, "Test"), "test", "", "", "tcp://localhost:1883")
	c.Connect()

	eui := types.EUI64{0x01, 0x03, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}
	dataDown := core.DataDownAppReq{
		Payload: []byte{0x01, 0x02, 0x03, 0x04},
	}

	token := c.PublishDownlink(types.AppEUI(eui), types.DevEUI(eui), dataDown)
	token.Wait()

	a.So(token.Error(), ShouldBeNil)
}

func TestSubscribeDeviceDownlink(t *testing.T) {
	a := New(t)
	c := NewClient(GetLogger(t, "Test"), "test", "", "", "tcp://localhost:1883")
	c.Connect()

	eui := types.EUI64{0x02, 0x03, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}

	token := c.SubscribeDeviceDownlink(types.AppEUI(eui), types.DevEUI(eui), func(client Client, appEUI types.AppEUI, devEUI types.DevEUI, req core.DataDownAppReq) {

	})
	token.Wait()

	a.So(token.Error(), ShouldBeNil)
}

func TestSubscribeAppDownlink(t *testing.T) {
	a := New(t)
	c := NewClient(GetLogger(t, "Test"), "test", "", "", "tcp://localhost:1883")
	c.Connect()

	eui := types.AppEUI{0x03, 0x03, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}

	token := c.SubscribeAppDownlink(eui, func(client Client, appEUI types.AppEUI, devEUI types.DevEUI, req core.DataDownAppReq) {

	})
	token.Wait()

	a.So(token.Error(), ShouldBeNil)
}

func TestSubscribeDownlink(t *testing.T) {
	a := New(t)
	c := NewClient(GetLogger(t, "Test"), "test", "", "", "tcp://localhost:1883")
	c.Connect()

	token := c.SubscribeDownlink(func(client Client, appEUI types.AppEUI, devEUI types.DevEUI, req core.DataDownAppReq) {

	})
	token.Wait()

	a.So(token.Error(), ShouldBeNil)
}

func TestPubSubDownlink(t *testing.T) {
	a := New(t)
	c := NewClient(GetLogger(t, "Test"), "test", "", "", "tcp://localhost:1883")
	c.Connect()

	appEUI := types.AppEUI{0x04, 0x03, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}
	devEUI := types.DevEUI{0x08, 0x07, 0x06, 0x05, 0x04, 0x03, 0x02, 0x01}

	var wg sync.WaitGroup

	wg.Add(1)

	c.SubscribeDeviceDownlink(appEUI, devEUI, func(client Client, appEUI types.AppEUI, devEUI types.DevEUI, req core.DataDownAppReq) {
		a.So(appEUI, ShouldResemble, appEUI)
		a.So(devEUI, ShouldResemble, devEUI)

		wg.Done()
	}).Wait()

	c.PublishDownlink(appEUI, devEUI, core.DataDownAppReq{Payload: []byte{0x01, 0x02, 0x03, 0x04}}).Wait()

	wg.Wait()
}

func TestPubSubAppDownlink(t *testing.T) {
	a := New(t)
	c := NewClient(GetLogger(t, "Test"), "test", "", "", "tcp://localhost:1883")
	c.Connect()

	appEUI := types.AppEUI{0x05, 0x03, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}
	devEUI1 := types.DevEUI{0x08, 0x07, 0x06, 0x05, 0x04, 0x03, 0x02, 0x01}
	devEUI2 := types.DevEUI{0x08, 0x07, 0x06, 0x05, 0x04, 0x03, 0x02, 0x02}

	var wg sync.WaitGroup

	wg.Add(2)

	c.SubscribeAppDownlink(appEUI, func(client Client, appEUI types.AppEUI, devEUI types.DevEUI, req core.DataDownAppReq) {
		a.So(appEUI, ShouldResemble, appEUI)
		a.So(req.Payload, ShouldResemble, []byte{0x01, 0x02, 0x03, 0x04})
		wg.Done()
	}).Wait()

	c.PublishDownlink(appEUI, devEUI1, core.DataDownAppReq{Payload: []byte{0x01, 0x02, 0x03, 0x04}}).Wait()
	c.PublishDownlink(appEUI, devEUI2, core.DataDownAppReq{Payload: []byte{0x01, 0x02, 0x03, 0x04}}).Wait()

	wg.Wait()
}

func TestInvalidDownlink(t *testing.T) {
	ctx := GetLogger(t, "TestInvalidDownlink")

	c := NewClient(ctx, "test", "", "", "tcp://localhost:1883")
	c.Connect()

	eui := types.AppEUI{0x06, 0x03, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}

	c.SubscribeAppDownlink(eui, func(client Client, appEUI types.AppEUI, devEUI types.DevEUI, req core.DataDownAppReq) {
		Ko(t, "Did not expect any message")
	}).Wait()

	// Invalid Topic
	c.(*defaultClient).mqtt.Publish("0603030405060708/devices/x/down", QoS, false, []byte{0x00}).Wait()

	// Invalid Payload
	c.(*defaultClient).mqtt.Publish("0603030405060708/devices/0602030405060708/down", QoS, false, []byte{0x00}).Wait()

	<-time.After(50 * time.Millisecond)

	ctx.Info("This test should have printed two warnings.")
}

// Activations pub/sub

func TestPublishActivations(t *testing.T) {
	a := New(t)
	c := NewClient(GetLogger(t, "Test"), "test", "", "", "tcp://localhost:1883")
	c.Connect()

	eui := types.EUI64{0x01, 0x04, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}
	dataActivations := core.OTAAAppReq{Metadata: []core.AppMetadata{core.AppMetadata{DataRate: "SF7BW125"}}}

	token := c.PublishActivation(types.AppEUI(eui), types.DevEUI(eui), dataActivations)
	token.Wait()

	a.So(token.Error(), ShouldBeNil)
}

func TestSubscribeDeviceActivations(t *testing.T) {
	a := New(t)
	c := NewClient(GetLogger(t, "Test"), "test", "", "", "tcp://localhost:1883")
	c.Connect()

	eui := types.EUI64{0x02, 0x04, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}

	token := c.SubscribeDeviceActivations(types.AppEUI(eui), types.DevEUI(eui), func(client Client, appEUI types.AppEUI, devEUI types.DevEUI, req core.OTAAAppReq) {

	})
	token.Wait()

	a.So(token.Error(), ShouldBeNil)
}

func TestSubscribeAppActivations(t *testing.T) {
	a := New(t)
	c := NewClient(GetLogger(t, "Test"), "test", "", "", "tcp://localhost:1883")
	c.Connect()

	eui := types.AppEUI{0x03, 0x04, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}

	token := c.SubscribeAppActivations(eui, func(client Client, appEUI types.AppEUI, devEUI types.DevEUI, req core.OTAAAppReq) {

	})
	token.Wait()

	a.So(token.Error(), ShouldBeNil)
}

func TestSubscribeActivations(t *testing.T) {
	a := New(t)
	c := NewClient(GetLogger(t, "Test"), "test", "", "", "tcp://localhost:1883")
	c.Connect()

	token := c.SubscribeActivations(func(client Client, appEUI types.AppEUI, devEUI types.DevEUI, req core.OTAAAppReq) {

	})
	token.Wait()

	a.So(token.Error(), ShouldBeNil)
}

func TestPubSubActivations(t *testing.T) {
	a := New(t)
	c := NewClient(GetLogger(t, "Test"), "test", "", "", "tcp://localhost:1883")
	c.Connect()

	appEUI := types.AppEUI{0x04, 0x04, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}
	devEUI := types.DevEUI{0x08, 0x07, 0x06, 0x05, 0x04, 0x03, 0x02, 0x01}

	var wg sync.WaitGroup

	wg.Add(1)

	c.SubscribeDeviceActivations(appEUI, devEUI, func(client Client, appEUI types.AppEUI, devEUI types.DevEUI, req core.OTAAAppReq) {
		a.So(appEUI, ShouldResemble, appEUI)
		a.So(devEUI, ShouldResemble, devEUI)

		wg.Done()
	}).Wait()

	c.PublishActivation(appEUI, devEUI, core.OTAAAppReq{Metadata: []core.AppMetadata{core.AppMetadata{DataRate: "SF7BW125"}}}).Wait()

	wg.Wait()
}

func TestPubSubAppActivations(t *testing.T) {
	a := New(t)
	c := NewClient(GetLogger(t, "Test"), "test", "", "", "tcp://localhost:1883")
	c.Connect()

	appEUI := types.AppEUI{0x05, 0x04, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}
	devEUI1 := types.DevEUI{0x08, 0x07, 0x06, 0x05, 0x04, 0x03, 0x02, 0x01}
	devEUI2 := types.DevEUI{0x08, 0x07, 0x06, 0x05, 0x04, 0x03, 0x02, 0x02}

	var wg sync.WaitGroup

	wg.Add(2)

	c.SubscribeAppActivations(appEUI, func(client Client, appEUI types.AppEUI, devEUI types.DevEUI, req core.OTAAAppReq) {
		a.So(appEUI, ShouldResemble, appEUI)
		a.So(req.Metadata[0].DataRate, ShouldEqual, "SF7BW125")
		wg.Done()
	}).Wait()

	c.PublishActivation(appEUI, devEUI1, core.OTAAAppReq{Metadata: []core.AppMetadata{core.AppMetadata{DataRate: "SF7BW125"}}}).Wait()
	c.PublishActivation(appEUI, devEUI2, core.OTAAAppReq{Metadata: []core.AppMetadata{core.AppMetadata{DataRate: "SF7BW125"}}}).Wait()

	wg.Wait()
}

func TestInvalidActivations(t *testing.T) {
	ctx := GetLogger(t, "TestInvalidActivations")

	c := NewClient(ctx, "test", "", "", "tcp://localhost:1883")
	c.Connect()

	eui := types.AppEUI{0x06, 0x04, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}

	c.SubscribeAppActivations(eui, func(client Client, appEUI types.AppEUI, devEUI types.DevEUI, req core.OTAAAppReq) {
		Ko(t, "Did not expect any message")
	}).Wait()

	// Invalid Topic
	c.(*defaultClient).mqtt.Publish("0604030405060708/devices/x/activations", QoS, false, []byte{0x00}).Wait()

	// Invalid Payload
	c.(*defaultClient).mqtt.Publish("0604030405060708/devices/0602030405060708/activations", QoS, false, []byte{0x00}).Wait()

	<-time.After(50 * time.Millisecond)

	ctx.Info("This test should have printed two warnings.")
}

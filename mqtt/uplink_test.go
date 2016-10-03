// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package mqtt

import (
	"fmt"
	"strings"
	"testing"
	"time"

	. "github.com/TheThingsNetwork/ttn/utils/testing"
	MQTT "github.com/eclipse/paho.mqtt.golang"
	. "github.com/smartystreets/assertions"
)

// Uplink pub/sub

func TestPublishUplink(t *testing.T) {
	a := New(t)
	c := NewClient(GetLogger(t, "Test"), "test", "", "", fmt.Sprintf("tcp://%s:1883", host))
	c.Connect()
	defer c.Disconnect()

	dataUp := UplinkMessage{
		AppID:      "someid",
		DevID:      "someid",
		PayloadRaw: []byte{0x01, 0x02, 0x03, 0x04},
	}

	token := c.PublishUplink(dataUp)
	waitForOK(token, a)
}

func TestPublishUplinkFields(t *testing.T) {
	a := New(t)
	ctx := GetLogger(t, "Test")
	c := NewClient(ctx, "test", "", "", fmt.Sprintf("tcp://%s:1883", host))

	c.Connect()
	defer c.Disconnect()

	subChan := make(chan bool)
	waitChan := make(chan bool)
	go func() {
		for i := 10; i > 0; i-- {
			<-subChan
		}
		close(subChan)
		waitChan <- true
	}()
	subToken := c.(*DefaultClient).mqtt.Subscribe("fields-app/devices/fields-dev/up/#", SubscribeQoS, func(_ MQTT.Client, msg MQTT.Message) {
		switch strings.TrimPrefix(msg.Topic(), "fields-app/devices/fields-dev/up/") {
		case "battery":
			a.So(string(msg.Payload()), ShouldEqual, "90")
		case "sensors":
			a.So(string(msg.Payload()), ShouldContainSubstring, `people":["`)
		case "sensors/color":
			a.So(string(msg.Payload()), ShouldEqual, `"blue"`)
		case "sensors/people":
			a.So(string(msg.Payload()), ShouldEqual, `["bob","alice"]`)
		case "sensors/water":
			a.So(string(msg.Payload()), ShouldEqual, "true")
		case "sensors/analog":
			a.So(string(msg.Payload()), ShouldEqual, `[0,255,500,1000]`)
		case "sensors/history":
			a.So(string(msg.Payload()), ShouldContainSubstring, `today":"`)
		case "sensors/history/today":
			a.So(string(msg.Payload()), ShouldEqual, `"not yet"`)
		case "sensors/history/yesterday":
			a.So(string(msg.Payload()), ShouldEqual, `"absolutely"`)
		case "gps":
			a.So(string(msg.Payload()), ShouldEqual, "[52.3736735,4.886663]")
		default:
			t.Errorf("Should not have received message on topic %s", msg.Topic())
			t.Fail()
		}
		subChan <- true
	})
	waitForOK(subToken, a)

	fields := map[string]interface{}{
		"battery": 90,
		"sensors": map[string]interface{}{
			"color":  "blue",
			"people": []string{"bob", "alice"},
			"water":  true,
			"analog": []int{0, 255, 500, 1000},
			"history": map[string]interface{}{
				"today":     "not yet",
				"yesterday": "absolutely",
			},
		},
		"gps": []float64{52.3736735, 4.886663},
	}

	pubToken := c.PublishUplinkFields("fields-app", "fields-dev", fields)
	waitForOK(pubToken, a)

	select {
	case <-waitChan:
	case <-time.After(1 * time.Second):
		panic("Did not receive fields")
	}
}

func TestSubscribeDeviceUplink(t *testing.T) {
	a := New(t)
	c := NewClient(GetLogger(t, "Test"), "test", "", "", fmt.Sprintf("tcp://%s:1883", host))
	c.Connect()
	defer c.Disconnect()

	subToken := c.SubscribeDeviceUplink("someid", "someid", func(client Client, appID string, devID string, req UplinkMessage) {

	})
	waitForOK(subToken, a)

	unsubToken := c.UnsubscribeDeviceUplink("someid", "someid")
	waitForOK(unsubToken, a)
}

func TestSubscribeAppUplink(t *testing.T) {
	a := New(t)
	c := NewClient(GetLogger(t, "Test"), "test", "", "", fmt.Sprintf("tcp://%s:1883", host))
	c.Connect()
	defer c.Disconnect()

	subToken := c.SubscribeAppUplink("someid", func(client Client, appID string, devID string, req UplinkMessage) {

	})
	waitForOK(subToken, a)

	unsubToken := c.UnsubscribeAppUplink("someid")
	waitForOK(unsubToken, a)
}

func TestSubscribeUplink(t *testing.T) {
	a := New(t)
	c := NewClient(GetLogger(t, "Test"), "test", "", "", fmt.Sprintf("tcp://%s:1883", host))
	c.Connect()
	defer c.Disconnect()

	subToken := c.SubscribeUplink(func(client Client, appID string, devID string, req UplinkMessage) {

	})
	waitForOK(subToken, a)

	unsubToken := c.UnsubscribeUplink()
	waitForOK(unsubToken, a)
}

func TestPubSubUplink(t *testing.T) {
	a := New(t)
	c := NewClient(GetLogger(t, "Test"), "test", "", "", fmt.Sprintf("tcp://%s:1883", host))
	c.Connect()
	defer c.Disconnect()

	waitChan := make(chan bool, 1)

	subToken := c.SubscribeDeviceUplink("app1", "dev1", func(client Client, appID string, devID string, req UplinkMessage) {
		a.So(appID, ShouldResemble, "app1")
		a.So(devID, ShouldResemble, "dev1")

		waitChan <- true
	})
	waitForOK(subToken, a)

	pubToken := c.PublishUplink(UplinkMessage{
		PayloadRaw: []byte{0x01, 0x02, 0x03, 0x04},
		AppID:      "app1",
		DevID:      "dev1",
	})
	waitForOK(pubToken, a)

	select {
	case <-waitChan:
	case <-time.After(1 * time.Second):
		panic("Did not receive Uplink")
	}

	unsubToken := c.UnsubscribeDeviceUplink("app1", "dev1")
	waitForOK(unsubToken, a)
}

func TestPubSubAppUplink(t *testing.T) {
	a := New(t)
	c := NewClient(GetLogger(t, "Test"), "test", "", "", fmt.Sprintf("tcp://%s:1883", host))
	c.Connect()
	defer c.Disconnect()

	var wg WaitGroup

	wg.Add(2)

	subToken := c.SubscribeAppUplink("app2", func(client Client, appID string, devID string, req UplinkMessage) {
		a.So(appID, ShouldResemble, "app2")
		a.So(req.PayloadRaw, ShouldResemble, []byte{0x01, 0x02, 0x03, 0x04})
		wg.Done()
	})
	waitForOK(subToken, a)

	pubToken := c.PublishUplink(UplinkMessage{
		AppID:      "app2",
		DevID:      "dev1",
		PayloadRaw: []byte{0x01, 0x02, 0x03, 0x04},
	})
	waitForOK(pubToken, a)
	pubToken = c.PublishUplink(UplinkMessage{
		AppID:      "app2",
		DevID:      "dev2",
		PayloadRaw: []byte{0x01, 0x02, 0x03, 0x04},
	})
	waitForOK(pubToken, a)

	a.So(wg.WaitFor(200*time.Millisecond), ShouldBeNil)

	unsubToken := c.UnsubscribeAppUplink("app1")
	waitForOK(unsubToken, a)
}

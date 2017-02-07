// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package handler

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/TheThingsNetwork/ttn/core/component"
	"github.com/TheThingsNetwork/ttn/core/handler/device"
	"github.com/TheThingsNetwork/ttn/core/types"
	"github.com/TheThingsNetwork/ttn/mqtt"
	. "github.com/TheThingsNetwork/ttn/utils/testing"
	. "github.com/smartystreets/assertions"
)

func TestHandleMQTT(t *testing.T) {
	host := os.Getenv("MQTT_ADDRESS")
	if host == "" {
		host = "localhost:1883"
	}

	a := New(t)
	var wg WaitGroup
	c := mqtt.NewClient(GetLogger(t, "TestHandleMQTT"), "test", "", "", fmt.Sprintf("tcp://%s", host))
	err := c.Connect()
	a.So(err, ShouldBeNil)
	appID := "handler-mqtt-app1"
	devID := "handler-mqtt-dev1"
	h := &handler{
		Component: &component.Component{Ctx: GetLogger(t, "TestHandleMQTT")},
		devices:   device.NewRedisDeviceStore(GetRedisClient(), "handler-test-handle-mqtt"),
	}
	h.devices.Set(&device.Device{
		AppID: appID,
		DevID: devID,
	})
	defer func() {
		h.devices.Delete(appID, devID)
	}()
	err = h.HandleMQTT("", "", fmt.Sprintf("tcp://%s", host))
	a.So(err, ShouldBeNil)

	c.PublishDownlink(types.DownlinkMessage{
		AppID:      appID,
		DevID:      devID,
		PayloadRaw: []byte{0xAA, 0xBC},
	}).Wait()
	<-time.After(50 * time.Millisecond)
	q, _ := h.devices.DownlinkQueue(appID, devID)
	downlink, _ := q.Next()
	a.So(downlink, ShouldNotBeNil)

	wg.Add(1)
	c.SubscribeDeviceUplink(appID, devID, func(client mqtt.Client, r_appID string, r_devID string, req types.UplinkMessage) {
		a.So(r_appID, ShouldEqual, appID)
		a.So(r_devID, ShouldEqual, devID)
		a.So(req.PayloadRaw, ShouldResemble, []byte{0xAA, 0xBC})
		wg.Done()
	}).Wait()

	h.mqttUp <- &types.UplinkMessage{
		DevID:      devID,
		AppID:      appID,
		PayloadRaw: []byte{0xAA, 0xBC},
		PayloadFields: map[string]interface{}{
			"field": "value",
		},
	}

	wg.Add(1)
	c.SubscribeDeviceActivations(appID, devID, func(client mqtt.Client, r_appID string, r_devID string, req types.Activation) {
		a.So(r_appID, ShouldEqual, appID)
		a.So(r_devID, ShouldEqual, devID)
		wg.Done()
	}).Wait()

	h.mqttEvent <- &types.DeviceEvent{
		DevID: devID,
		AppID: appID,
		Event: types.ActivationEvent,
	}

	a.So(wg.WaitFor(200*time.Millisecond), ShouldBeNil)
}

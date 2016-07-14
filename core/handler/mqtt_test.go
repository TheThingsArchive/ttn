// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package handler

import (
	"sync"
	"testing"
	"time"

	"github.com/TheThingsNetwork/ttn/core"
	"github.com/TheThingsNetwork/ttn/core/handler/device"
	"github.com/TheThingsNetwork/ttn/mqtt"
	. "github.com/TheThingsNetwork/ttn/utils/testing"
	. "github.com/smartystreets/assertions"
)

func TestHandleMQTT(t *testing.T) {
	a := New(t)
	var wg sync.WaitGroup
	c := mqtt.NewClient(GetLogger(t, "TestHandleMQTT"), "test", "", "", "tcp://localhost:1883")
	c.Connect()
	appID := "app1"
	devID := "dev1"
	h := &handler{
		Component: &core.Component{Ctx: GetLogger(t, "TestHandleMQTT")},
		devices:   device.NewDeviceStore(),
	}
	h.devices.Set(&device.Device{
		AppID: appID,
		DevID: devID,
	})
	err := h.HandleMQTT("", "", "tcp://localhost:1883")
	a.So(err, ShouldBeNil)

	c.PublishDownlink(appID, devID, mqtt.DownlinkMessage{
		Payload: []byte{0xAA, 0xBC},
	}).Wait()
	<-time.After(50 * time.Millisecond)
	dev, _ := h.devices.Get(appID, devID)
	a.So(dev.NextDownlink, ShouldNotBeNil)

	wg.Add(1)
	c.SubscribeUplink(func(client mqtt.Client, r_appID string, r_devID string, req mqtt.UplinkMessage) {
		a.So(r_appID, ShouldEqual, appID)
		a.So(r_devID, ShouldEqual, devID)
		a.So(req.Payload, ShouldResemble, []byte{0xAA, 0xBC})
		wg.Done()
	})

	h.mqttUp <- &mqtt.UplinkMessage{
		DevID:   devID,
		AppID:   appID,
		Payload: []byte{0xAA, 0xBC},
	}

	wg.Add(1)
	c.SubscribeActivations(func(client mqtt.Client, r_appID string, r_devID string, req mqtt.Activation) {
		a.So(r_appID, ShouldEqual, appID)
		a.So(r_devID, ShouldEqual, devID)
		wg.Done()
	})

	h.mqttActivation <- &mqtt.Activation{
		DevID: devID,
		AppID: appID,
	}

	wg.Wait()
}

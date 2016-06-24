// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package handler

import (
	"sync"
	"testing"
	"time"

	"github.com/TheThingsNetwork/ttn/core"
	"github.com/TheThingsNetwork/ttn/core/handler/device"
	"github.com/TheThingsNetwork/ttn/core/types"
	"github.com/TheThingsNetwork/ttn/mqtt"
	. "github.com/TheThingsNetwork/ttn/utils/testing"
	. "github.com/smartystreets/assertions"
)

func TestHandleMQTT(t *testing.T) {
	a := New(t)
	var wg sync.WaitGroup
	c := mqtt.NewClient(GetLogger(t, "TestHandleMQTT"), "test", "", "", "tcp://localhost:1883")
	c.Connect()
	appEUI := types.AppEUI{1, 2, 3, 4, 5, 6, 7, 8}
	devEUI := types.DevEUI{1, 2, 3, 4, 5, 6, 7, 8}
	h := &handler{
		Component: &core.Component{Ctx: GetLogger(t, "TestHandleMQTT")},
		devices:   device.NewDeviceStore(),
	}
	h.devices.Set(&device.Device{
		AppEUI: appEUI,
		DevEUI: devEUI,
	})
	err := h.HandleMQTT("", "", "tcp://localhost:1883")
	a.So(err, ShouldBeNil)

	c.PublishDownlink(appEUI, devEUI, mqtt.DownlinkMessage{
		Payload: []byte{0xAA, 0xBC},
	}).Wait()
	<-time.After(50 * time.Millisecond)
	dev, _ := h.devices.Get(appEUI, devEUI)
	a.So(dev.NextDownlink, ShouldNotBeNil)

	wg.Add(1)
	c.SubscribeUplink(func(client mqtt.Client, r_appEUI types.AppEUI, r_devEUI types.DevEUI, req mqtt.UplinkMessage) {
		a.So(r_appEUI, ShouldEqual, appEUI)
		a.So(r_devEUI, ShouldEqual, devEUI)
		a.So(req.Payload, ShouldResemble, []byte{0xAA, 0xBC})
		wg.Done()
	})

	h.mqttUp <- &mqtt.UplinkMessage{
		DevEUI:  devEUI,
		AppEUI:  appEUI,
		Payload: []byte{0xAA, 0xBC},
	}

	wg.Add(1)
	c.SubscribeActivations(func(client mqtt.Client, r_appEUI types.AppEUI, r_devEUI types.DevEUI, req mqtt.Activation) {
		a.So(r_appEUI, ShouldEqual, appEUI)
		a.So(r_devEUI, ShouldEqual, devEUI)
		wg.Done()
	})

	h.mqttActivation <- &mqtt.Activation{
		DevEUI: devEUI,
		AppEUI: appEUI,
	}

	wg.Wait()
}

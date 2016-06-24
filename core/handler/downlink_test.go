// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package handler

import (
	"testing"

	pb_broker "github.com/TheThingsNetwork/ttn/api/broker"
	"github.com/TheThingsNetwork/ttn/core"
	"github.com/TheThingsNetwork/ttn/core/handler/device"
	"github.com/TheThingsNetwork/ttn/core/types"
	"github.com/TheThingsNetwork/ttn/mqtt"
	. "github.com/TheThingsNetwork/ttn/utils/testing"
	. "github.com/smartystreets/assertions"
)

func TestEnqueueDownlink(t *testing.T) {
	a := New(t)
	appEUI := types.AppEUI{1, 2, 3, 4, 5, 6, 7, 8}
	devEUI := types.DevEUI{1, 2, 3, 4, 5, 6, 7, 8}
	h := &handler{
		Component: &core.Component{Ctx: GetLogger(t, "TestEnqueueDownlink")},
		devices:   device.NewDeviceStore(),
	}
	err := h.EnqueueDownlink(&mqtt.DownlinkMessage{
		AppEUI: appEUI,
		DevEUI: devEUI,
	})
	a.So(err, ShouldNotBeNil)
	h.devices.Set(&device.Device{
		AppEUI: appEUI,
		DevEUI: devEUI,
	})
	err = h.EnqueueDownlink(&mqtt.DownlinkMessage{
		AppEUI: appEUI,
		DevEUI: devEUI,
		Fields: map[string]interface{}{
			"string": "hello!",
			"int":    42,
			"bool":   true,
		},
	})
	a.So(err, ShouldBeNil)
	dev, _ := h.devices.Get(appEUI, devEUI)
	a.So(dev.NextDownlink, ShouldNotBeEmpty)
	a.So(dev.NextDownlink.Fields, ShouldHaveLength, 3)
}

func TestHandleDownlink(t *testing.T) {
	a := New(t)
	appEUI := types.AppEUI{1, 2, 3, 4, 5, 6, 7, 8}
	devEUI := types.DevEUI{1, 2, 3, 4, 5, 6, 7, 8}
	h := &handler{
		Component: &core.Component{Ctx: GetLogger(t, "TestHandleDownlink")},
		devices:   device.NewDeviceStore(),
	}
	err := h.HandleDownlink(&mqtt.DownlinkMessage{
		AppEUI: appEUI,
		DevEUI: devEUI,
	}, &pb_broker.DownlinkMessage{
		AppEui: &appEUI,
		DevEui: &devEUI,
	})
	a.So(err, ShouldNotBeNil)
	h.devices.Set(&device.Device{
		AppEUI: appEUI,
		DevEUI: devEUI,
	})
	err = h.HandleDownlink(&mqtt.DownlinkMessage{
		AppEUI: appEUI,
		DevEUI: devEUI,
	}, &pb_broker.DownlinkMessage{
		AppEui:  &appEUI,
		DevEui:  &devEUI,
		Payload: []byte{96, 4, 3, 2, 1, 0, 1, 0, 1, 0, 0, 0, 0},
	})
	a.So(err, ShouldBeNil)
	h.downlink = make(chan *pb_broker.DownlinkMessage)
	go func() {
		dl := <-h.downlink
		a.So(dl.Payload, ShouldNotBeEmpty)
	}()
	err = h.HandleDownlink(&mqtt.DownlinkMessage{
		AppEUI:  appEUI,
		DevEUI:  devEUI,
		Payload: []byte{0xAA, 0xBC},
	}, &pb_broker.DownlinkMessage{
		AppEui:  &appEUI,
		DevEui:  &devEUI,
		Payload: []byte{96, 4, 3, 2, 1, 0, 1, 0, 1, 0, 0, 0, 0},
	})
	a.So(err, ShouldBeNil)
}

// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package handler

import (
	"errors"
	"testing"
	"time"

	pb_broker "github.com/TheThingsNetwork/ttn/api/broker"
	"github.com/TheThingsNetwork/ttn/core"
	"github.com/TheThingsNetwork/ttn/core/handler/application"
	"github.com/TheThingsNetwork/ttn/core/handler/device"
	"github.com/TheThingsNetwork/ttn/core/types"
	"github.com/TheThingsNetwork/ttn/mqtt"
	. "github.com/TheThingsNetwork/ttn/utils/testing"
	. "github.com/smartystreets/assertions"
)

func TestEnqueueDownlink(t *testing.T) {
	a := New(t)
	appID := "app1"
	devID := "dev1"
	h := &handler{
		Component: &core.Component{Ctx: GetLogger(t, "TestEnqueueDownlink")},
		devices:   device.NewDeviceStore(),
	}
	err := h.EnqueueDownlink(&mqtt.DownlinkMessage{
		AppID: appID,
		DevID: devID,
	})
	a.So(err, ShouldNotBeNil)
	h.devices.Set(&device.Device{
		AppID: appID,
		DevID: devID,
	})
	err = h.EnqueueDownlink(&mqtt.DownlinkMessage{
		AppID: appID,
		DevID: devID,
		Fields: map[string]interface{}{
			"string": "hello!",
			"int":    42,
			"bool":   true,
		},
	})
	a.So(err, ShouldBeNil)
	dev, _ := h.devices.Get(appID, devID)
	a.So(dev.NextDownlink, ShouldNotBeEmpty)
	a.So(dev.NextDownlink.Fields, ShouldHaveLength, 3)
}

func TestHandleDownlink(t *testing.T) {
	a := New(t)
	appID := "app2"
	devID := "dev2"
	appEUI := types.AppEUI([8]byte{1, 2, 3, 4, 5, 6, 7, 8})
	devEUI := types.DevEUI([8]byte{1, 2, 3, 4, 5, 6, 7, 8})
	h := &handler{
		Component:    &core.Component{Ctx: GetLogger(t, "TestHandleDownlink")},
		devices:      device.NewDeviceStore(),
		applications: application.NewApplicationStore(),
	}
	// Neither payload nor Fields provided : ERROR
	err := h.HandleDownlink(&mqtt.DownlinkMessage{
		AppID: appID,
		DevID: devID,
	}, &pb_broker.DownlinkMessage{
		AppEui: &appEUI,
		DevEui: &devEUI,
	})
	a.So(err, ShouldNotBeNil)

	h.devices.Set(&device.Device{
		AppID: appID,
		DevID: devID,
	})
	err = h.HandleDownlink(&mqtt.DownlinkMessage{
		AppID: appID,
		DevID: devID,
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

	select {
	case <-h.downlink:
		t.Log("No timeout")
	case <-time.After(time.Millisecond * 50):
		err = errors.New("Timeout: can't write in the channel")
	}

	// Payload provided
	err = h.HandleDownlink(&mqtt.DownlinkMessage{
		AppID:   appID,
		DevID:   devID,
		Payload: []byte{0xAA, 0xBC},
	}, &pb_broker.DownlinkMessage{
		AppEui:  &appEUI,
		DevEui:  &devEUI,
		Payload: []byte{96, 4, 3, 2, 1, 0, 1, 0, 1, 0, 0, 0, 0},
	})
	a.So(err, ShouldBeNil)

	h.applications.Set(&application.Application{
		AppID: appID,
		Encoder: `function (payload){
	  		return [96, 4, 3, 2, 1, 0, 1, 0, 1, 0, 0, 0, 0]
			}`,
	})
	jsonFields := make(map[string]interface{})
	jsonFields["key"] = "value"
	jsonFields["key2"] = 11
	// Both Payload and Fields provided : ERROR
	err = h.HandleDownlink(&mqtt.DownlinkMessage{
		FPort:   1,
		AppID:   appID,
		DevID:   devID,
		Fields:  jsonFields,
		Payload: []byte{0xAA, 0xBC},
	}, &pb_broker.DownlinkMessage{
		AppEui: &appEUI,
		DevEui: &devEUI,
	})
	a.So(err, ShouldNotBeNil)

	// Wait for the first goroutine to end
	go func() {
		time.Sleep(time.Millisecond * 50)
		dl := <-h.downlink
		a.So(dl.Payload, ShouldNotBeEmpty)
	}()

	select {
	case <-h.downlink:
		t.Log("No timeout")
	case <-time.After(time.Millisecond * 50):
		err = errors.New("Timeout: can't write in the channel")
	}

	// Json Fields provided
	err = h.HandleDownlink(&mqtt.DownlinkMessage{
		FPort:  1,
		AppID:  appID,
		DevID:  devID,
		Fields: jsonFields,
	}, &pb_broker.DownlinkMessage{
		AppEui: &appEUI,
		DevEui: &devEUI,
	})
	a.So(err, ShouldBeNil)
}

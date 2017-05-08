// Copyright © 2017 The Things Industries B.V.

package handler

import (
	"context"
	"testing"

	pb_lorawan "github.com/TheThingsNetwork/ttn/api/protocol/lorawan"
	"github.com/TheThingsNetwork/ttn/core/component"
	"github.com/TheThingsNetwork/ttn/core/handler/application"
	"github.com/TheThingsNetwork/ttn/core/handler/device"
	"github.com/TheThingsNetwork/ttn/core/types"
	. "github.com/TheThingsNetwork/ttn/utils/testing"
	. "github.com/smartystreets/assertions"
)

func TestHandlerManager_EventSelectCreate(t *testing.T) {
	a := New(t)

	h := &handler{
		Component:    &component.Component{Ctx: GetLogger(t, "TestEventSelect")},
		applications: application.NewRedisApplicationStore(GetRedisClient(), "handler-test-event-select"),
		devices:      device.NewRedisDeviceStore(GetRedisClient(), "handler-test-event-select"),
		qEvent:       make(chan *types.DeviceEvent, 10),
	}
	hm := handlerManager{handler: h}
	evt, err := hm.eventSelect(context.TODO(), nil, nil, "")
	a.So(err, ShouldBeNil)
	a.So(evt, ShouldEqual, types.CreateEvent)
}

func TestHandlerManager_EventSelectUpdate(t *testing.T) {
	a := New(t)

	h := &handler{
		Component:    &component.Component{Ctx: GetLogger(t, "TestEventSelect")},
		applications: application.NewRedisApplicationStore(GetRedisClient(), "handler-test-event-select"),
		devices:      device.NewRedisDeviceStore(GetRedisClient(), "handler-test-event-select"),
		qEvent:       make(chan *types.DeviceEvent, 10),
	}
	hm := handlerManager{handler: h}
	appEUI := types.AppEUI([8]byte{0x10})
	devEUI := types.DevEUI([8]byte{0x10})
	l := &pb_lorawan.Device{
		AppEui: &appEUI,
		DevEui: &devEUI,
	}
	evt, err := hm.eventSelect(context.TODO(), &device.Device{AppEUI: [8]byte{0x10}, DevEUI: [8]byte{0x10}}, l, "")
	a.So(err, ShouldBeNil)
	a.So(evt, ShouldEqual, types.UpdateEvent)
}

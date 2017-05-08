// Copyright © 2017 The Things Industries B.V.

package handler

import (
	"context"
	"testing"

	pb "github.com/TheThingsNetwork/ttn/api/handler"
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

func TestHandlerManager_attrControl(t *testing.T) {
	a := New(t)

	h := &handler{
		devices: device.NewRedisDeviceStore(GetRedisClient(), "handler-test-builtin-attribute"),
	}
	hm := handlerManager{handler: h}

	//Basic
	testMap1 := map[string]string{
		"hello": "bonjour",
		"test":  "TeSt",
	}
	in := &pb.Device{Attributes: testMap1}
	hm.attrControl(in)
	a.So(in.Attributes, ShouldNotBeNil)
	a.So(in.Attributes["hello"], ShouldEqual, testMap1["hello"])
	a.So(in.Attributes["test"], ShouldEqual, testMap1["test"])

	//Past limit of 5
	testMap2 := map[string]string{
		"hello":   "bonjour",
		"test":    "TeSt",
		"beer":    "cold",
		"weather": "hot",
		"heart":   "pique",
		"square":  "trefle",
	}
	in.Attributes = testMap2
	hm.attrControl(in)
	a.So(len(in.Attributes), ShouldEqual, 5)

	//Past limit of 5 and builtin attributes
	builtinAttr := "ttn-Battery:ttn-Model"
	h.WithBuiltinAttr(builtinAttr)
	testMap3 := map[string]string{
		"hello":       "bonjour",
		"test":        "TeSt",
		"beer":        "cold",
		"weather":     "hot",
		"heart":       "pique",
		"square":      "trefle",
		"ttn-Battery": "quatre-ving-dix pourcent",
	}
	m := make(map[string]string, len(testMap3))
	for key, val := range testMap3 {
		m[key] = val
	}
	in.Attributes = m
	hm.attrControl(in)
	a.So(len(in.Attributes), ShouldEqual, 6)
	a.So(in.Attributes["ttn-Battery"], ShouldEqual, testMap3["ttn-Battery"])
}

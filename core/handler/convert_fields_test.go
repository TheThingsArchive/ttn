// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package handler

import (
	"testing"

	pb_broker "github.com/TheThingsNetwork/ttn/api/broker"

	"github.com/TheThingsNetwork/ttn/core/handler/application"
	"github.com/TheThingsNetwork/ttn/core/handler/device"
	"github.com/TheThingsNetwork/ttn/core/types"
	. "github.com/TheThingsNetwork/ttn/utils/testing"
	. "github.com/smartystreets/assertions"
)

func buildCustomUplink(appID string) (*pb_broker.DeduplicatedUplinkMessage, *types.UplinkMessage) {
	ttnUp := &pb_broker.DeduplicatedUplinkMessage{
		AppId: appID,
		DevId: "DevID-1",
	}
	appUp := &types.UplinkMessage{
		FPort:      1,
		AppID:      appID,
		DevID:      "DevID-1",
		PayloadRaw: []byte{0x08, 0x70},
	}
	return ttnUp, appUp
}

var attributes = map[string]string{
	"ttn-model": "The Things Uno",
}

func TestConvertFieldsUpCustom(t *testing.T) {
	a := New(t)
	appID := "AppID-1"

	h := &handler{
		applications: application.NewRedisApplicationStore(GetRedisClient(), "handler-test-convert-fields-up"),
		qEvent:       make(chan *types.DeviceEvent, 1),
	}

	dev := new(device.Device)
	dev.Attributes = attributes

	// No functions
	{
		ttnUp, appUp := buildCustomUplink(appID)
		err := h.ConvertFieldsUp(GetLogger(t, "TestConvertFieldsUpCustom"), ttnUp, appUp, dev)
		a.So(err, ShouldBeNil)
		a.So(appUp.PayloadFields, ShouldBeEmpty)
	}

	app := &application.Application{
		AppID:         appID,
		PayloadFormat: application.PayloadFormatCustom,
		CustomDecoder: `function Decoder (data) { return { temperature: ((data[0] << 8) | data[1]) / 100 }; }`,
	}

	// Normal flow
	{
		a.So(h.applications.Set(app), ShouldBeNil)
		defer func() {
			h.applications.Delete(appID)
		}()
		ttnUp, appUp := buildCustomUplink(appID)
		err := h.ConvertFieldsUp(GetLogger(t, "TestConvertFieldsUpCustom"), ttnUp, appUp, dev)
		a.So(err, ShouldBeNil)
		a.So(appUp.PayloadFields, ShouldResemble, map[string]interface{}{
			"temperature": 21.6,
		})
		a.So(appUp.Attributes, ShouldResemble, attributes)
	}

	// Invalidate data
	{
		app.StartUpdate()
		app.CustomValidator = `function Validator (data) { return false; }`
		h.applications.Set(app)
		ttnUp, appUp := buildCustomUplink(appID)
		err := h.ConvertFieldsUp(GetLogger(t, "TestConvertFieldsUpCustom"), ttnUp, appUp, dev)
		a.So(err, ShouldNotBeNil)
		a.So(appUp.PayloadFields, ShouldBeEmpty)
	}

	// Function error
	{
		app.StartUpdate()
		app.CustomValidator = `function Validator (data) { throw new Error("expected"); }`
		h.applications.Set(app)
		ttnUp, appUp := buildCustomUplink(appID)
		err := h.ConvertFieldsUp(GetLogger(t, "TestConvertFieldsUpCustom"), ttnUp, appUp, dev)
		a.So(err, ShouldBeNil)
		a.So(appUp.PayloadFields, ShouldBeEmpty)
		a.So(len(h.qEvent), ShouldEqual, 1)
		evt := <-h.qEvent
		_, ok := evt.Data.(types.ErrorEventData)
		a.So(ok, ShouldBeTrue)
	}

	// Invalid output error
	{
		app.StartUpdate()
		app.CustomValidator = ""
		app.CustomDecoder = `function Decoder (data) { return { infinite: 10/0 }; }`
		h.applications.Set(app)
		ttnUp, appUp := buildCustomUplink(appID)
		err := h.ConvertFieldsUp(GetLogger(t, "TestConvertFieldsUpCustom"), ttnUp, appUp, dev)
		a.So(err, ShouldBeNil)
		a.So(appUp.PayloadFields, ShouldBeEmpty)
		a.So(len(h.qEvent), ShouldEqual, 1)
		evt := <-h.qEvent
		errEvt, ok := evt.Data.(types.ErrorEventData)
		a.So(ok, ShouldBeTrue)
		a.So(errEvt.Error, ShouldContainSubstring, "cannot be marshaled")
	}
}

func buildCayenneLPPUplink(appID string) (*pb_broker.DeduplicatedUplinkMessage, *types.UplinkMessage) {
	ttnUp := &pb_broker.DeduplicatedUplinkMessage{
		AppId: appID,
		DevId: "DevID-1",
	}
	appUp := &types.UplinkMessage{
		FPort:      1,
		AppID:      appID,
		DevID:      "DevID-1",
		PayloadRaw: []byte{10, 115, 41, 239}, // Channel 10, Barometric Pressure of 1073.5
	}
	return ttnUp, appUp
}

func TestConvertFieldsUpCayenneLPP(t *testing.T) {
	a := New(t)
	appID := "AppID-1"
	ctx := GetLogger(t, "TestConvertFieldsUpCayenneLPP")

	h := &handler{
		applications: application.NewRedisApplicationStore(GetRedisClient(), "handler-test-convert-fields-up"),
		qEvent:       make(chan *types.DeviceEvent, 1),
	}

	dev := new(device.Device)
	dev.Attributes = attributes

	// No application
	{
		ttnUp, appUp := buildCayenneLPPUplink(appID)
		err := h.ConvertFieldsUp(ctx, ttnUp, appUp, dev)
		a.So(err, ShouldBeNil)
		a.So(appUp.PayloadFields, ShouldBeEmpty)
	}

	// Normal flow
	{
		app := &application.Application{
			AppID:         appID,
			PayloadFormat: application.PayloadFormatCayenneLPP,
		}
		a.So(h.applications.Set(app), ShouldBeNil)
		defer func() {
			h.applications.Delete(appID)
		}()
		ttnUp, appUp := buildCayenneLPPUplink(appID)
		err := h.ConvertFieldsUp(ctx, ttnUp, appUp, dev)
		a.So(err, ShouldBeNil)
		a.So(appUp.PayloadFields, ShouldResemble, map[string]interface{}{
			"barometric_pressure_10": float32(1073.5),
		})
		a.So(appUp.Attributes, ShouldResemble, attributes)
	}
}

func buildCustomDownlink() (*pb_broker.DownlinkMessage, *types.DownlinkMessage) {
	appEUI := types.AppEUI([8]byte{1, 2, 3, 4, 5, 6, 7, 8})
	devEUI := types.DevEUI([8]byte{1, 2, 3, 4, 5, 6, 7, 8})
	ttnDown := &pb_broker.DownlinkMessage{
		AppEui: &appEUI,
		DevEui: &devEUI,
	}
	appDown := &types.DownlinkMessage{
		FPort:         1,
		AppID:         "AppID-1",
		DevID:         "DevID-1",
		PayloadFields: map[string]interface{}{"temperature": 30},
	}
	return ttnDown, appDown
}

func TestConvertFieldsDownCustom(t *testing.T) {
	a := New(t)
	appID := "AppID-1"

	h := &handler{
		applications: application.NewRedisApplicationStore(GetRedisClient(), "handler-test-convert-fields-down"),
	}

	// No Encoder
	{
		ttnDown, appDown := buildCustomDownlink()
		err := h.ConvertFieldsDown(GetLogger(t, "TestConvertFieldsDownCustom"), appDown, ttnDown, nil)
		a.So(err, ShouldBeNil)
		a.So(appDown.PayloadRaw, ShouldBeEmpty)
	}

	// Normal flow with Encoder
	{
		h.applications.Set(&application.Application{
			AppID:         appID,
			PayloadFormat: application.PayloadFormatCustom,
			CustomEncoder: `function Encoder (payload, port){
  		return [ port, 1, 2, 3, 4, 5, 6, 7 ]
		}`,
		})
		defer func() {
			h.applications.Delete(appID)
		}()
		ttnDown, appDown := buildCustomDownlink()
		err := h.ConvertFieldsDown(GetLogger(t, "TestConvertFieldsDownCustom"), appDown, ttnDown, nil)
		a.So(err, ShouldBeNil)
		a.So(appDown.PayloadRaw, ShouldResemble, []byte{byte(appDown.FPort), 1, 2, 3, 4, 5, 6, 7})
	}
}

func TestConvertFieldsDownCustomNoPort(t *testing.T) {
	a := New(t)
	appID := "AppID-1"

	h := &handler{
		applications: application.NewRedisApplicationStore(GetRedisClient(), "handler-test-convert-fields-down"),
	}

	// No Encoder
	{
		ttnDown, appDown := buildCustomDownlink()
		err := h.ConvertFieldsDown(GetLogger(t, "TestConvertFieldsDownCustomNoPort"), appDown, ttnDown, nil)
		a.So(err, ShouldBeNil)
		a.So(appDown.PayloadRaw, ShouldBeEmpty)
	}

	// Normal flow with Encoder
	{
		h.applications.Set(&application.Application{
			AppID:         appID,
			PayloadFormat: application.PayloadFormatCustom,
			CustomEncoder: `function Encoder (payload){
  		return [ 1, 2, 3, 4, 5, 6, 7 ]
		}`,
		})
		defer func() {
			h.applications.Delete(appID)
		}()
		ttnDown, appDown := buildCustomDownlink()
		err := h.ConvertFieldsDown(GetLogger(t, "TestConvertFieldsDownCustomNoPort"), appDown, ttnDown, nil)
		a.So(err, ShouldBeNil)
		a.So(appDown.PayloadRaw, ShouldResemble, []byte{1, 2, 3, 4, 5, 6, 7})
	}
}

func buildCayenneLPPDownlink() (*pb_broker.DownlinkMessage, *types.DownlinkMessage) {
	appEUI := types.AppEUI([8]byte{1, 2, 3, 4, 5, 6, 7, 8})
	devEUI := types.DevEUI([8]byte{1, 2, 3, 4, 5, 6, 7, 8})
	ttnDown := &pb_broker.DownlinkMessage{
		AppEui: &appEUI,
		DevEui: &devEUI,
	}
	appDown := &types.DownlinkMessage{
		FPort:         1,
		AppID:         "AppID-1",
		DevID:         "DevID-1",
		PayloadFields: map[string]interface{}{"value_7": -15.6},
	}
	return ttnDown, appDown
}

func TestConvertFieldsDownCayenneLPP(t *testing.T) {
	a := New(t)
	appID := "AppID-1"
	ctx := GetLogger(t, "TestConvertFieldsDownCustom")

	h := &handler{
		applications: application.NewRedisApplicationStore(GetRedisClient(), "handler-test-convert-fields-down"),
	}

	// No Encoder
	{
		ttnDown, appDown := buildCayenneLPPDownlink()
		err := h.ConvertFieldsDown(ctx, appDown, ttnDown, nil)
		a.So(err, ShouldBeNil)
		a.So(appDown.PayloadRaw, ShouldBeEmpty)
	}

	// Normal flow with Encoder
	{
		h.applications.Set(&application.Application{
			AppID:         appID,
			PayloadFormat: application.PayloadFormatCayenneLPP,
		})
		defer func() {
			h.applications.Delete(appID)
		}()
		ttnDown, appDown := buildCayenneLPPDownlink()
		err := h.ConvertFieldsDown(ctx, appDown, ttnDown, nil)
		a.So(err, ShouldBeNil)
		a.So(appDown.PayloadRaw, ShouldResemble, []byte{7, 249, 232})
	}
}

// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package handler

import (
	"testing"

	pb_broker "github.com/TheThingsNetwork/ttn/api/broker"
	"github.com/TheThingsNetwork/ttn/mqtt"

	"github.com/TheThingsNetwork/ttn/core/handler/application"
	"github.com/TheThingsNetwork/ttn/core/types"
	. "github.com/TheThingsNetwork/ttn/utils/testing"
	. "github.com/smartystreets/assertions"
)

func buildConversionUplink() (*pb_broker.DeduplicatedUplinkMessage, *mqtt.UplinkMessage) {
	appEUI, _ := types.ParseAppEUI("0102030405060708")
	ttnUp := &pb_broker.DeduplicatedUplinkMessage{
		AppId: "AppID-1",
	}
	appUp := &mqtt.UplinkMessage{
		FPort:   1,
		AppEUI:  appEUI,
		Payload: []byte{0x08, 0x70},
	}
	return ttnUp, appUp
}

func TestConvertFields(t *testing.T) {
	a := New(t)
	appID := "AppID-1"

	h := &handler{
		applications: application.NewApplicationStore(),
	}

	// No functions
	ttnUp, appUp := buildConversionUplink()
	err := h.ConvertFields(GetLogger(t, "TestConvertFields"), ttnUp, appUp)
	a.So(err, ShouldBeNil)
	a.So(appUp.Fields, ShouldBeEmpty)

	// Normal flow
	h.applications.Set(&application.Application{
		AppID:   appID,
		Decoder: `function(data) { return { temperature: ((data[0] << 8) | data[1]) / 100 }; }`,
	})
	ttnUp, appUp = buildConversionUplink()
	err = h.ConvertFields(GetLogger(t, "TestConvertFields"), ttnUp, appUp)
	a.So(err, ShouldBeNil)

	a.So(appUp.Fields, ShouldResemble, map[string]interface{}{
		"temperature": 21.6,
	})

	// Invalidate data
	h.applications.Set(&application.Application{
		AppID:     appID,
		Decoder:   `function(data) { return { temperature: ((data[0] << 8) | data[1]) / 100 }; }`,
		Validator: `function(data) { return false; }`,
	})
	ttnUp, appUp = buildConversionUplink()
	err = h.ConvertFields(GetLogger(t, "TestConvertFields"), ttnUp, appUp)
	a.So(err, ShouldNotBeNil)
	a.So(appUp.Fields, ShouldBeEmpty)

	// Function error
	h.applications.Set(&application.Application{
		AppID:     appID,
		Decoder:   `function(data) { return { temperature: ((data[0] << 8) | data[1]) / 100 }; }`,
		Converter: `function(data) { throw "expected"; }`,
	})
	ttnUp, appUp = buildConversionUplink()
	err = h.ConvertFields(GetLogger(t, "TestConvertFields"), ttnUp, appUp)
	a.So(err, ShouldBeNil)
	a.So(appUp.Fields, ShouldBeEmpty)
}

func TestDecode(t *testing.T) {
	a := New(t)

	functions := &Functions{
		Decoder: `function(payload) {
  return {
    value: (payload[0] << 8) | payload[1]
  };
}`,
	}
	payload := []byte{0x48, 0x65}

	m, err := functions.Decode(payload)
	a.So(err, ShouldBeNil)

	size, ok := m["value"]
	a.So(ok, ShouldBeTrue)
	a.So(size, ShouldEqual, 18533)
}

func TestConvert(t *testing.T) {
	a := New(t)

	withFunction := &Functions{
		Converter: `function(data) {
  return {
    celcius: data.temperature * 2
  };
}`,
	}
	data, err := withFunction.Convert(map[string]interface{}{"temperature": 11})
	a.So(err, ShouldBeNil)
	a.So(data["celcius"], ShouldEqual, 22)

	withoutFunction := &Functions{}
	data, err = withoutFunction.Convert(map[string]interface{}{"temperature": 11})
	a.So(err, ShouldBeNil)
	a.So(data["temperature"], ShouldEqual, 11)
}

func TestValidate(t *testing.T) {
	a := New(t)

	withFunction := &Functions{
		Validator: `function(data) {
      return data.temperature < 20;
    }`,
	}
	valid, err := withFunction.Validate(map[string]interface{}{"temperature": 10})
	a.So(err, ShouldBeNil)
	a.So(valid, ShouldBeTrue)
	valid, err = withFunction.Validate(map[string]interface{}{"temperature": 30})
	a.So(err, ShouldBeNil)
	a.So(valid, ShouldBeFalse)

	withoutFunction := &Functions{}
	valid, err = withoutFunction.Validate(map[string]interface{}{"temperature": 10})
	a.So(err, ShouldBeNil)
	a.So(valid, ShouldBeTrue)
}

func TestProcess(t *testing.T) {
	a := New(t)

	functions := &Functions{
		Decoder: `function(payload) {
	return {
		temperature: payload[0],
		humidity: payload[1]
	}
}`,
		Converter: `function(data) {
	data.temperature /= 2;
	return data;
}`,
		Validator: `function(data) {
	return data.humidity >= 0 && data.humidity <= 100;
}`,
	}

	data, valid, err := functions.Process([]byte{40, 110})
	a.So(err, ShouldBeNil)
	a.So(valid, ShouldBeFalse)
	a.So(data["temperature"], ShouldEqual, 20)
	a.So(data["humidity"], ShouldEqual, 110)
}

func TestProcessInvalidFunction(t *testing.T) {
	a := New(t)

	// Empty Function
	functions := &Functions{
		Decoder: ``,
	}
	_, _, err := functions.Process([]byte{40, 110})
	a.So(err, ShouldNotBeNil)

	// Invalid Function
	functions = &Functions{
		Decoder: `this is not valid JavaScript`,
	}
	_, _, err = functions.Process([]byte{40, 110})
	a.So(err, ShouldNotBeNil)

	// Invalid return
	functions = &Functions{
		Decoder: `function(payload) { return "Hello" }`,
	}
	_, _, err = functions.Process([]byte{40, 110})
	a.So(err, ShouldNotBeNil)

	// Invalid Function
	functions = &Functions{
		Decoder:   `function(payload) { return { temperature: payload[0] } }`,
		Converter: `this is not valid JavaScript`,
	}
	_, _, err = functions.Process([]byte{40, 110})
	a.So(err, ShouldNotBeNil)

	// Invalid Return
	functions = &Functions{
		Decoder:   `function(payload) { return { temperature: payload[0] } }`,
		Converter: `function(data) { return "Hello" }`,
	}
	_, _, err = functions.Process([]byte{40, 110})
	a.So(err, ShouldNotBeNil)

	// Invalid Function
	functions = &Functions{
		Decoder:   `function(payload) { return { temperature: payload[0] } }`,
		Validator: `this is not valid JavaScript`,
	}
	_, _, err = functions.Process([]byte{40, 110})
	a.So(err, ShouldNotBeNil)

	// Invalid Return
	functions = &Functions{
		Decoder:   `function(payload) { return { temperature: payload[0] } }`,
		Validator: `function(data) { return "Hello" }`,
	}
	_, _, err = functions.Process([]byte{40, 110})
	a.So(err, ShouldNotBeNil)
}

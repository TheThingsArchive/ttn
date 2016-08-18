// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package handler

import (
	"testing"
	"time"

	pb_broker "github.com/TheThingsNetwork/ttn/api/broker"
	"github.com/TheThingsNetwork/ttn/mqtt"

	"github.com/TheThingsNetwork/ttn/core/handler/application"
	"github.com/TheThingsNetwork/ttn/core/types"
	. "github.com/TheThingsNetwork/ttn/utils/testing"
	. "github.com/smartystreets/assertions"
)

func buildConversionUplink() (*pb_broker.DeduplicatedUplinkMessage, *mqtt.UplinkMessage) {
	ttnUp := &pb_broker.DeduplicatedUplinkMessage{
		AppId: "AppID-1",
		DevId: "DevID-1",
	}
	appUp := &mqtt.UplinkMessage{
		FPort:   1,
		AppID:   "AppID-1",
		DevID:   "DevID-1",
		Payload: []byte{0x08, 0x70},
	}
	return ttnUp, appUp
}

func TestConvertFieldsUp(t *testing.T) {
	a := New(t)
	appID := "AppID-1"

	h := &handler{
		applications: application.NewApplicationStore(),
	}

	// No functions
	ttnUp, appUp := buildConversionUplink()
	err := h.ConvertFieldsUp(GetLogger(t, "TestConvertFieldsUp"), ttnUp, appUp)
	a.So(err, ShouldBeNil)
	a.So(appUp.Fields, ShouldBeEmpty)

	// Normal flow
	h.applications.Set(&application.Application{
		AppID:   appID,
		Decoder: `function(data) { return { temperature: ((data[0] << 8) | data[1]) / 100 }; }`,
	})
	ttnUp, appUp = buildConversionUplink()
	err = h.ConvertFieldsUp(GetLogger(t, "TestConvertFieldsUp"), ttnUp, appUp)
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
	err = h.ConvertFieldsUp(GetLogger(t, "TestConvertFieldsUp"), ttnUp, appUp)
	a.So(err, ShouldNotBeNil)
	a.So(appUp.Fields, ShouldBeEmpty)

	// Function error
	h.applications.Set(&application.Application{
		AppID:     appID,
		Decoder:   `function(data) { return { temperature: ((data[0] << 8) | data[1]) / 100 }; }`,
		Converter: `function(data) { throw "expected"; }`,
	})
	ttnUp, appUp = buildConversionUplink()
	err = h.ConvertFieldsUp(GetLogger(t, "TestConvertFieldsUp"), ttnUp, appUp)
	a.So(err, ShouldBeNil)
	a.So(appUp.Fields, ShouldBeEmpty)
}

func TestDecode(t *testing.T) {
	a := New(t)

	functions := &UplinkFunctions{
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

	withFunction := &UplinkFunctions{
		Converter: `function(data) {
  return {
    celcius: data.temperature * 2
  };
}`,
	}
	data, err := withFunction.Convert(map[string]interface{}{"temperature": 11})
	a.So(err, ShouldBeNil)
	a.So(data["celcius"], ShouldEqual, 22)

	withoutFunction := &UplinkFunctions{}
	data, err = withoutFunction.Convert(map[string]interface{}{"temperature": 11})
	a.So(err, ShouldBeNil)
	a.So(data["temperature"], ShouldEqual, 11)
}

func TestValidate(t *testing.T) {
	a := New(t)

	withFunction := &UplinkFunctions{
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

	withoutFunction := &UplinkFunctions{}
	valid, err = withoutFunction.Validate(map[string]interface{}{"temperature": 10})
	a.So(err, ShouldBeNil)
	a.So(valid, ShouldBeTrue)
}

func TestProcessUplink(t *testing.T) {
	a := New(t)

	functions := &UplinkFunctions{
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

func TestProcessInvalidUplinkFunction(t *testing.T) {
	a := New(t)

	// Empty Function
	functions := &UplinkFunctions{
		Decoder: ``,
	}
	_, _, err := functions.Process([]byte{40, 110})
	a.So(err, ShouldNotBeNil)

	// Invalid Function
	functions = &UplinkFunctions{
		Decoder: `this is not valid JavaScript`,
	}
	_, _, err = functions.Process([]byte{40, 110})
	a.So(err, ShouldNotBeNil)

	// Invalid return
	functions = &UplinkFunctions{
		Decoder: `function(payload) { return "Hello" }`,
	}
	_, _, err = functions.Process([]byte{40, 110})
	a.So(err, ShouldNotBeNil)

	// Invalid Function
	functions = &UplinkFunctions{
		Decoder:   `function(payload) { return { temperature: payload[0] } }`,
		Converter: `this is not valid JavaScript`,
	}
	_, _, err = functions.Process([]byte{40, 110})
	a.So(err, ShouldNotBeNil)

	// Invalid Return
	functions = &UplinkFunctions{
		Decoder:   `function(payload) { return { temperature: payload[0] } }`,
		Converter: `function(data) { return "Hello" }`,
	}
	_, _, err = functions.Process([]byte{40, 110})
	a.So(err, ShouldNotBeNil)

	// Invalid Function
	functions = &UplinkFunctions{
		Decoder:   `function(payload) { return { temperature: payload[0] } }`,
		Validator: `this is not valid JavaScript`,
	}
	_, _, err = functions.Process([]byte{40, 110})
	a.So(err, ShouldNotBeNil)

	// Invalid Return
	functions = &UplinkFunctions{
		Decoder:   `function(payload) { return { temperature: payload[0] } }`,
		Validator: `function(data) { return "Hello" }`,
	}
	_, _, err = functions.Process([]byte{40, 110})
	a.So(err, ShouldNotBeNil)

	// Invalid Object (Arrays are Objects too, but don't jive well with
	// map[string]interface{})
	functions = &UplinkFunctions{
		Decoder: `function(payload) { return [1] }`,
	}
	_, _, err = functions.Process([]byte{40, 110})
	a.So(err, ShouldNotBeNil)

	// Invalid Object (Arrays are Objects too, but don't jive well with
	// map[string]interface{})
	functions = &UplinkFunctions{
		Decoder:   `function(payload) { return { temperature: payload[0] } }`,
		Converter: `function(payload) { return [1] }`,
	}
	_, _, err = functions.Process([]byte{40, 110})
	a.So(err, ShouldNotBeNil)

	// Invalid Object (Arrays are Objects too), this should work error because
	// we're expecting a Boolean
	functions = &UplinkFunctions{
		Decoder:   `function(payload) { return { temperature: payload[0] } }`,
		Validator: `function(payload) { return [1] }`,
	}
	_, _, err = functions.Process([]byte{40, 110})
	a.So(err, ShouldNotBeNil)
}

func TestTimeoutExceeded(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	a := New(t)
	start := time.Now()

	functions := &UplinkFunctions{
		Decoder: `function(payload){ while (true) { } }`,
	}

	interrupted := make(chan bool, 2)
	go func() {
		_, _, err := functions.Process([]byte{0})
		a.So(time.Since(start), ShouldAlmostEqual, 100*time.Millisecond, 0.5e9)
		a.So(err, ShouldNotBeNil)
		interrupted <- true
	}()

	<-time.After(200 * time.Millisecond)
	a.So(interrupted, ShouldHaveLength, 1)
}

func TestEncode(t *testing.T) {
	a := New(t)

	// This function return an array of bytes (random)
	functions := &DownlinkFunctions{
		Encoder: `function test(payload){
  		return [ 1, 2, 3, 4, 5, 6, 7 ]
		}`,
	}

	// The payload is a JSON structure
	payload := map[string]interface{}{"temperature": 11}

	m, err := functions.Encode(payload)
	a.So(err, ShouldBeNil)

	a.So(m, ShouldHaveLength, 7)
	a.So(m, ShouldResemble, []byte{1, 2, 3, 4, 5, 6, 7})
}

func buildConversionDownlink() (*pb_broker.DownlinkMessage, *mqtt.DownlinkMessage) {
	appEUI := types.AppEUI([8]byte{1, 2, 3, 4, 5, 6, 7, 8})
	devEUI := types.DevEUI([8]byte{1, 2, 3, 4, 5, 6, 7, 8})
	ttnDown := &pb_broker.DownlinkMessage{
		AppEui: &appEUI,
		DevEui: &devEUI,
	}
	appDown := &mqtt.DownlinkMessage{
		FPort:  1,
		AppID:  "AppID-1",
		DevID:  "DevID-1",
		Fields: map[string]interface{}{"temperature": 30},
		// We want to "build" the payload with the content of the fields
	}
	return ttnDown, appDown
}

func TestConvertFieldsDown(t *testing.T) {
	a := New(t)
	appID := "AppID-1"

	h := &handler{
		applications: application.NewApplicationStore(),
	}

	// Case1: No Encoder
	ttnDown, appDown := buildConversionDownlink()
	err := h.ConvertFieldsDown(GetLogger(t, "TestConvertFieldsDown"), appDown, ttnDown)
	a.So(err, ShouldBeNil)
	a.So(ttnDown.Payload, ShouldBeEmpty)

	// Case2: Normal flow with Encoder
	h.applications.Set(&application.Application{
		AppID: appID,
		// Encoder takes JSON fields as argument and return the payload as []byte
		Encoder: `function test(payload){
  		return [ 1, 2, 3, 4, 5, 6, 7 ]
		}`,
	})
	ttnDown, appDown = buildConversionDownlink()
	err = h.ConvertFieldsDown(GetLogger(t, "TestConvertFieldsDown"), appDown, ttnDown)
	a.So(err, ShouldBeNil)
	a.So(ttnDown.Payload, ShouldResemble, []byte{1, 2, 3, 4, 5, 6, 7})
}

func TestProcessDownlinkInvalidFunction(t *testing.T) {
	a := New(t)

	// Empty Function
	functions := &DownlinkFunctions{
		Encoder: ``,
	}
	_, _, err := functions.Process(map[string]interface{}{"key": 11})
	a.So(err, ShouldNotBeNil)

	// Invalid Function
	functions = &DownlinkFunctions{
		Encoder: `this is not valid JavaScript`,
	}
	_, _, err = functions.Process(map[string]interface{}{"key": 11})
	a.So(err, ShouldNotBeNil)

	// Invalid return
	functions = &DownlinkFunctions{
		Encoder: `function(payload) { return "Hello" }`,
	}
	_, _, err = functions.Process(map[string]interface{}{"key": 11})
	a.So(err, ShouldNotBeNil)

	// Invalid return
	functions = &DownlinkFunctions{
		Encoder: `function(payload) { return [ 100, 2256, 7 ] }`,
	}
	_, _, err = functions.Process(map[string]interface{}{"key": 11})
	a.So(err, ShouldNotBeNil)

	// Invalid return
	functions = &DownlinkFunctions{
		Encoder: `function(payload) {
	return {
		temperature: payload[0],
		humidity: payload[1]
	}
} }`,
	}
	_, _, err = functions.Process(map[string]interface{}{"key": 11})
	a.So(err, ShouldNotBeNil)
}

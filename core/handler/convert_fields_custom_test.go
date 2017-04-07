// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package handler

import (
	"fmt"
	"testing"
	"time"

	. "github.com/smartystreets/assertions"
)

func TestCustomDecode(t *testing.T) {
	a := New(t)

	functions := &CustomUplinkFunctions{
		Decoder: `function Decoder (payload, port) {
  return {
    value: (payload[0] << 8) | payload[1],
	port: port,
  };
}`,
	}
	payload := []byte{0x48, 0x65}

	m, err := functions.decode(payload, 12)
	a.So(err, ShouldBeNil)

	size, ok := m["value"]
	a.So(ok, ShouldBeTrue)
	a.So(size, ShouldEqual, 18533)

	port, ok := m["port"]
	a.So(ok, ShouldBeTrue)
	a.So(port, ShouldEqual, 12)
}

func TestCustomConvert(t *testing.T) {
	a := New(t)

	withFunction := &CustomUplinkFunctions{
		Converter: `function Converter (data, port) {
  return {
    celcius: data.temperature * 2
	port: port,
  };
}`,
	}
	data, err := withFunction.convert(map[string]interface{}{"temperature": 11}, 33)
	a.So(err, ShouldBeNil)
	a.So(data["celcius"], ShouldEqual, 22)
	a.So(data["port"], ShouldEqual, 33)

	withoutFunction := &CustomUplinkFunctions{}
	data, err = withoutFunction.convert(map[string]interface{}{"temperature": 11}, 33)
	a.So(err, ShouldBeNil)
	a.So(data["temperature"], ShouldEqual, 11)
}

func TestCustomValidate(t *testing.T) {
	a := New(t)

	withFunction := &CustomUplinkFunctions{
		Validator: `function Validator (data) {
      return data.temperature < 20;
    }`,
	}
	valid, err := withFunction.validate(map[string]interface{}{"temperature": 10}, 1)
	a.So(err, ShouldBeNil)
	a.So(valid, ShouldBeTrue)
	valid, err = withFunction.validate(map[string]interface{}{"temperature": 30}, 1)
	a.So(err, ShouldBeNil)
	a.So(valid, ShouldBeFalse)

	withoutFunction := &CustomUplinkFunctions{}
	valid, err = withoutFunction.validate(map[string]interface{}{"temperature": 10}, 1)
	a.So(err, ShouldBeNil)
	a.So(valid, ShouldBeTrue)
}

func TestCustomDecodeUplink(t *testing.T) {
	a := New(t)

	functions := &CustomUplinkFunctions{
		Decoder: `function Decoder (payload) {
	return {
		temperature: payload[0],
		humidity: payload[1]
	}
}`,
		Converter: `function Converter (data) {
	data.temperature /= 2;
	return data;
}`,
		Validator: `function Validator (data) {
	return data.humidity >= 0 && data.humidity <= 100;
}`,
	}

	data, valid, err := functions.Decode([]byte{40, 110}, 1)
	a.So(err, ShouldBeNil)
	a.So(valid, ShouldBeFalse)
	a.So(data["temperature"], ShouldEqual, 20)
	a.So(data["humidity"], ShouldEqual, 110)
}

func TestCustomDecodeInvalidUplinkFunction(t *testing.T) {
	a := New(t)

	// Empty Function
	functions := &CustomUplinkFunctions{
		Decoder: ``,
	}
	_, _, err := functions.Decode([]byte{40, 110}, 1)
	a.So(err, ShouldBeNil)

	// Invalid Function
	functions = &CustomUplinkFunctions{
		Decoder: `this is not valid JavaScript`,
	}
	_, _, err = functions.Decode([]byte{40, 110}, 1)
	a.So(err, ShouldNotBeNil)

	// Invalid return
	functions = &CustomUplinkFunctions{
		Decoder: `function Decoder (payload) { return "Hello" }`,
	}
	_, _, err = functions.Decode([]byte{40, 110}, 1)
	a.So(err, ShouldNotBeNil)

	// Invalid Function
	functions = &CustomUplinkFunctions{
		Decoder:   `function Decoder (payload) { return { temperature: payload[0] } }`,
		Converter: `this is not valid JavaScript`,
	}
	_, _, err = functions.Decode([]byte{40, 110}, 1)
	a.So(err, ShouldNotBeNil)

	// Invalid Return
	functions = &CustomUplinkFunctions{
		Decoder:   `function Decoder (payload) { return { temperature: payload[0] } }`,
		Converter: `function Converter (data) { return "Hello" }`,
	}
	_, _, err = functions.Decode([]byte{40, 110}, 1)
	a.So(err, ShouldNotBeNil)

	// Invalid Function
	functions = &CustomUplinkFunctions{
		Decoder:   `function Decoder (payload) { return { temperature: payload[0] } }`,
		Validator: `this is not valid JavaScript`,
	}
	_, _, err = functions.Decode([]byte{40, 110}, 1)
	a.So(err, ShouldNotBeNil)

	// Invalid Return
	functions = &CustomUplinkFunctions{
		Decoder:   `function Decoder (payload) { return { temperature: payload[0] } }`,
		Validator: `function Validator (data) { return "Hello" }`,
	}
	_, _, err = functions.Decode([]byte{40, 110}, 1)
	a.So(err, ShouldNotBeNil)

	// Invalid Object (Arrays are Objects too, but don't jive well with
	// map[string]interface{})
	functions = &CustomUplinkFunctions{
		Decoder: `function Decoder (payload) { return [1] }`,
	}
	_, _, err = functions.Decode([]byte{40, 110}, 1)
	a.So(err, ShouldNotBeNil)

	// Invalid Object (Arrays are Objects too, but don't jive well with
	// map[string]interface{})
	functions = &CustomUplinkFunctions{
		Decoder:   `function Decoder (payload) { return { temperature: payload[0] } }`,
		Converter: `function Converter (payload) { return [1] }`,
	}
	_, _, err = functions.Decode([]byte{40, 110}, 1)
	a.So(err, ShouldNotBeNil)

	// Invalid Object (Arrays are Objects too), this should work error because
	// we're expecting a Boolean
	functions = &CustomUplinkFunctions{
		Decoder:   `function Decoder (payload) { return { temperature: payload[0] } }`,
		Validator: `function Validator (payload) { return [1] }`,
	}
	_, _, err = functions.Decode([]byte{40, 110}, 1)
	a.So(err, ShouldNotBeNil)
}

func TestCustomTimeoutExceeded(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	a := New(t)
	start := time.Now()

	functions := &CustomUplinkFunctions{
		Decoder: `function(payload){ while (true) { } }`,
	}

	interrupted := make(chan bool, 2)
	go func() {
		_, _, err := functions.Decode([]byte{0}, 1)
		a.So(time.Since(start), ShouldAlmostEqual, 100*time.Millisecond, 0.5e9)
		a.So(err, ShouldNotBeNil)
		interrupted <- true
	}()

	<-time.After(200 * time.Millisecond)
	a.So(interrupted, ShouldHaveLength, 1)
}

func TestCustomEncode(t *testing.T) {
	a := New(t)

	// This function return an array of bytes (random)
	functions := &CustomDownlinkFunctions{
		Encoder: `function Encoder (payload){
  		return [ 1, 2, 3, 4, 5, 6, 7 ]
		}`,
	}

	// The payload is a JSON structure
	payload := map[string]interface{}{"temperature": 11}

	m, err := functions.encode(payload, 1)
	a.So(err, ShouldBeNil)

	a.So(m, ShouldHaveLength, 7)
	a.So(m, ShouldResemble, []byte{1, 2, 3, 4, 5, 6, 7})

	// Return int type
	functions = &CustomDownlinkFunctions{
		Encoder: `function Encoder (payload, port) { var x = [1, 2, 3 ]; return [ x.length || 0 ] }`,
	}
	_, _, err = functions.Encode(map[string]interface{}{"key": 11}, 1)
	a.So(err, ShouldBeNil)
}

func TestCustomProcessDownlinkInvalidFunction(t *testing.T) {
	a := New(t)

	// Empty Function
	functions := &CustomDownlinkFunctions{
		Encoder: ``,
	}
	_, _, err := functions.Encode(map[string]interface{}{"key": 11}, 1)
	a.So(err, ShouldNotBeNil)

	// Invalid Function
	functions = &CustomDownlinkFunctions{
		Encoder: `this is not valid JavaScript`,
	}
	_, _, err = functions.Encode(map[string]interface{}{"key": 11}, 1)
	a.So(err, ShouldNotBeNil)

	// Invalid return
	functions = &CustomDownlinkFunctions{
		Encoder: `function Encoder (payload) { return "Hello" }`,
	}
	_, _, err = functions.Encode(map[string]interface{}{"key": 11}, 1)
	a.So(err, ShouldNotBeNil)

	// Invalid return
	functions = &CustomDownlinkFunctions{
		Encoder: `function Encoder (payload) { return [ 100, 2256, 7 ] }`,
	}
	_, _, err = functions.Encode(map[string]interface{}{"key": 11}, 1)
	a.So(err, ShouldNotBeNil)

	// Invalid return
	functions = &CustomDownlinkFunctions{
		Encoder: `function Encoder (payload) { return [0, -1, "blablabla"] }`,
	}
	_, _, err = functions.Encode(map[string]interface{}{"key": 11}, 1)
	a.So(err, ShouldNotBeNil)

	// Invalid return
	functions = &CustomDownlinkFunctions{
		Encoder: `function Encoder (payload) {
	return {
		temperature: payload[0],
		humidity: payload[1]
	}
} }`,
	}
	_, _, err = functions.Encode(map[string]interface{}{"key": 11}, 1)
	a.So(err, ShouldNotBeNil)

	functions = &CustomDownlinkFunctions{
		Encoder: `function Encoder (payload) { return [ 1, 1.5 ] }`,
	}
	_, _, err = functions.Encode(map[string]interface{}{"key": 11}, 1)
	a.So(err, ShouldNotBeNil)
}

func TestEncodeCharCode(t *testing.T) {
	a := New(t)

	// Return array of character codes
	functions := &CustomDownlinkFunctions{
		Encoder: `function Encoder(obj) {
			return "Hi".split('').map(function(char) {
				return char.charCodeAt();
			});
		}`,
	}
	val, _, err := functions.Encode(map[string]interface{}{"key": 11}, 1)
	a.So(err, ShouldBeNil)

	fmt.Println("VALUE", val)
}

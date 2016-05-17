// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package fields

import (
	"testing"

	. "github.com/smartystreets/assertions"
)

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

// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package fields

import (
	"errors"
	"testing"
	"time"

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

	// Invalid Object (Arrays are Objects too, but don't jive well with
	// map[string]interface{})
	functions = &Functions{
		Decoder: `function(payload) { return [1] }`,
	}
	_, _, err = functions.Process([]byte{40, 110})
	a.So(err, ShouldNotBeNil)

	// Invalid Object (Arrays are Objects too, but don't jive well with
	// map[string]interface{})
	functions = &Functions{
		Decoder:   `function(payload) { return { temperature: payload[0] } }`,
		Converter: `function(payload) { return [1] }`,
	}
	_, _, err = functions.Process([]byte{40, 110})
	a.So(err, ShouldNotBeNil)

	// Invalid Object (Arrays are Objects too), this should work error because
	// we're expecting a Boolean
	functions = &Functions{
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
	functions := &Functions{
		Decoder: `function(payload){ while (true) { } }`,
	}

	go func() {
		time.Sleep(4 * time.Second)
		panic(errors.New("Payload function was not interrupted"))
	}()

	_, _, err := functions.Process([]byte{0})
	a.So(time.Since(start), ShouldAlmostEqual, time.Second, 0.5e9)
	a.So(err, ShouldNotBeNil)
}

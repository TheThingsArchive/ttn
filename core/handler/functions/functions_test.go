// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package functions

import (
	"testing"
	"time"

	pb_handler "github.com/TheThingsNetwork/ttn/api/handler"
	"github.com/robertkrimen/otto"

	. "github.com/smartystreets/assertions"
)

func TestRunCode(t *testing.T) {
	a := New(t)

	logger := NewEntryLogger()
	foo := 10
	env := map[string]interface{}{
		"foo": foo,
		"bar": "baz",
	}

	code := `
		(function (foo, bar) {
			console.log("hello", foo, bar)
			return foo
		})(foo,bar)
	`

	val, err := RunCode("test", code, env, time.Second, logger)
	a.So(err, ShouldBeNil)
	a.So(val, ShouldEqual, foo)
	a.So(logger.Entries(), ShouldResemble, []*pb_handler.LogEntry{
		&pb_handler.LogEntry{
			Function: "test",
			Fields:   []string{`"hello"`, "10", `"baz"`},
		},
	})
}

func TestRunCodeThrow(t *testing.T) {
	a := New(t)

	logger := NewEntryLogger()
	env := map[string]interface{}{}

	code := `
		(function () {
			throw new Error("This is an error")
			return 10
		})()
	`

	_, err := RunCode("test", code, env, time.Second, logger)
	a.So(err, ShouldNotBeNil)
}

var result string

func BenchmarkJSON(b *testing.B) {
	v, _ := otto.ToValue("foo")
	var r string
	for n := 0; n < b.N; n++ {
		r = JSON(v)
	}
	result = r
}

func TestRunInvalidCode(t *testing.T) {
	a := New(t)

	logger := NewEntryLogger()
	foo := 10
	env := map[string]interface{}{
		"foo": foo,
		"bar": "baz",
	}

	code := `
		(function (foo, bar) {
			derp
		})(foo,bar)
	`

	_, err := RunCode("test", code, env, time.Second, logger)
	a.So(err, ShouldNotBeNil)
}

func TestRunDangerousCode(t *testing.T) {
	a := New(t)

	logger := NewEntryLogger()
	env := map[string]interface{}{}

	code := `
		(function () {
			var obj = {foo: "bar"};
			obj.ob = obj;
			return obj;
		})()
	`

	out, err := RunCode("test", code, env, time.Second, logger)
	a.So(err, ShouldNotBeNil)
	a.So(out, ShouldBeNil)
}

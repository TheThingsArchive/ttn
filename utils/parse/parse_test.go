package parse

import (
	"testing"

	. "github.com/smartystreets/assertions"
)

func TestPort(t *testing.T) {
	a := New(t)

	p, err := Port("localhost")
	a.So(p, ShouldEqual, 0)
	a.So(err, ShouldNotBeNil)

	p, err = Port("localhost:test")
	a.So(p, ShouldEqual, 0)
	a.So(err, ShouldNotBeNil)

	p, err = Port("localhost:-1234")
	a.So(p, ShouldEqual, 0)
	a.So(err, ShouldNotBeNil)

	p, err = Port(":1234")
	a.So(p, ShouldEqual, 1234)
	a.So(err, ShouldBeNil)

	p, err = Port("localhost:1234")
	a.So(p, ShouldEqual, 1234)
	a.So(err, ShouldBeNil)

	p, err = Port("127.0.0.1:1234")
	a.So(p, ShouldEqual, 1234)
	a.So(err, ShouldBeNil)

	p, err = Port("user:pass@host:1234")
	a.So(p, ShouldEqual, 1234)
	a.So(err, ShouldBeNil)
}

func TestPayloadFunction(t *testing.T) {
	a := New(t)

	function := `
		function Myfunction(foo, bar) {
			if (i = 0; i < 2; i++) // missing closing brace
				return 1;
			}
			return 2;
		}
	`
	err := PayloadFunction(function)
	a.So(err, ShouldNotBeNil)

	function = `
		function Myfunction(foo bar) { // missing coma to separate args
			if (i = 0; i < 2; i++) {
				return 1;
			}
			return 2;
		}
	`
	err = PayloadFunction(function)
	a.So(err, ShouldNotBeNil)

	function = `
		function Myfunction(foo, bar) {
			retun 1; // Syntax mispelling
		}
	`
	err = PayloadFunction(function)
	a.So(err, ShouldNotBeNil)

	// Valid javascript syntax
	function = `
		function Myfunction(foo, bar) {
			return 1;
		}
	`
	err = PayloadFunction(function)
	a.So(err, ShouldBeNil)
}

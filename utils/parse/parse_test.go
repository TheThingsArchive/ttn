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

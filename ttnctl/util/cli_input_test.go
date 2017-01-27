package util

import (
	"errors"
	"testing"

	. "github.com/smartystreets/assertions"
)

func TestIsErrOutOfRange(t *testing.T) {
	a := New(t)

	err := errors.New("Dummy error")
	ok := IsErrOutOfRange(err)
	a.So(ok, ShouldBeFalse)

	err = errors.New("value out of range")
	ok = IsErrOutOfRange(err)
	a.So(ok, ShouldBeTrue)

	err = errors.New("whatever message: value out of range")
	ok = IsErrOutOfRange(err)
	a.So(ok, ShouldBeTrue)
}

func TestParsePort(t *testing.T) {
	a := New(t)

	n, err := parsePort("121212")
	a.So(err, ShouldNotBeNil)
	a.So(n, ShouldEqual, 0)

	n, err = parsePort("-12")
	a.So(err, ShouldNotBeNil)
	a.So(n, ShouldEqual, 0)

	n, err = parsePort("test")
	a.So(err, ShouldNotBeNil)
	a.So(n, ShouldEqual, 0)

	n, err = parsePort("1")
	a.So(err, ShouldBeNil)
	a.So(n, ShouldEqual, 1)
}

func TestParseFields(t *testing.T) {
	a := New(t)

	f, err := parseFields("test")
	a.So(err, ShouldNotBeNil)
	a.So(f, ShouldBeNil)

	// Syntax error: Invalid JSON object
	f, err = parseFields(`{ "test": "value"`)
	a.So(err, ShouldNotBeNil)
	a.So(f, ShouldBeNil)

	f, err = parseFields(`{ "test": "value"}`)
	a.So(err, ShouldBeNil)
	a.So(f, ShouldResemble, map[string]interface{}{"test": "value"})
}

func TestParsePayload(t *testing.T) {
	a := New(t)

	// Invalid input: Not hexadecimal
	p, err := parsePayload("test")
	a.So(err, ShouldNotBeNil)
	a.So(p, ShouldBeNil)

	// Invalid input: not hexadecimal
	p, err = parsePayload("123")
	a.So(err, ShouldEqual, ErrInvalidPayload)
	a.So(p, ShouldBeNil)

	p, err = parsePayload("1234")
	a.So(err, ShouldBeNil)
}

// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package storage

import (
	"testing"

	"github.com/smartystreets/assertions"
)

func TestFloat32(t *testing.T) {
	a := assertions.New(t)
	a.So(FormatFloat32(123.456), assertions.ShouldEqual, "123.456")
	f32, err := ParseFloat32("123.456")
	a.So(err, assertions.ShouldBeNil)
	a.So(f32, assertions.ShouldEqual, 123.456)
}

func TestFloat64(t *testing.T) {
	a := assertions.New(t)
	a.So(FormatFloat64(123.456), assertions.ShouldEqual, "123.456")
	f64, err := ParseFloat64("123.456")
	a.So(err, assertions.ShouldBeNil)
	a.So(f64, assertions.ShouldEqual, 123.456)
}

func TestInt32(t *testing.T) {
	a := assertions.New(t)
	a.So(FormatInt32(-123456), assertions.ShouldEqual, "-123456")
	i32, err := ParseInt32("-123456")
	a.So(err, assertions.ShouldBeNil)
	a.So(i32, assertions.ShouldEqual, -123456)
}

func TestInt64(t *testing.T) {
	a := assertions.New(t)
	a.So(FormatInt64(-123456), assertions.ShouldEqual, "-123456")
	i64, err := ParseInt64("-123456")
	a.So(err, assertions.ShouldBeNil)
	a.So(i64, assertions.ShouldEqual, -123456)
}

func TestUint32(t *testing.T) {
	a := assertions.New(t)
	a.So(FormatUint32(123456), assertions.ShouldEqual, "123456")
	i32, err := ParseUint32("123456")
	a.So(err, assertions.ShouldBeNil)
	a.So(i32, assertions.ShouldEqual, 123456)
}

func TestUint64(t *testing.T) {
	a := assertions.New(t)
	a.So(FormatUint64(123456), assertions.ShouldEqual, "123456")
	i64, err := ParseUint64("123456")
	a.So(err, assertions.ShouldBeNil)
	a.So(i64, assertions.ShouldEqual, 123456)
}

func TestBool(t *testing.T) {
	a := assertions.New(t)
	a.So(FormatBool(true), assertions.ShouldEqual, "true")
	b, err := ParseBool("true")
	a.So(err, assertions.ShouldBeNil)
	a.So(b, assertions.ShouldEqual, true)
	a.So(FormatBool(false), assertions.ShouldEqual, "false")
	b, err = ParseBool("false")
	a.So(err, assertions.ShouldBeNil)
	a.So(b, assertions.ShouldEqual, false)
}

func TestBytes(t *testing.T) {
	a := assertions.New(t)
	a.So(FormatBytes([]byte{0x12, 0x34, 0xcd, 0xef}), assertions.ShouldEqual, "1234CDEF")
	i64, err := ParseBytes("1234CDEF")
	a.So(err, assertions.ShouldBeNil)
	a.So(i64, assertions.ShouldResemble, []byte{0x12, 0x34, 0xcd, 0xef})
}

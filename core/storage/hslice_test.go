package storage

import (
	"testing"

	"github.com/smartystreets/assertions"
)

func TestHSliceFloat32(t *testing.T) {
	a := assertions.New(t)
	s := NewHSlice()
	_, err := s.GetFloat32("Float32")
	a.So(err, assertions.ShouldEqual, ErrDoesNotExist)
	s.SetFloat32("Float32", 123.456)
	f32, err := s.GetFloat32("Float32")
	a.So(err, assertions.ShouldBeNil)
	a.So(f32, assertions.ShouldEqual, 123.456)
}

func TestHSliceFloat64(t *testing.T) {
	a := assertions.New(t)
	s := NewHSlice()
	_, err := s.GetFloat64("Float64")
	a.So(err, assertions.ShouldEqual, ErrDoesNotExist)
	s.SetFloat64("Float64", 123.456)
	f32, err := s.GetFloat64("Float64")
	a.So(err, assertions.ShouldBeNil)
	a.So(f32, assertions.ShouldEqual, 123.456)
}

func TestHSliceInt32(t *testing.T) {
	a := assertions.New(t)
	s := NewHSlice()
	_, err := s.GetInt32("Int32")
	a.So(err, assertions.ShouldEqual, ErrDoesNotExist)
	s.SetInt32("Int32", -123456)
	res, err := s.GetInt32("Int32")
	a.So(err, assertions.ShouldBeNil)
	a.So(res, assertions.ShouldEqual, -123456)
}

func TestHSliceInt64(t *testing.T) {
	a := assertions.New(t)
	s := NewHSlice()
	_, err := s.GetInt64("Int64")
	a.So(err, assertions.ShouldEqual, ErrDoesNotExist)
	s.SetInt64("Int64", -123456)
	res, err := s.GetInt64("Int64")
	a.So(err, assertions.ShouldBeNil)
	a.So(res, assertions.ShouldEqual, -123456)
}

func TestHSliceUint32(t *testing.T) {
	a := assertions.New(t)
	s := NewHSlice()
	_, err := s.GetUint32("Uint32")
	a.So(err, assertions.ShouldEqual, ErrDoesNotExist)
	s.SetUint32("Uint32", 123456)
	res, err := s.GetUint32("Uint32")
	a.So(err, assertions.ShouldBeNil)
	a.So(res, assertions.ShouldEqual, 123456)
}

func TestHSliceUint64(t *testing.T) {
	a := assertions.New(t)
	s := NewHSlice()
	_, err := s.GetUint64("Uint64")
	a.So(err, assertions.ShouldEqual, ErrDoesNotExist)
	s.SetUint64("Uint64", 123456)
	res, err := s.GetUint64("Uint64")
	a.So(err, assertions.ShouldBeNil)
	a.So(res, assertions.ShouldEqual, 123456)
}

func TestHSliceBool(t *testing.T) {
	a := assertions.New(t)
	s := NewHSlice()
	_, err := s.GetBool("Bool")
	a.So(err, assertions.ShouldEqual, ErrDoesNotExist)
	s.SetBool("Bool", true)
	res, err := s.GetBool("Bool")
	a.So(err, assertions.ShouldBeNil)
	a.So(res, assertions.ShouldEqual, true)
	s.SetBool("Bool", false)
	res, err = s.GetBool("Bool")
	a.So(err, assertions.ShouldBeNil)
	a.So(res, assertions.ShouldEqual, false)
}

func TestHSliceString(t *testing.T) {
	a := assertions.New(t)
	s := NewHSlice()
	_, err := s.GetString("String")
	a.So(err, assertions.ShouldEqual, ErrDoesNotExist)
	s.SetString("String", "the string")
	res, err := s.GetString("String")
	a.So(err, assertions.ShouldBeNil)
	a.So(res, assertions.ShouldEqual, "the string")
}

func TestHSliceBytes(t *testing.T) {
	a := assertions.New(t)
	s := NewHSlice()
	_, err := s.GetBytes("Bytes")
	a.So(err, assertions.ShouldEqual, ErrDoesNotExist)
	s.SetBytes("Bytes", []byte{0x12, 0x34, 0xcd, 0xef})
	res, err := s.GetBytes("Bytes")
	a.So(err, assertions.ShouldBeNil)
	a.So(res, assertions.ShouldResemble, []byte{0x12, 0x34, 0xcd, 0xef})
}

func TestHSliceSlice(t *testing.T) {
	a := assertions.New(t)
	s := NewHSlice()
	s.SetString("String", "the string")
	slice := s.MarshalHSlice()
	a.So(slice, assertions.ShouldResemble, []string{"String", "the string"})
}

func TestHSliceFrom(t *testing.T) {
	a := assertions.New(t)
	s := NewHSlice()
	err := s.UnmarshalHSlice([]string{"String", "the string"})
	a.So(err, assertions.ShouldBeNil)
	res, err := s.GetString("String")
	a.So(err, assertions.ShouldBeNil)
	a.So(res, assertions.ShouldEqual, "the string")
	err = s.UnmarshalHSlice([]string{"String", "the string", "another string"})
	a.So(err, assertions.ShouldEqual, ErrInvalidLength)
}

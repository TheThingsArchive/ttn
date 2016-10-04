package scramble

import (
	"testing"

	. "github.com/smartystreets/assertions"
)

func TestFloat32(t *testing.T) {
	a := New(t)

	var val, delta float32 = 13.234567, 1.54123

	newVal, err := Float32(val, delta)
	a.So(err, ShouldBeNil)
	a.So(newVal, ShouldBeBetween, val-delta, val+delta)
}

func TestFloat64(t *testing.T) {
	a := New(t)

	var val, delta float64 = 13.234567, 1.54123

	newVal, err := Float64(val, float32(delta))
	a.So(err, ShouldBeNil)
	a.So(newVal, ShouldBeBetween, val-delta, val+delta)
}

func TestInt(t *testing.T) {
	a := New(t)

	var val, delta int = 1234, 13

	newVal, err := Int(val, delta)
	a.So(err, ShouldBeNil)
	a.So(newVal, ShouldBeBetween, val-delta, val+delta)
}

func TestInt32(t *testing.T) {
	a := New(t)

	var val, delta int32 = 1234, 13

	newVal, err := Int32(val, delta)
	a.So(err, ShouldBeNil)
	a.So(newVal, ShouldBeBetween, val-delta, val+delta)
}

func TestInt64(t *testing.T) {
	a := New(t)

	var val, delta int64 = 1234, 13

	newVal, err := Int64(val, delta)
	a.So(err, ShouldBeNil)
	a.So(newVal, ShouldBeBetween, val-delta, val+delta)
}

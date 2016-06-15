package types

import (
	"testing"

	. "github.com/smartystreets/assertions"
)

func TestDevAddr(t *testing.T) {
	a := New(t)

	// Setup
	addr := DevAddr{1, 2, 254, 255}
	str := "0102FEFF"
	bin := []byte{0x01, 0x02, 0xfe, 0xff}

	// Bytes
	a.So(addr.Bytes(), ShouldResemble, bin)

	// String
	a.So(addr.String(), ShouldEqual, str)

	// MarshalText
	mtOut, err := addr.MarshalText()
	a.So(err, ShouldBeNil)
	a.So(mtOut, ShouldResemble, []byte(str))

	// MarshalBinary
	mbOut, err := addr.MarshalBinary()
	a.So(err, ShouldBeNil)
	a.So(mbOut, ShouldResemble, bin)

	// Marshal
	mOut, err := addr.Marshal()
	a.So(err, ShouldBeNil)
	a.So(mOut, ShouldResemble, bin)

	// Parse
	pOut, err := ParseDevAddr(str)
	a.So(err, ShouldBeNil)
	a.So(pOut, ShouldEqual, addr)

	// UnmarshalText
	utOut := &DevAddr{}
	err = utOut.UnmarshalText([]byte(str))
	a.So(err, ShouldBeNil)
	a.So(*utOut, ShouldEqual, addr)

	// UnmarshalBinary
	ubOut := &DevAddr{}
	err = ubOut.UnmarshalBinary(bin)
	a.So(err, ShouldBeNil)
	a.So(*ubOut, ShouldEqual, addr)

	// Unmarshal
	uOut := &DevAddr{}
	err = uOut.Unmarshal(bin)
	a.So(err, ShouldBeNil)
	a.So(*uOut, ShouldEqual, addr)

	// IsEmpty
	var empty DevAddr
	a.So(empty.IsEmpty(), ShouldEqual, true)
	a.So(addr.IsEmpty(), ShouldEqual, false)
}

func TestDevAddrMask(t *testing.T) {
	a := New(t)
	d1 := DevAddr{255, 255, 255, 255}
	a.So(d1.Mask(1), ShouldEqual, DevAddr{128, 0, 0, 0})
	a.So(d1.Mask(2), ShouldEqual, DevAddr{192, 0, 0, 0})
	a.So(d1.Mask(3), ShouldEqual, DevAddr{224, 0, 0, 0})
	a.So(d1.Mask(4), ShouldEqual, DevAddr{240, 0, 0, 0})
	a.So(d1.Mask(5), ShouldEqual, DevAddr{248, 0, 0, 0})
	a.So(d1.Mask(6), ShouldEqual, DevAddr{252, 0, 0, 0})
	a.So(d1.Mask(7), ShouldEqual, DevAddr{254, 0, 0, 0})
	a.So(d1.Mask(8), ShouldEqual, DevAddr{255, 0, 0, 0})
}

func TestDevAddrSetPrefix(t *testing.T) {
	a := New(t)
	d1 := DevAddr{255, 255, 255, 255}
	a.So(d1.SetPrefix(DevAddr{1, 2, 3, 4}, 7), ShouldEqual, DevAddr{1, 255, 255, 255})
}

func TestDevAddrHasPrefix(t *testing.T) {
	a := New(t)
	a.So(DevAddr{1, 2, 3, 4}.HasPrefix(0, []byte{}), ShouldBeTrue)
	a.So(DevAddr{1, 2, 3, 4}.HasPrefix(32, []byte{1, 2, 3, 4}), ShouldBeTrue)
	a.So(DevAddr{1, 2, 3, 4}.HasPrefix(31, []byte{1, 2, 3, 4}), ShouldBeTrue)
	a.So(DevAddr{1, 2, 3, 4}.HasPrefix(31, []byte{2, 2, 3, 4}), ShouldBeFalse)
	a.So(DevAddr{1, 2, 3, 4}.HasPrefix(31, []byte{1, 1, 3, 4}), ShouldBeFalse)
	a.So(DevAddr{1, 2, 3, 4}.HasPrefix(15, []byte{1, 1}), ShouldBeFalse)
}

func TestParseDevAddrPrefix(t *testing.T) {
	a := New(t)
	addr, length, err := ParseDevAddrPrefix("XYZ")
	a.So(err, ShouldNotBeNil)
	addr, length, err = ParseDevAddrPrefix("00/bla")
	a.So(err, ShouldNotBeNil)
	addr, length, err = ParseDevAddrPrefix("00/1")
	a.So(err, ShouldNotBeNil)
	addr, length, err = ParseDevAddrPrefix("01020304/1")
	a.So(err, ShouldBeNil)
	a.So(addr, ShouldEqual, DevAddr{0, 0, 0, 0})
	a.So(length, ShouldEqual, 1)
	addr, length, err = ParseDevAddrPrefix("ff020304/1")
	a.So(err, ShouldBeNil)
	a.So(addr, ShouldEqual, DevAddr{128, 0, 0, 0})
	a.So(length, ShouldEqual, 1)
}

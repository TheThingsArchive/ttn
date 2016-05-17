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

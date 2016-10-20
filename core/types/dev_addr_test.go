// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

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

	// MarshalTo
	bOut := make([]byte, 4)
	_, err = addr.MarshalTo(bOut)
	a.So(err, ShouldBeNil)
	a.So(bOut, ShouldResemble, bin)

	// Size
	s := addr.Size()
	a.So(s, ShouldEqual, 4)

	// Parse
	pOut, err := ParseDevAddr(str)
	a.So(err, ShouldBeNil)
	a.So(pOut, ShouldEqual, addr)

	_, err = ParseDevAddr("no-dev-addr")
	a.So(err, ShouldNotBeNil)

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

	// Empty
	var empty DevAddr
	a.So(empty.IsEmpty(), ShouldEqual, true)
	a.So(addr.IsEmpty(), ShouldEqual, false)
	a.So(empty.String(), ShouldEqual, "")
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

func TestDevAddrWithPrefix(t *testing.T) {
	a := New(t)
	addr := DevAddr{0xAA, 0xAA, 0xAA, 0xAA}
	prefix := DevAddr{0x55, 0x55, 0x55, 0x55}
	a.So(addr.WithPrefix(DevAddrPrefix{prefix, 4}), ShouldEqual, DevAddr{0x5A, 0xAA, 0xAA, 0xAA})
	a.So(addr.WithPrefix(DevAddrPrefix{prefix, 8}), ShouldEqual, DevAddr{0x55, 0xAA, 0xAA, 0xAA})
	a.So(addr.WithPrefix(DevAddrPrefix{prefix, 12}), ShouldEqual, DevAddr{0x55, 0x5A, 0xAA, 0xAA})
	a.So(addr.WithPrefix(DevAddrPrefix{prefix, 16}), ShouldEqual, DevAddr{0x55, 0x55, 0xAA, 0xAA})
}

func TestDevAddrHasPrefix(t *testing.T) {
	a := New(t)
	addr := DevAddr{1, 2, 3, 4}
	a.So(addr.HasPrefix(DevAddrPrefix{DevAddr{0, 0, 0, 0}, 0}), ShouldBeTrue)
	a.So(addr.HasPrefix(DevAddrPrefix{DevAddr{1, 2, 3, 0}, 24}), ShouldBeTrue)
	a.So(addr.HasPrefix(DevAddrPrefix{DevAddr{2, 2, 3, 4}, 31}), ShouldBeFalse)
	a.So(addr.HasPrefix(DevAddrPrefix{DevAddr{1, 1, 3, 4}, 31}), ShouldBeFalse)
	a.So(addr.HasPrefix(DevAddrPrefix{DevAddr{1, 1, 1, 1}, 15}), ShouldBeFalse)
}

func TestParseDevAddrPrefix(t *testing.T) {
	a := New(t)
	prefix, err := ParseDevAddrPrefix("XYZ")
	a.So(err, ShouldNotBeNil)
	prefix, err = ParseDevAddrPrefix("00/bla")
	a.So(err, ShouldNotBeNil)
	prefix, err = ParseDevAddrPrefix("00/1")
	a.So(err, ShouldNotBeNil)
	prefix, err = ParseDevAddrPrefix("01020304/1")
	a.So(err, ShouldBeNil)
	a.So(prefix.DevAddr, ShouldEqual, DevAddr{0, 0, 0, 0})
	a.So(prefix.Length, ShouldEqual, 1)
	prefix, err = ParseDevAddrPrefix("ff020304/1")
	a.So(err, ShouldBeNil)
	a.So(prefix.DevAddr, ShouldEqual, DevAddr{128, 0, 0, 0})
	a.So(prefix.Length, ShouldEqual, 1)
}

func TestMarshalUnmarshalTextDevAddrPrefix(t *testing.T) {
	a := New(t)
	var prefix DevAddrPrefix

	txt, err := prefix.MarshalText()
	a.So(err, ShouldBeNil)
	a.So(string(txt), ShouldEqual, "00000000/0")

	err = prefix.UnmarshalText([]byte("ff556677/1"))
	a.So(err, ShouldBeNil)
	a.So(prefix.DevAddr, ShouldEqual, DevAddr{128, 0, 0, 0})
	a.So(prefix.Length, ShouldEqual, 1)

	txt, err = prefix.MarshalText()
	a.So(err, ShouldBeNil)
	a.So(string(txt), ShouldEqual, "80000000/1")
}

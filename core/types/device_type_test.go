package types

import (
	"testing"

	. "github.com/smartystreets/assertions"
)

func TestDeviceType(t *testing.T) {
	a := New(t)

	// Setup
	abp := ABP
	otaa := OTAA
	abpStr := "ABP"
	otaaStr := "OTAA"
	abpBin := []byte{0x00}
	otaaBin := []byte{0x01}

	// String
	a.So(abp.String(), ShouldEqual, abpStr)
	a.So(otaa.String(), ShouldEqual, otaaStr)

	// MarshalText
	mtOut, err := abp.MarshalText()
	a.So(err, ShouldBeNil)
	a.So(mtOut, ShouldResemble, []byte(abpStr))
	mtOut, err = otaa.MarshalText()
	a.So(err, ShouldBeNil)
	a.So(mtOut, ShouldResemble, []byte(otaaStr))

	// MarshalBinary
	mbOut, err := abp.MarshalBinary()
	a.So(err, ShouldBeNil)
	a.So(mbOut, ShouldResemble, abpBin)
	mbOut, err = otaa.MarshalBinary()
	a.So(err, ShouldBeNil)
	a.So(mbOut, ShouldResemble, otaaBin)

	// Marshal
	mOut, err := abp.Marshal()
	a.So(err, ShouldBeNil)
	a.So(mOut, ShouldResemble, abpBin)
	mOut, err = otaa.Marshal()
	a.So(err, ShouldBeNil)
	a.So(mOut, ShouldResemble, otaaBin)

	// Parse
	pOut, err := ParseDeviceType(abpStr)
	a.So(err, ShouldBeNil)
	a.So(pOut, ShouldEqual, abp)
	pOut, err = ParseDeviceType(otaaStr)
	a.So(err, ShouldBeNil)
	a.So(pOut, ShouldEqual, otaa)

	// UnmarshalText
	utOut := ABP
	err = utOut.UnmarshalText([]byte(otaaStr))
	a.So(err, ShouldBeNil)
	a.So(utOut, ShouldEqual, otaa)
	err = utOut.UnmarshalText([]byte(abpStr))
	a.So(err, ShouldBeNil)
	a.So(utOut, ShouldEqual, abp)

	// UnmarshalBinary
	ubOut := ABP
	err = ubOut.UnmarshalBinary(otaaBin)
	a.So(err, ShouldBeNil)
	a.So(ubOut, ShouldEqual, otaa)
	err = ubOut.UnmarshalBinary(abpBin)
	a.So(err, ShouldBeNil)
	a.So(ubOut, ShouldEqual, abp)

	// Unmarshal
	uOut := ABP
	err = uOut.Unmarshal(otaaBin)
	a.So(err, ShouldBeNil)
	a.So(uOut, ShouldEqual, otaa)
	err = uOut.Unmarshal(abpBin)
	a.So(err, ShouldBeNil)
	a.So(uOut, ShouldEqual, abp)
}

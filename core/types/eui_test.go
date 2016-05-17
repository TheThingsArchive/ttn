package types

import (
	"testing"

	. "github.com/smartystreets/assertions"
)

func TestEUI64(t *testing.T) {
	a := New(t)

	// Setup
	eui := EUI64{1, 2, 3, 4, 252, 253, 254, 255}
	str := "01020304FCFDFEFF"
	bin := []byte{0x01, 0x02, 0x03, 0x04, 0xfc, 0xfd, 0xfe, 0xff}

	// Bytes
	a.So(eui.Bytes(), ShouldResemble, bin)

	// String
	a.So(eui.String(), ShouldEqual, str)

	// MarshalText
	mtOut, err := eui.MarshalText()
	a.So(err, ShouldBeNil)
	a.So(mtOut, ShouldResemble, []byte(str))

	// MarshalBinary
	mbOut, err := eui.MarshalBinary()
	a.So(err, ShouldBeNil)
	a.So(mbOut, ShouldResemble, bin)

	// Marshal
	mOut, err := eui.Marshal()
	a.So(err, ShouldBeNil)
	a.So(mOut, ShouldResemble, bin)

	// Parse
	pOut, err := ParseEUI64(str)
	a.So(err, ShouldBeNil)
	a.So(pOut, ShouldResemble, eui)

	// UnmarshalText
	utOut := &EUI64{}
	err = utOut.UnmarshalText([]byte(str))
	a.So(err, ShouldBeNil)
	a.So(utOut, ShouldResemble, &eui)

	// UnmarshalBinary
	ubOut := &EUI64{}
	err = ubOut.UnmarshalBinary(bin)
	a.So(err, ShouldBeNil)
	a.So(ubOut, ShouldResemble, &eui)

	// Unmarshal
	uOut := &EUI64{}
	err = uOut.Unmarshal(bin)
	a.So(err, ShouldBeNil)
	a.So(uOut, ShouldResemble, &eui)

	// IsEmpty
	var empty EUI64
	a.So(empty.IsEmpty(), ShouldEqual, true)
	a.So(eui.IsEmpty(), ShouldEqual, false)
}

func TestAppEUI(t *testing.T) {
	a := New(t)

	// Setup
	eui := AppEUI{1, 2, 3, 4, 252, 253, 254, 255}
	str := "01020304FCFDFEFF"
	bin := []byte{0x01, 0x02, 0x03, 0x04, 0xfc, 0xfd, 0xfe, 0xff}

	// Bytes
	a.So(eui.Bytes(), ShouldResemble, bin)

	// String
	a.So(eui.String(), ShouldEqual, str)

	// MarshalText
	mtOut, err := eui.MarshalText()
	a.So(err, ShouldBeNil)
	a.So(mtOut, ShouldResemble, []byte(str))

	// MarshalBinary
	mbOut, err := eui.MarshalBinary()
	a.So(err, ShouldBeNil)
	a.So(mbOut, ShouldResemble, bin)

	// Marshal
	mOut, err := eui.Marshal()
	a.So(err, ShouldBeNil)
	a.So(mOut, ShouldResemble, bin)

	// Parse
	pOut, err := ParseAppEUI(str)
	a.So(err, ShouldBeNil)
	a.So(pOut, ShouldEqual, eui)

	// UnmarshalText
	utOut := &AppEUI{}
	err = utOut.UnmarshalText([]byte(str))
	a.So(err, ShouldBeNil)
	a.So(*utOut, ShouldEqual, eui)

	// UnmarshalBinary
	ubOut := &AppEUI{}
	err = ubOut.UnmarshalBinary(bin)
	a.So(err, ShouldBeNil)
	a.So(*ubOut, ShouldEqual, eui)

	// Unmarshal
	uOut := &AppEUI{}
	err = uOut.Unmarshal(bin)
	a.So(err, ShouldBeNil)
	a.So(*uOut, ShouldEqual, eui)

	// IsEmpty
	var empty AppEUI
	a.So(empty.IsEmpty(), ShouldEqual, true)
	a.So(eui.IsEmpty(), ShouldEqual, false)
}

func TestDevEUI(t *testing.T) {
	a := New(t)

	// Setup
	eui := DevEUI{1, 2, 3, 4, 252, 253, 254, 255}
	str := "01020304FCFDFEFF"
	bin := []byte{0x01, 0x02, 0x03, 0x04, 0xfc, 0xfd, 0xfe, 0xff}

	// Bytes
	a.So(eui.Bytes(), ShouldResemble, bin)

	// String
	a.So(eui.String(), ShouldEqual, str)

	// MarshalText
	mtOut, err := eui.MarshalText()
	a.So(err, ShouldBeNil)
	a.So(mtOut, ShouldResemble, []byte(str))

	// MarshalBinary
	mbOut, err := eui.MarshalBinary()
	a.So(err, ShouldBeNil)
	a.So(mbOut, ShouldResemble, bin)

	// Marshal
	mOut, err := eui.Marshal()
	a.So(err, ShouldBeNil)
	a.So(mOut, ShouldResemble, bin)

	// Parse
	pOut, err := ParseDevEUI(str)
	a.So(err, ShouldBeNil)
	a.So(pOut, ShouldEqual, eui)

	// UnmarshalText
	utOut := &DevEUI{}
	err = utOut.UnmarshalText([]byte(str))
	a.So(err, ShouldBeNil)
	a.So(*utOut, ShouldEqual, eui)

	// UnmarshalBinary
	ubOut := &DevEUI{}
	err = ubOut.UnmarshalBinary(bin)
	a.So(err, ShouldBeNil)
	a.So(*ubOut, ShouldEqual, eui)

	// Unmarshal
	uOut := &DevEUI{}
	err = uOut.Unmarshal(bin)
	a.So(err, ShouldBeNil)
	a.So(*uOut, ShouldEqual, eui)

	// IsEmpty
	var empty DevEUI
	a.So(empty.IsEmpty(), ShouldEqual, true)
	a.So(eui.IsEmpty(), ShouldEqual, false)
}

func TestGatewayEUI(t *testing.T) {
	a := New(t)

	// Setup
	eui := GatewayEUI{1, 2, 3, 4, 252, 253, 254, 255}
	str := "01020304FCFDFEFF"
	bin := []byte{0x01, 0x02, 0x03, 0x04, 0xfc, 0xfd, 0xfe, 0xff}

	// Bytes
	a.So(eui.Bytes(), ShouldResemble, bin)

	// String
	a.So(eui.String(), ShouldEqual, str)

	// MarshalText
	mtOut, err := eui.MarshalText()
	a.So(err, ShouldBeNil)
	a.So(mtOut, ShouldResemble, []byte(str))

	// MarshalBinary
	mbOut, err := eui.MarshalBinary()
	a.So(err, ShouldBeNil)
	a.So(mbOut, ShouldResemble, bin)

	// Marshal
	mOut, err := eui.Marshal()
	a.So(err, ShouldBeNil)
	a.So(mOut, ShouldResemble, bin)

	// Parse
	pOut, err := ParseGatewayEUI(str)
	a.So(err, ShouldBeNil)
	a.So(pOut, ShouldEqual, eui)

	// UnmarshalText
	utOut := &GatewayEUI{}
	err = utOut.UnmarshalText([]byte(str))
	a.So(err, ShouldBeNil)
	a.So(*utOut, ShouldEqual, eui)

	// UnmarshalBinary
	ubOut := &GatewayEUI{}
	err = ubOut.UnmarshalBinary(bin)
	a.So(err, ShouldBeNil)
	a.So(*ubOut, ShouldEqual, eui)

	// Unmarshal
	uOut := &GatewayEUI{}
	err = uOut.Unmarshal(bin)
	a.So(err, ShouldBeNil)
	a.So(*uOut, ShouldEqual, eui)

	// IsEmpty
	var empty GatewayEUI
	a.So(empty.IsEmpty(), ShouldEqual, true)
	a.So(eui.IsEmpty(), ShouldEqual, false)
}

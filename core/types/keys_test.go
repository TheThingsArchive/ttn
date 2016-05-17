package types

import (
	"testing"

	. "github.com/smartystreets/assertions"
)

func TestAES128Key(t *testing.T) {
	a := New(t)

	// Setup
	key := AES128Key{1, 2, 3, 4, 5, 6, 7, 8, 249, 250, 251, 252, 253, 254, 255, 0}
	str := "0102030405060708F9FAFBFCFDFEFF00"
	bin := []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0xf9, 0xfa, 0xfb, 0xfc, 0xfd, 0xfe, 0xff, 0x00}

	// Bytes
	a.So(key.Bytes(), ShouldResemble, bin)

	// String
	a.So(key.String(), ShouldEqual, str)

	// MarshalText
	mtOut, err := key.MarshalText()
	a.So(err, ShouldBeNil)
	a.So(mtOut, ShouldResemble, []byte(str))

	// MarshalBinary
	mbOut, err := key.MarshalBinary()
	a.So(err, ShouldBeNil)
	a.So(mbOut, ShouldResemble, bin)

	// Marshal
	mOut, err := key.Marshal()
	a.So(err, ShouldBeNil)
	a.So(mOut, ShouldResemble, bin)

	// Parse
	pOut, err := ParseAES128Key(str)
	a.So(err, ShouldBeNil)
	a.So(pOut, ShouldEqual, key)

	// UnmarshalText
	utOut := &AES128Key{}
	err = utOut.UnmarshalText([]byte(str))
	a.So(err, ShouldBeNil)
	a.So(*utOut, ShouldEqual, key)

	// UnmarshalBinary
	ubOut := &AES128Key{}
	err = ubOut.UnmarshalBinary(bin)
	a.So(err, ShouldBeNil)
	a.So(*ubOut, ShouldEqual, key)

	// Unmarshal
	uOut := &AES128Key{}
	err = uOut.Unmarshal(bin)
	a.So(err, ShouldBeNil)
	a.So(*uOut, ShouldEqual, key)

	// IsEmpty
	var empty AES128Key
	a.So(empty.IsEmpty(), ShouldBeTrue)
	a.So(key.IsEmpty(), ShouldBeFalse)
}

func TestAppKey(t *testing.T) {
	a := New(t)

	// Setup
	key := AppKey{1, 2, 3, 4, 5, 6, 7, 8, 249, 250, 251, 252, 253, 254, 255, 0}
	str := "0102030405060708F9FAFBFCFDFEFF00"
	bin := []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0xf9, 0xfa, 0xfb, 0xfc, 0xfd, 0xfe, 0xff, 0x00}

	// Bytes
	a.So(key.Bytes(), ShouldResemble, bin)

	// String
	a.So(key.String(), ShouldEqual, str)

	// MarshalText
	mtOut, err := key.MarshalText()
	a.So(err, ShouldBeNil)
	a.So(mtOut, ShouldResemble, []byte(str))

	// MarshalBinary
	mbOut, err := key.MarshalBinary()
	a.So(err, ShouldBeNil)
	a.So(mbOut, ShouldResemble, bin)

	// Marshal
	mOut, err := key.Marshal()
	a.So(err, ShouldBeNil)
	a.So(mOut, ShouldResemble, bin)

	// Parse
	pOut, err := ParseAppKey(str)
	a.So(err, ShouldBeNil)
	a.So(pOut, ShouldEqual, key)

	// UnmarshalText
	utOut := &AppKey{}
	err = utOut.UnmarshalText([]byte(str))
	a.So(err, ShouldBeNil)
	a.So(*utOut, ShouldEqual, key)

	// UnmarshalBinary
	ubOut := &AppKey{}
	err = ubOut.UnmarshalBinary(bin)
	a.So(err, ShouldBeNil)
	a.So(*ubOut, ShouldEqual, key)

	// Unmarshal
	uOut := &AppKey{}
	err = uOut.Unmarshal(bin)
	a.So(err, ShouldBeNil)
	a.So(*uOut, ShouldEqual, key)

	// IsEmpty
	var empty AppKey
	a.So(empty.IsEmpty(), ShouldBeTrue)
	a.So(key.IsEmpty(), ShouldBeFalse)
}

func TestNwkSKey(t *testing.T) {
	a := New(t)

	// Setup
	key := NwkSKey{1, 2, 3, 4, 5, 6, 7, 8, 249, 250, 251, 252, 253, 254, 255, 0}
	str := "0102030405060708F9FAFBFCFDFEFF00"
	bin := []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0xf9, 0xfa, 0xfb, 0xfc, 0xfd, 0xfe, 0xff, 0x00}

	// Bytes
	a.So(key.Bytes(), ShouldResemble, bin)

	// String
	a.So(key.String(), ShouldEqual, str)

	// MarshalText
	mtOut, err := key.MarshalText()
	a.So(err, ShouldBeNil)
	a.So(mtOut, ShouldResemble, []byte(str))

	// MarshalBinary
	mbOut, err := key.MarshalBinary()
	a.So(err, ShouldBeNil)
	a.So(mbOut, ShouldResemble, bin)

	// Marshal
	mOut, err := key.Marshal()
	a.So(err, ShouldBeNil)
	a.So(mOut, ShouldResemble, bin)

	// Parse
	pOut, err := ParseNwkSKey(str)
	a.So(err, ShouldBeNil)
	a.So(pOut, ShouldEqual, key)

	// UnmarshalText
	utOut := &NwkSKey{}
	err = utOut.UnmarshalText([]byte(str))
	a.So(err, ShouldBeNil)
	a.So(*utOut, ShouldEqual, key)

	// UnmarshalBinary
	ubOut := &NwkSKey{}
	err = ubOut.UnmarshalBinary(bin)
	a.So(err, ShouldBeNil)
	a.So(*ubOut, ShouldEqual, key)

	// Unmarshal
	uOut := &NwkSKey{}
	err = uOut.Unmarshal(bin)
	a.So(err, ShouldBeNil)
	a.So(*uOut, ShouldEqual, key)

	// IsEmpty
	var empty NwkSKey
	a.So(empty.IsEmpty(), ShouldBeTrue)
	a.So(key.IsEmpty(), ShouldBeFalse)
}

func TestAppSKey(t *testing.T) {
	a := New(t)

	// Setup
	key := AppSKey{1, 2, 3, 4, 5, 6, 7, 8, 249, 250, 251, 252, 253, 254, 255, 0}
	str := "0102030405060708F9FAFBFCFDFEFF00"
	bin := []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0xf9, 0xfa, 0xfb, 0xfc, 0xfd, 0xfe, 0xff, 0x00}

	// Bytes
	a.So(key.Bytes(), ShouldResemble, bin)

	// String
	a.So(key.String(), ShouldEqual, str)

	// MarshalText
	mtOut, err := key.MarshalText()
	a.So(err, ShouldBeNil)
	a.So(mtOut, ShouldResemble, []byte(str))

	// MarshalBinary
	mbOut, err := key.MarshalBinary()
	a.So(err, ShouldBeNil)
	a.So(mbOut, ShouldResemble, bin)

	// Marshal
	mOut, err := key.Marshal()
	a.So(err, ShouldBeNil)
	a.So(mOut, ShouldResemble, bin)

	// Parse
	pOut, err := ParseAppSKey(str)
	a.So(err, ShouldBeNil)
	a.So(pOut, ShouldEqual, key)

	// UnmarshalText
	utOut := &AppSKey{}
	err = utOut.UnmarshalText([]byte(str))
	a.So(err, ShouldBeNil)
	a.So(*utOut, ShouldEqual, key)

	// UnmarshalBinary
	ubOut := &AppSKey{}
	err = ubOut.UnmarshalBinary(bin)
	a.So(err, ShouldBeNil)
	a.So(*ubOut, ShouldEqual, key)

	// Unmarshal
	uOut := &AppSKey{}
	err = uOut.Unmarshal(bin)
	a.So(err, ShouldBeNil)
	a.So(*uOut, ShouldEqual, key)

	// IsEmpty
	var empty AppSKey
	a.So(empty.IsEmpty(), ShouldBeTrue)
	a.So(key.IsEmpty(), ShouldBeFalse)
}

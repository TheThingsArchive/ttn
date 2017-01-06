// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package types

import (
	"testing"

	. "github.com/smartystreets/assertions"
)

func TestDevNonce(t *testing.T) {
	a := New(t)

	// Setup
	nonce := DevNonce{1, 2}
	str := "0102"
	bin := []byte{0x01, 0x02}

	// Bytes
	a.So(nonce.Bytes(), ShouldResemble, bin)

	// String
	a.So(nonce.String(), ShouldEqual, str)

	// MarshalText
	mtOut, err := nonce.MarshalText()
	a.So(err, ShouldBeNil)
	a.So(mtOut, ShouldResemble, []byte(str))

	// MarshalBinary
	mbOut, err := nonce.MarshalBinary()
	a.So(err, ShouldBeNil)
	a.So(mbOut, ShouldResemble, bin)

	// Marshal
	mOut, err := nonce.Marshal()
	a.So(err, ShouldBeNil)
	a.So(mOut, ShouldResemble, bin)

	// MarshalTo
	bOut := make([]byte, 2)
	_, err = nonce.MarshalTo(bOut)
	a.So(err, ShouldBeNil)
	a.So(bOut, ShouldResemble, bin)

	// Size
	s := nonce.Size()
	a.So(s, ShouldEqual, 2)

	// UnmarshalText
	utOut := &DevNonce{}
	err = utOut.UnmarshalText([]byte(str))
	a.So(err, ShouldBeNil)
	a.So(utOut, ShouldResemble, &nonce)

	// UnmarshalBinary
	ubOut := &DevNonce{}
	err = ubOut.UnmarshalBinary(bin)
	a.So(err, ShouldBeNil)
	a.So(ubOut, ShouldResemble, &nonce)

	// Unmarshal
	uOut := &DevNonce{}
	err = uOut.Unmarshal(bin)
	a.So(err, ShouldBeNil)
	a.So(uOut, ShouldResemble, &nonce)
}

func TestAppNonce(t *testing.T) {
	a := New(t)

	// Setup
	nonce := AppNonce{1, 2, 3}
	str := "010203"
	bin := []byte{0x01, 0x02, 0x03}

	// Bytes
	a.So(nonce.Bytes(), ShouldResemble, bin)

	// String
	a.So(nonce.String(), ShouldEqual, str)

	// MarshalText
	mtOut, err := nonce.MarshalText()
	a.So(err, ShouldBeNil)
	a.So(mtOut, ShouldResemble, []byte(str))

	// MarshalBinary
	mbOut, err := nonce.MarshalBinary()
	a.So(err, ShouldBeNil)
	a.So(mbOut, ShouldResemble, bin)

	// Marshal
	mOut, err := nonce.Marshal()
	a.So(err, ShouldBeNil)
	a.So(mOut, ShouldResemble, bin)

	// MarshalTo
	bOut := make([]byte, 3)
	_, err = nonce.MarshalTo(bOut)
	a.So(err, ShouldBeNil)
	a.So(bOut, ShouldResemble, bin)

	// Size
	s := nonce.Size()
	a.So(s, ShouldEqual, 3)

	// UnmarshalText
	utOut := &AppNonce{}
	err = utOut.UnmarshalText([]byte(str))
	a.So(err, ShouldBeNil)
	a.So(utOut, ShouldResemble, &nonce)

	// UnmarshalBinary
	ubOut := &AppNonce{}
	err = ubOut.UnmarshalBinary(bin)
	a.So(err, ShouldBeNil)
	a.So(ubOut, ShouldResemble, &nonce)

	// Unmarshal
	uOut := &AppNonce{}
	err = uOut.Unmarshal(bin)
	a.So(err, ShouldBeNil)
	a.So(uOut, ShouldResemble, &nonce)
}

func TestNetID(t *testing.T) {
	a := New(t)

	// Setup
	nid := NetID{1, 2, 3}
	str := "010203"
	bin := []byte{0x01, 0x02, 0x03}

	// Bytes
	a.So(nid.Bytes(), ShouldResemble, bin)

	// String
	a.So(nid.String(), ShouldEqual, str)

	// MarshalText
	mtOut, err := nid.MarshalText()
	a.So(err, ShouldBeNil)
	a.So(mtOut, ShouldResemble, []byte(str))

	// MarshalBinary
	mbOut, err := nid.MarshalBinary()
	a.So(err, ShouldBeNil)
	a.So(mbOut, ShouldResemble, bin)

	// Marshal
	mOut, err := nid.Marshal()
	a.So(err, ShouldBeNil)
	a.So(mOut, ShouldResemble, bin)

	// MarshalTo
	bOut := make([]byte, 3)
	_, err = nid.MarshalTo(bOut)
	a.So(err, ShouldBeNil)
	a.So(bOut, ShouldResemble, bin)

	// Size
	s := nid.Size()
	a.So(s, ShouldEqual, 3)

	// UnmarshalText
	utOut := &NetID{}
	err = utOut.UnmarshalText([]byte(str))
	a.So(err, ShouldBeNil)
	a.So(utOut, ShouldResemble, &nid)

	// UnmarshalBinary
	ubOut := &NetID{}
	err = ubOut.UnmarshalBinary(bin)
	a.So(err, ShouldBeNil)
	a.So(ubOut, ShouldResemble, &nid)

	// Unmarshal
	uOut := &NetID{}
	err = uOut.Unmarshal(bin)
	a.So(err, ShouldBeNil)
	a.So(uOut, ShouldResemble, &nid)
}

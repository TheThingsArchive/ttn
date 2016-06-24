// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package types

import (
	"testing"

	. "github.com/smartystreets/assertions"
)

func TestDataRate(t *testing.T) {
	a := New(t)

	// Setup
	datr := DataRate{7, 125}
	str := "SF7BW125"
	bin := []byte("SF7BW125")

	// Bytes
	a.So(datr.Bytes(), ShouldResemble, bin)

	// String
	a.So(datr.String(), ShouldEqual, str)

	// MarshalText
	mtOut, err := datr.MarshalText()
	a.So(err, ShouldBeNil)
	a.So(mtOut, ShouldResemble, []byte(str))

	// MarshalBinary
	mbOut, err := datr.MarshalBinary()
	a.So(err, ShouldBeNil)
	a.So(mbOut, ShouldResemble, bin)

	// Marshal
	mOut, err := datr.Marshal()
	a.So(err, ShouldBeNil)
	a.So(mOut, ShouldResemble, bin)

	// Parse
	pOut, err := ParseDataRate(str)
	a.So(err, ShouldBeNil)
	a.So(*pOut, ShouldResemble, datr)

	// UnmarshalText
	utOut := &DataRate{}
	err = utOut.UnmarshalText([]byte(str))
	a.So(err, ShouldBeNil)
	a.So(*utOut, ShouldResemble, datr)

	// UnmarshalBinary
	ubOut := &DataRate{}
	err = ubOut.UnmarshalBinary(bin)
	a.So(err, ShouldBeNil)
	a.So(*ubOut, ShouldResemble, datr)

	// Unmarshal
	uOut := &DataRate{}
	err = uOut.Unmarshal(bin)
	a.So(err, ShouldBeNil)
	a.So(*uOut, ShouldResemble, datr)
}

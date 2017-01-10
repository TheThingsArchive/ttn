// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package types

import (
	"testing"

	. "github.com/smartystreets/assertions"
)

func TestParseHex(t *testing.T) {
	a := New(t)
	b, err := ParseHEX("AABC", 2)
	a.So(err, ShouldBeNil)
	a.So(b, ShouldResemble, []byte{0xaa, 0xbc})

	b, err = ParseHEX("", 2)
	a.So(err, ShouldBeNil)
	a.So(b, ShouldResemble, []byte{0x00, 0x00})

	_, err = ParseHEX("ab", 2)
	a.So(err, ShouldNotBeNil)
}

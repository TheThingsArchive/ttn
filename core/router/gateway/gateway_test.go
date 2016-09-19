// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package gateway

import (
	"testing"

	. "github.com/TheThingsNetwork/ttn/utils/testing"
	. "github.com/smartystreets/assertions"
)

func TestNewGateway(t *testing.T) {
	a := New(t)
	gtw := NewGateway(GetLogger(t, "TestNewGateway"), "eui-0102030405060708")
	a.So(gtw, ShouldNotBeNil)
}

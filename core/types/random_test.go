// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package types_test

import (
	"math/rand"
	"testing"
	"time"

	. "github.com/TheThingsNetwork/ttn/core/types"
	. "github.com/smartystreets/assertions"
)

var random Rand = rand.New(rand.NewSource(time.Now().UnixNano()))

func TestNewPopulatedDevAddr(t *testing.T) {
	a := New(t)
	var devAddr *DevAddr
	a.So(func() {
		devAddr = NewPopulatedDevAddr(random)
	}, ShouldNotPanic)
	a.So(devAddr.IsEmpty(), ShouldBeFalse)
	a.So(devAddr.Size(), ShouldEqual, 4)
}

func TestNewPopulatedAppEUI(t *testing.T) {
	a := New(t)
	var appEUI *AppEUI
	a.So(func() {
		appEUI = NewPopulatedAppEUI(random)
	}, ShouldNotPanic)
	a.So(appEUI.IsEmpty(), ShouldBeFalse)
	a.So(appEUI.Size(), ShouldEqual, 8)
}

func TestNewPopulatedDevEUI(t *testing.T) {
	a := New(t)
	var devEUI *DevEUI
	a.So(func() {
		devEUI = NewPopulatedDevEUI(random)
	}, ShouldNotPanic)
	a.So(devEUI.IsEmpty(), ShouldBeFalse)
	a.So(devEUI.Size(), ShouldEqual, 8)
}

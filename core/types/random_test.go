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

func TestNewPopulatedAppKey(t *testing.T) {
	a := New(t)
	var appKey *AppKey
	a.So(func() {
		appKey = NewPopulatedAppKey(random)
	}, ShouldNotPanic)
	a.So(appKey.IsEmpty(), ShouldBeFalse)
	a.So(appKey.Size(), ShouldEqual, 16)
}

func TestNewPopulatedAppSKey(t *testing.T) {
	a := New(t)
	var appSKey *AppSKey
	a.So(func() {
		appSKey = NewPopulatedAppSKey(random)
	}, ShouldNotPanic)
	a.So(appSKey.IsEmpty(), ShouldBeFalse)
	a.So(appSKey.Size(), ShouldEqual, 16)
}

func TestNewPopulatedNwkSKey(t *testing.T) {
	a := New(t)
	var nwkSKey *NwkSKey
	a.So(func() {
		nwkSKey = NewPopulatedNwkSKey(random)
	}, ShouldNotPanic)
	a.So(nwkSKey.IsEmpty(), ShouldBeFalse)
	a.So(nwkSKey.Size(), ShouldEqual, 16)
}

func TestNewPopulatedDevNonce(t *testing.T) {
	a := New(t)
	var devNonce *DevNonce
	a.So(func() {
		devNonce = NewPopulatedDevNonce(random)
	}, ShouldNotPanic)
	a.So(devNonce.Size(), ShouldEqual, 2)
}

func TestNewPopulatedAppNonce(t *testing.T) {
	a := New(t)
	var appNonce *AppNonce
	a.So(func() {
		appNonce = NewPopulatedAppNonce(random)
	}, ShouldNotPanic)
	a.So(appNonce.Size(), ShouldEqual, 3)
}

func TestNewPopulatedNetID(t *testing.T) {
	a := New(t)
	var netID *NetID
	a.So(func() {
		netID = NewPopulatedNetID(random)
	}, ShouldNotPanic)
	a.So(netID.Size(), ShouldEqual, 3)
}

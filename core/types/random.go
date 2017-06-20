// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package types

import (
	"fmt"
)

// A Rand is a source of random int64 numbers
type Rand interface {
	Int63() int64
}

func randRead(r Rand, b []byte) (n int, err error) {
	// Adapted from go stdlib https://goo.gl/i3vwnE
	pos := 7
	val := r.Int63()
	for n := range b {
		if pos == 0 {
			val = r.Int63()
			pos = 7
		}
		b[n] = byte(val)
		val >>= 8
		pos--
	}
	return
}

// NewPopulatedDevAddr returns random DevAddr
func NewPopulatedDevAddr(r Rand) (devAddr *DevAddr) {
	devAddr = &DevAddr{}
	if _, err := randRead(r, devAddr[:]); err != nil {
		panic(fmt.Errorf("types.NewPopulatedDevAddr: %s", err))
	}
	return
}

// NewPopulatedAppEUI returns random AppEUI
func NewPopulatedAppEUI(r Rand) (appEUI *AppEUI) {
	appEUI = &AppEUI{}
	if _, err := randRead(r, appEUI[:]); err != nil {
		panic(fmt.Errorf("types.NewPopulatedAppEUI: %s", err))
	}
	return
}

// NewPopulatedDevEUI returns random DevEUI
func NewPopulatedDevEUI(r Rand) (devEUI *DevEUI) {
	devEUI = &DevEUI{}
	if _, err := randRead(r, devEUI[:]); err != nil {
		panic(fmt.Errorf("types.NewPopulatedDevEUI: %s", err))
	}
	return
}

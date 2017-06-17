// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package types

import (
	"fmt"
)

type Rand interface {
	Read(b []byte) (n int, err error)
}

func NewPopulatedDevAddr(r Rand) (devAddr DevAddr) {
	if _, err := r.Read(devAddr[:]); err != nil {
		panic(fmt.Errorf("types.NewPopulatedAppEUI: %s", err))
	}
	return
}

func NewPopulatedAppEUI(r Rand) (appEUI AppEUI) {
	if _, err := r.Read(appEUI[:]); err != nil {
		panic(fmt.Errorf("types.NewPopulatedAppEUI: %s", err))
	}
	return
}

func NewPopulatedDevEUI(r Rand) (devEUI DevEUI) {
	if _, err := r.Read(devEUI[:]); err != nil {
		panic(fmt.Errorf("types.NewPopulatedDevEUI: %s", err))
	}
	return
}

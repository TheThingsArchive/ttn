// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package random

import (
	"testing"

	"github.com/TheThingsNetwork/ttn/api"
	s "github.com/smartystreets/assertions"
)

func TestRssi(t *testing.T) {
	s.New(t).So(Rssi(), s.ShouldBeBetween, -120, 0)
}

func TestValidID(t *testing.T) {
	for name, id := range map[string]string{
		"ID":    ID(),
		"AppID": AppID(),
		"DevID": DevID(),
	} {
		t.Run(name, func(t *testing.T) {
			s.New(t).So(api.ValidID(id), s.ShouldBeTrue)
		})
	}
}

func TestEmptiers(t *testing.T) {
	type isEmptier interface {
		IsEmpty() bool
	}

	for name, eui := range map[string]isEmptier{
		"EUI64":   EUI64(),
		"DevEUI":  DevEUI(),
		"AppEUI":  AppEUI(),
		"NetID":   NetID(),
		"DevAddr": DevAddr(),
	} {
		t.Run(name, func(t *testing.T) {
			s.New(t).So(eui.IsEmpty(), s.ShouldBeFalse)
		})
	}
}

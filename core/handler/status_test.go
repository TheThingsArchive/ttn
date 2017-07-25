// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package handler

import (
	"testing"

	pb "github.com/TheThingsNetwork/api/handler"
	. "github.com/smartystreets/assertions"
)

func TestStatus(t *testing.T) {
	a := New(t)
	h := new(handler)
	a.So(h.GetStatus(), ShouldResemble, new(pb.Status))
	h.InitStatus()
	a.So(h.status, ShouldNotBeNil)
	status := h.GetStatus()
	a.So(status.Uplink.Rate1, ShouldEqual, 0)
}

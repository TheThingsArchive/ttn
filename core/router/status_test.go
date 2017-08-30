// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package router

import (
	"testing"

	pb "github.com/TheThingsNetwork/api/router"
	. "github.com/smartystreets/assertions"
)

func TestStatus(t *testing.T) {
	a := New(t)
	r := new(router)
	a.So(r.GetStatus(), ShouldResemble, new(pb.Status))
	r.InitStatus()
	a.So(r.status, ShouldNotBeNil)
	status := r.GetStatus()
	a.So(status.Uplink.Rate1, ShouldEqual, 0)
}

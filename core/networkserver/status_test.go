// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package networkserver

import (
	"testing"

	pb "github.com/TheThingsNetwork/api/networkserver"
	. "github.com/smartystreets/assertions"
)

func TestStatus(t *testing.T) {
	a := New(t)
	ns := new(networkServer)
	a.So(ns.GetStatus(), ShouldResemble, new(pb.Status))
	ns.InitStatus()
	a.So(ns.status, ShouldNotBeNil)
	status := ns.GetStatus()
	a.So(status.Uplink.Rate1, ShouldEqual, 0)
}

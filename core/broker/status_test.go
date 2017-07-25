// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package broker

import (
	"testing"

	pb "github.com/TheThingsNetwork/api/broker"
	. "github.com/smartystreets/assertions"
)

func TestStatus(t *testing.T) {
	a := New(t)
	b := new(broker)
	a.So(b.GetStatus(), ShouldResemble, new(pb.Status))
	b.InitStatus()
	a.So(b.status, ShouldNotBeNil)
	status := b.GetStatus()
	a.So(status.Uplink.Rate1, ShouldEqual, 0)
}

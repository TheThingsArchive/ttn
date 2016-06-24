// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package broker

import (
	"testing"

	pb "github.com/TheThingsNetwork/ttn/api/broker"
	"github.com/TheThingsNetwork/ttn/core"
	"github.com/TheThingsNetwork/ttn/core/types"
	. "github.com/TheThingsNetwork/ttn/utils/testing"
	. "github.com/smartystreets/assertions"
)

func TestDownlink(t *testing.T) {
	a := New(t)

	appEUI := types.AppEUI{0, 1, 2, 3, 4, 5, 6, 7}
	devEUI := types.DevEUI{0, 1, 2, 3, 4, 5, 6, 7}

	dlch := make(chan *pb.DownlinkMessage, 2)

	b := &broker{
		Component: &core.Component{
			Ctx: GetLogger(t, "TestDownlink"),
		},
		ns: &mockNetworkServer{},
		routers: map[string]chan *pb.DownlinkMessage{
			"routerID": dlch,
		},
	}

	err := b.HandleDownlink(&pb.DownlinkMessage{
		DevEui: &devEUI,
		AppEui: &appEUI,
		DownlinkOption: &pb.DownlinkOption{
			Identifier: "fakeID",
		},
	})
	a.So(err, ShouldNotBeNil)

	err = b.HandleDownlink(&pb.DownlinkMessage{
		DevEui: &devEUI,
		AppEui: &appEUI,
		DownlinkOption: &pb.DownlinkOption{
			Identifier: "nonExistentRouterID:scheduleID",
		},
	})
	a.So(err, ShouldNotBeNil)

	err = b.HandleDownlink(&pb.DownlinkMessage{
		DevEui: &devEUI,
		AppEui: &appEUI,
		DownlinkOption: &pb.DownlinkOption{
			Identifier: "routerID:scheduleID",
		},
	})
	a.So(err, ShouldBeNil)
	a.So(len(dlch), ShouldEqual, 1)
}

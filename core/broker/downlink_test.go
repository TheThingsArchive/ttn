package broker

import (
	"testing"

	pb "github.com/TheThingsNetwork/ttn/api/broker"
	. "github.com/smartystreets/assertions"
)

func TestDownlink(t *testing.T) {
	a := New(t)

	dlch := make(chan *pb.DownlinkMessage, 2)

	b := &broker{
		ns: &mockNetworkServer{},
		routers: map[string]chan *pb.DownlinkMessage{
			"routerID": dlch,
		},
	}

	err := b.HandleDownlink(&pb.DownlinkMessage{
		DownlinkOption: &pb.DownlinkOption{
			Identifier: "fakeID",
		},
	})
	a.So(err, ShouldNotBeNil)

	err = b.HandleDownlink(&pb.DownlinkMessage{
		DownlinkOption: &pb.DownlinkOption{
			Identifier: "nonExistentRouterID:scheduleID",
		},
	})
	a.So(err, ShouldNotBeNil)

	err = b.HandleDownlink(&pb.DownlinkMessage{
		DownlinkOption: &pb.DownlinkOption{
			Identifier: "routerID:scheduleID",
		},
	})
	a.So(err, ShouldBeNil)
	a.So(len(dlch), ShouldEqual, 1)
}

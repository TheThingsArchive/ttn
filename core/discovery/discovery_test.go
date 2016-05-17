package discovery

import (
	"testing"

	pb "github.com/TheThingsNetwork/ttn/api/discovery"
	. "github.com/smartystreets/assertions"
)

func TestDiscoveryDiscover(t *testing.T) {
	a := New(t)

	router := &pb.Announcement{Id: "router", Token: "abcd"}
	broker1 := &pb.Announcement{Id: "broker1"}
	broker2 := &pb.Announcement{Id: "broker2"}

	d := &discovery{
		services: map[string]map[string]*pb.Announcement{
			"router": map[string]*pb.Announcement{
				"router": router,
			},
			"broker": map[string]*pb.Announcement{
				"broker1": broker1,
				"broker2": broker2,
			},
		},
	}

	_, err := d.Discover("random")
	a.So(err, ShouldNotBeNil)

	services, err := d.Discover("router")
	a.So(err, ShouldBeNil)
	a.So(services, ShouldHaveLength, 1)
	a.So(services[0].Id, ShouldEqual, router.Id)
	a.So(services[0].Token, ShouldBeEmpty)

	services, err = d.Discover("broker")
	a.So(err, ShouldBeNil)
	a.So(services, ShouldHaveLength, 2)
}

func TestDiscoveryAnnounce(t *testing.T) {
	a := New(t)

	broker1a := &pb.Announcement{ServiceName: "broker", Id: "broker1", Token: "abcd", NetAddress: "old address"}
	broker1aNoToken := &pb.Announcement{ServiceName: "broker", Id: "broker1", NetAddress: "new address"}
	broker1b := &pb.Announcement{ServiceName: "broker", Id: "broker1", Token: "abcd", NetAddress: "new address"}
	broker2 := &pb.Announcement{ServiceName: "broker", Id: "broker2", NetAddress: "other address"}

	d := &discovery{
		services: map[string]map[string]*pb.Announcement{},
	}

	err := d.Announce(broker1a)
	a.So(err, ShouldBeNil)

	err = d.Announce(broker1aNoToken)
	a.So(err, ShouldNotBeNil)

	services, err := d.Discover("broker")
	a.So(err, ShouldBeNil)
	a.So(services, ShouldHaveLength, 1)
	a.So(services[0].NetAddress, ShouldEqual, "old address")

	err = d.Announce(broker1b)
	a.So(err, ShouldBeNil)

	services, err = d.Discover("broker")
	a.So(err, ShouldBeNil)
	a.So(services, ShouldHaveLength, 1)
	a.So(services[0].NetAddress, ShouldEqual, "new address")

	err = d.Announce(broker2)
	a.So(err, ShouldBeNil)

	services, err = d.Discover("broker")
	a.So(err, ShouldBeNil)
	a.So(services, ShouldHaveLength, 2)

}

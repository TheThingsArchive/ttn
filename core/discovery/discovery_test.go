package discovery

import (
	"testing"

	pb "github.com/TheThingsNetwork/ttn/api/discovery"
	. "github.com/smartystreets/assertions"
)

func TestDiscoveryDiscover(t *testing.T) {
	a := New(t)

	router := &pb.Announcement{Token: "router"}
	broker1 := &pb.Announcement{Token: "broker1"}
	broker2 := &pb.Announcement{Token: "broker2"}

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
	a.So(services, ShouldContain, router)
	a.So(services, ShouldNotContain, broker1)

	services, err = d.Discover("broker")
	a.So(err, ShouldBeNil)
	a.So(services, ShouldContain, broker1)
	a.So(services, ShouldContain, broker2)
	a.So(services, ShouldNotContain, router)
}

func TestDiscoveryAnnounce(t *testing.T) {
	a := New(t)

	broker1a := &pb.Announcement{ServiceName: "broker", Token: "broker1", NetAddress: "old address"}
	broker1b := &pb.Announcement{ServiceName: "broker", Token: "broker1", NetAddress: "new address"}
	broker2 := &pb.Announcement{ServiceName: "broker", Token: "broker2", NetAddress: "other address"}

	d := &discovery{
		services: map[string]map[string]*pb.Announcement{},
	}

	err := d.Announce(broker1a)
	a.So(err, ShouldBeNil)

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

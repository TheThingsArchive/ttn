// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package discovery

import (
	"testing"

	"gopkg.in/redis.v3"

	pb "github.com/TheThingsNetwork/ttn/api/discovery"
	. "github.com/smartystreets/assertions"
)

func getRedisClient(db int64) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       db, // use default DB
	})
}

func TestDiscoveryAnnounce(t *testing.T) {
	a := New(t)

	localDiscovery := &discovery{
		services: map[string]map[string]*pb.Announcement{},
	}

	client := getRedisClient(1)
	redisDiscovery := NewRedisDiscovery(client)
	defer func() {
		client.Del("service:broker:broker1.1")
		client.Del("service:broker:broker1.2")
	}()

	discoveries := map[string]Discovery{
		"local": localDiscovery,
		"redis": redisDiscovery,
	}

	for name, d := range discoveries {
		broker1a := &pb.Announcement{ServiceName: "broker", Id: "broker1.1", Token: "abcd", NetAddress: "current address"}
		broker1aNoToken := &pb.Announcement{ServiceName: "broker", Id: "broker1.1", NetAddress: "attacker address"}
		broker1b := &pb.Announcement{ServiceName: "broker", Id: "broker1.1", Token: "abcd", NetAddress: "updated address"}
		broker2 := &pb.Announcement{ServiceName: "broker", Id: "broker1.2", NetAddress: "other address"}

		t.Logf("Testing %s\n", name)

		err := d.Announce(broker1a)
		a.So(err, ShouldBeNil)

		err = d.Announce(broker1aNoToken)
		a.So(err, ShouldNotBeNil)

		services, err := d.Discover("broker")
		a.So(err, ShouldBeNil)
		a.So(services, ShouldHaveLength, 1)
		a.So(services[0].NetAddress, ShouldEqual, "current address")

		err = d.Announce(broker1b)
		a.So(err, ShouldBeNil)

		services, err = d.Discover("broker")
		a.So(err, ShouldBeNil)
		a.So(services, ShouldHaveLength, 1)
		a.So(services[0].NetAddress, ShouldEqual, "updated address")

		err = d.Announce(broker2)
		a.So(err, ShouldBeNil)

		services, err = d.Discover("broker")
		a.So(err, ShouldBeNil)
		a.So(services, ShouldHaveLength, 2)
	}

}

func TestDiscoveryDiscover(t *testing.T) {
	a := New(t)

	router := &pb.Announcement{ServiceName: "router", Id: "router2.0", Token: "abcd"}
	broker1 := &pb.Announcement{ServiceName: "broker", Id: "broker2.1"}
	broker2 := &pb.Announcement{ServiceName: "broker", Id: "broker2.2"}

	localDiscovery := &discovery{
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

	client := getRedisClient(2)
	redisDiscovery := NewRedisDiscovery(client)
	defer func() {
		client.Del("service:router:router2.0")
		client.Del("service:broker:broker2.1")
		client.Del("service:broker:broker2.2")
	}()

	// This depends on the previous test to pass
	redisDiscovery.Announce(router)
	redisDiscovery.Announce(broker1)
	redisDiscovery.Announce(broker2)

	discoveries := map[string]Discovery{
		"local": localDiscovery,
		"redis": redisDiscovery,
	}

	for name, d := range discoveries {
		t.Logf("Testing %s\n", name)

		services, err := d.Discover("random")
		a.So(err, ShouldBeNil)
		a.So(services, ShouldBeEmpty)

		services, err = d.Discover("router")
		a.So(err, ShouldBeNil)
		a.So(services, ShouldHaveLength, 1)
		a.So(services[0].Id, ShouldEqual, router.Id)
		a.So(services[0].Token, ShouldBeEmpty)

		services, err = d.Discover("broker")
		a.So(err, ShouldBeNil)
		a.So(services, ShouldHaveLength, 2)

	}
}

// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package discovery

import (
	"fmt"
	"os"
	"testing"

	"gopkg.in/redis.v3"

	pb "github.com/TheThingsNetwork/ttn/api/discovery"
	"github.com/TheThingsNetwork/ttn/core/discovery/announcement"
	"github.com/TheThingsNetwork/ttn/core/discovery/kv"
	. "github.com/smartystreets/assertions"
)

func getRedisClient(db int64) *redis.Client {
	host := os.Getenv("REDIS_HOST")
	if host == "" {
		host = "localhost"
	}
	return redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:6379", host),
		Password: "", // no password set
		DB:       db,
	})
}

func TestDiscoveryAnnounce(t *testing.T) {
	a := New(t)

	localDiscovery := &discovery{
		services: announcement.NewAnnouncementStore(),
		appIDs:   kv.NewKVStore(),
	}

	client := getRedisClient(1)
	redisDiscovery := NewRedisDiscovery(client)
	defer func() {
		client.Del("discovery:announcement:broker:broker1.1")
		client.Del("discovery:announcement:broker:broker1.2")
	}()

	discoveries := map[string]Discovery{
		"local": localDiscovery,
		"redis": redisDiscovery,
	}

	for name, d := range discoveries {
		broker1a := &pb.Announcement{ServiceName: "broker", Id: "broker1.1", NetAddress: "current address"}
		broker1b := &pb.Announcement{ServiceName: "broker", Id: "broker1.1", NetAddress: "updated address"}
		broker2 := &pb.Announcement{ServiceName: "broker", Id: "broker1.2", NetAddress: "other address"}

		t.Logf("Testing %s\n", name)

		err := d.Announce(broker1a)
		a.So(err, ShouldBeNil)

		services, err := d.GetAll("broker")
		a.So(err, ShouldBeNil)
		a.So(services, ShouldHaveLength, 1)
		a.So(services[0].NetAddress, ShouldEqual, "current address")

		err = d.Announce(broker1b)
		a.So(err, ShouldBeNil)

		services, err = d.GetAll("broker")
		a.So(err, ShouldBeNil)
		a.So(services, ShouldHaveLength, 1)
		a.So(services[0].NetAddress, ShouldEqual, "updated address")

		err = d.Announce(broker2)
		a.So(err, ShouldBeNil)

		services, err = d.GetAll("broker")
		a.So(err, ShouldBeNil)
		a.So(services, ShouldHaveLength, 2)
	}

}

func TestDiscoveryDiscover(t *testing.T) {
	a := New(t)

	router := &pb.Announcement{ServiceName: "router", Id: "router2.0"}
	broker1 := &pb.Announcement{ServiceName: "broker", Id: "broker2.1"}
	broker2 := &pb.Announcement{ServiceName: "broker", Id: "broker2.2"}

	localDiscovery := &discovery{
		services: announcement.NewAnnouncementStore(),
		appIDs:   kv.NewKVStore(),
	}

	localDiscovery.services.Set(router)
	localDiscovery.services.Set(broker1)
	localDiscovery.services.Set(broker2)

	client := getRedisClient(2)
	redisDiscovery := NewRedisDiscovery(client)
	defer func() {
		client.Del("discovery:announcement:router:router2.0")
		client.Del("discovery:announcement:broker:broker2.1")
		client.Del("discovery:announcement:broker:broker2.2")
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

		services, err := d.GetAll("random")
		a.So(err, ShouldBeNil)
		a.So(services, ShouldBeEmpty)

		services, err = d.GetAll("router")
		a.So(err, ShouldBeNil)
		a.So(services, ShouldHaveLength, 1)
		a.So(services[0].Id, ShouldEqual, router.Id)

		services, err = d.GetAll("broker")
		a.So(err, ShouldBeNil)
		a.So(services, ShouldHaveLength, 2)

	}
}

func TestDiscoveryMetadata(t *testing.T) {
	a := New(t)

	localDiscovery := &discovery{
		services: announcement.NewAnnouncementStore(),
		appIDs:   kv.NewKVStore(),
	}

	client := getRedisClient(1)
	redisDiscovery := NewRedisDiscovery(client)
	defer func() {
		client.Del("discovery:announcement:broker:broker3")
		client.Del("discovery:announcement:broker:broker4")
		client.Del("discovery:app-id:app-id-2")
	}()

	discoveries := map[string]Discovery{
		"local": localDiscovery,
		"redis": redisDiscovery,
	}

	for name, d := range discoveries {
		broker3 := &pb.Announcement{ServiceName: "broker", Id: "broker3", Metadata: []*pb.Metadata{&pb.Metadata{
			Key:   pb.Metadata_APP_ID,
			Value: []byte("app-id-1"),
		}}}
		broker4 := &pb.Announcement{ServiceName: "broker", Id: "broker4", Metadata: []*pb.Metadata{&pb.Metadata{
			Key:   pb.Metadata_APP_ID,
			Value: []byte("app-id-2"),
		}}}

		t.Logf("Testing %s\n", name)

		// Announce should not change metadata
		err := d.Announce(broker3)
		a.So(err, ShouldBeNil)
		service, err := d.Get("broker", "broker3")
		a.So(err, ShouldBeNil)
		a.So(service.Metadata, ShouldHaveLength, 0)

		d.Announce(broker4)

		// AddMetadata should add one
		err = d.AddMetadata("broker", "broker3", &pb.Metadata{
			Key:   pb.Metadata_APP_ID,
			Value: []byte("app-id-2"),
		})
		a.So(err, ShouldBeNil)
		service, err = d.Get("broker", "broker3")
		a.So(err, ShouldBeNil)
		a.So(service.Metadata, ShouldHaveLength, 1)

		// And should remove it from the other broker
		service, err = d.Get("broker", "broker4")
		a.So(err, ShouldBeNil)
		a.So(service.Metadata, ShouldHaveLength, 0)

		// AddMetadata again should not add one
		err = d.AddMetadata("broker", "broker3", &pb.Metadata{
			Key:   pb.Metadata_APP_ID,
			Value: []byte("app-id-2"),
		})
		service, err = d.Get("broker", "broker3")
		a.So(err, ShouldBeNil)
		a.So(service.Metadata, ShouldHaveLength, 1)

		// DeleteMetadata for non-existing should not delete one
		err = d.DeleteMetadata("broker", "broker3", &pb.Metadata{
			Key:   pb.Metadata_APP_ID,
			Value: []byte("app-id-3"),
		})
		a.So(err, ShouldBeNil)
		service, err = d.Get("broker", "broker3")
		a.So(err, ShouldBeNil)
		a.So(service.Metadata, ShouldHaveLength, 1)

		// Announce should not change metadata
		err = d.Announce(broker3)
		a.So(err, ShouldBeNil)
		service, err = d.Get("broker", "broker3")
		a.So(err, ShouldBeNil)
		a.So(service.Metadata, ShouldHaveLength, 1)

		// DeleteMetadata should delete one
		err = d.DeleteMetadata("broker", "broker3", &pb.Metadata{
			Key:   pb.Metadata_APP_ID,
			Value: []byte("app-id-2"),
		})
		a.So(err, ShouldBeNil)
		service, err = d.Get("broker", "broker3")
		a.So(err, ShouldBeNil)
		a.So(service.Metadata, ShouldHaveLength, 0)

	}

}

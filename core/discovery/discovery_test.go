// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package discovery

import (
	"fmt"
	"os"
	"testing"

	"gopkg.in/redis.v5"

	pb "github.com/TheThingsNetwork/ttn/api/discovery"
	. "github.com/smartystreets/assertions"
)

func getRedisClient(db int) *redis.Client {
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

	client := getRedisClient(1)
	d := NewRedisDiscovery(client)
	defer func() {
		client.Del("discovery:announcement:broker:broker1.1")
		client.Del("discovery:announcement:broker:broker1.2")
	}()

	broker1a := &pb.Announcement{ServiceName: "broker", Id: "broker1.1", NetAddress: "current address"}
	broker1b := &pb.Announcement{ServiceName: "broker", Id: "broker1.1", NetAddress: "updated address"}
	broker2 := &pb.Announcement{ServiceName: "broker", Id: "broker1.2", NetAddress: "other address"}

	err := d.Announce(broker1a)
	a.So(err, ShouldBeNil)

	services, err := d.GetAll("broker", 0, 0)
	a.So(err, ShouldBeNil)
	a.So(services, ShouldHaveLength, 1)
	a.So(services[0].NetAddress, ShouldEqual, "current address")

	err = d.Announce(broker1b)
	a.So(err, ShouldBeNil)

	services, err = d.GetAll("broker", 0, 0)
	a.So(err, ShouldBeNil)
	a.So(services, ShouldHaveLength, 1)
	a.So(services[0].NetAddress, ShouldEqual, "updated address")

	err = d.Announce(broker2)
	a.So(err, ShouldBeNil)

	services, err = d.GetAll("broker", 0, 0)
	a.So(err, ShouldBeNil)
	a.So(services, ShouldHaveLength, 2)

}

func TestDiscoveryDiscover(t *testing.T) {
	a := New(t)

	client := getRedisClient(2)
	d := NewRedisDiscovery(client)
	defer func() {
		client.Del("discovery:announcement:router:router2.0")
		client.Del("discovery:announcement:broker:broker2.1")
		client.Del("discovery:announcement:broker:broker2.2")
	}()

	// This depends on the previous test to pass
	d.Announce(&pb.Announcement{ServiceName: "router", Id: "router2.0"})
	d.Announce(&pb.Announcement{ServiceName: "broker", Id: "broker2.1"})
	d.Announce(&pb.Announcement{ServiceName: "broker", Id: "broker2.2"})

	services, err := d.GetAll("random", 0, 0)
	a.So(err, ShouldBeNil)
	a.So(services, ShouldBeEmpty)

	services, err = d.GetAll("router", 0, 0)
	a.So(err, ShouldBeNil)
	a.So(services, ShouldHaveLength, 1)
	a.So(services[0].Id, ShouldEqual, "router2.0")

	services, err = d.GetAll("broker", 0, 0)
	a.So(err, ShouldBeNil)
	a.So(services, ShouldHaveLength, 2)

}

func TestDiscoveryMetadata(t *testing.T) {
	a := New(t)

	client := getRedisClient(1)
	d := NewRedisDiscovery(client)
	defer func() {
		client.Del("discovery:announcement:broker:broker3")
		client.Del("discovery:announcement:broker:broker4")
		client.Del("discovery:app-id:app-id-2")
	}()

	broker3 := &pb.Announcement{ServiceName: "broker", Id: "broker3", Metadata: []*pb.Metadata{&pb.Metadata{
		Metadata: &pb.Metadata_AppId{AppId: "app-id-1"},
	}}}
	broker4 := &pb.Announcement{ServiceName: "broker", Id: "broker4", Metadata: []*pb.Metadata{&pb.Metadata{
		Metadata: &pb.Metadata_AppId{AppId: "app-id-2"},
	}}}

	// Announce should not change metadata
	err := d.Announce(broker3)
	a.So(err, ShouldBeNil)
	service, err := d.Get("broker", "broker3")
	a.So(err, ShouldBeNil)
	a.So(service.Metadata, ShouldHaveLength, 0)

	d.Announce(broker4)

	// AddMetadata should add one
	err = d.AddMetadata("broker", "broker3", &pb.Metadata{
		Metadata: &pb.Metadata_AppId{AppId: "app-id-2"},
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
		Metadata: &pb.Metadata_AppId{AppId: "app-id-2"},
	})
	service, err = d.Get("broker", "broker3")
	a.So(err, ShouldBeNil)
	a.So(service.Metadata, ShouldHaveLength, 1)

	// DeleteMetadata for non-existing should not delete one
	err = d.DeleteMetadata("broker", "broker3", &pb.Metadata{
		Metadata: &pb.Metadata_AppId{AppId: "app-id-3"},
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
		Metadata: &pb.Metadata_AppId{AppId: "app-id-2"},
	})
	a.So(err, ShouldBeNil)
	service, err = d.Get("broker", "broker3")
	a.So(err, ShouldBeNil)
	a.So(service.Metadata, ShouldHaveLength, 0)

}

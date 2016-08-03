// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package announcement

import (
	"fmt"
	"os"
	"testing"

	pb "github.com/TheThingsNetwork/ttn/api/discovery"
	. "github.com/smartystreets/assertions"
	"gopkg.in/redis.v3"
)

func getRedisClient() *redis.Client {
	host := os.Getenv("REDIS_HOST")
	if host == "" {
		host = "localhost"
	}
	return redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:6379", host),
		Password: "", // no password set
		DB:       1,  // use default DB
	})
}

func TestAnnouncementStore(t *testing.T) {
	a := New(t)

	stores := map[string]Store{
		"local": NewAnnouncementStore(),
		"redis": NewRedisAnnouncementStore(getRedisClient()),
	}

	for name, s := range stores {

		t.Logf("Testing %s store", name)

		// Get non-existing
		dev, err := s.Get("router", "router1")
		a.So(err, ShouldNotBeNil)
		a.So(dev, ShouldBeNil)

		// Create
		err = s.Set(&pb.Announcement{
			ServiceName: "router",
			Id:          "router1",
		})
		a.So(err, ShouldBeNil)

		// Get existing
		dev, err = s.Get("router", "router1")
		a.So(err, ShouldBeNil)
		a.So(dev, ShouldNotBeNil)

		// Create extra
		err = s.Set(&pb.Announcement{
			ServiceName: "broker",
			Id:          "broker1",
		})
		a.So(err, ShouldBeNil)

		// List
		announcements, err := s.List()
		a.So(err, ShouldBeNil)
		a.So(announcements, ShouldHaveLength, 2)

		// List
		announcements, err = s.ListService("router")
		a.So(err, ShouldBeNil)
		a.So(announcements, ShouldHaveLength, 1)

		// Delete
		err = s.Delete("router", "router1")
		a.So(err, ShouldBeNil)

		// Get deleted
		dev, err = s.Get("router", "router1")
		a.So(err, ShouldNotBeNil)
		a.So(dev, ShouldBeNil)

		// Cleanup
		s.Delete("broker", "broker1")
	}

}

// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package kv

import (
	"fmt"
	"os"
	"testing"

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
		"local": NewKVStore(),
		"redis": NewRedisKVStore(getRedisClient(), "test"),
	}

	for name, s := range stores {

		t.Logf("Testing %s store", name)

		// Get non-existing
		res, err := s.Get("some-key")
		a.So(err, ShouldNotBeNil)
		a.So(res, ShouldBeEmpty)

		// Create
		err = s.Set("some-key", "some-value")
		a.So(err, ShouldBeNil)

		// Get existing
		res, err = s.Get("some-key")
		a.So(err, ShouldBeNil)
		a.So(res, ShouldEqual, "some-value")

		// Create extra
		err = s.Set("other-key", "other-value")
		a.So(err, ShouldBeNil)

		// List
		resps, err := s.List()
		a.So(err, ShouldBeNil)
		a.So(resps, ShouldHaveLength, 2)

		// Delete
		err = s.Delete("other-key")
		a.So(err, ShouldBeNil)

		// Get deleted
		res, err = s.Get("other-key")
		a.So(err, ShouldNotBeNil)
		a.So(res, ShouldBeEmpty)

		// Cleanup
		s.Delete("some-key")
	}

}

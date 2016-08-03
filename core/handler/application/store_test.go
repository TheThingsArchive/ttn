// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package application

import (
	"fmt"
	"os"
	"testing"

	"gopkg.in/redis.v3"

	. "github.com/smartystreets/assertions"
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

func TestApplicationStore(t *testing.T) {
	a := New(t)

	stores := map[string]Store{
		"local": NewApplicationStore(),
		"redis": NewRedisApplicationStore(getRedisClient()),
	}

	for name, s := range stores {

		t.Logf("Testing %s store", name)

		appID := "AppID-1"

		// Get non-existing
		app, err := s.Get(appID)
		a.So(err, ShouldNotBeNil)
		a.So(app, ShouldBeNil)

		// Create
		err = s.Set(&Application{
			AppID: appID,
		})
		a.So(err, ShouldBeNil)

		// Get existing
		app, err = s.Get(appID)
		a.So(err, ShouldBeNil)
		a.So(app, ShouldNotBeNil)

		// List
		apps, err := s.List()
		a.So(err, ShouldBeNil)
		a.So(apps, ShouldHaveLength, 1)

		// Delete
		err = s.Delete(appID)
		a.So(err, ShouldBeNil)

		// Get deleted
		app, err = s.Get(appID)
		a.So(err, ShouldNotBeNil)
		a.So(app, ShouldBeNil)
	}
}

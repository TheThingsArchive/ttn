// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package gateway

import (
	"fmt"
	"os"
	"testing"

	pb_gateway "github.com/TheThingsNetwork/ttn/api/gateway"
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

func TestNewGatewayStatusStore(t *testing.T) {
	a := New(t)
	client := getRedisClient()
	id := "0000000000000000"
	store := NewRedisStatusStore(client, id)
	a.So(store, ShouldNotBeNil)
	store = NewStatusStore()
	a.So(store, ShouldNotBeNil)
}

func TestStatusGetUpsert(t *testing.T) {
	a := New(t)
	store := NewStatusStore()

	// Get non-existing gateway status -> expect empty
	status, err := store.Get()
	a.So(err, ShouldBeNil)
	a.So(status, ShouldNotBeNil)
	a.So(*status, ShouldResemble, pb_gateway.Status{})

	// Update -> expect no error
	statusMessage := &pb_gateway.Status{Description: "Fake Gateway"}
	err = store.Update(statusMessage)
	a.So(err, ShouldBeNil)

	// Get existing gateway status -> expect status
	status, err = store.Get()
	a.So(err, ShouldBeNil)
	a.So(status, ShouldNotBeNil)
	a.So(*status, ShouldResemble, *statusMessage)
}

func TestRedisStatusGetUpsert(t *testing.T) {
	a := New(t)
	id := "0000000000000001"
	client := getRedisClient()
	store := NewRedisStatusStore(client, id)

	// Cleanup before and after
	client.Del(store.(*redisStatusStore).key)
	defer client.Del(store.(*redisStatusStore).key)

	// Get non-existing gateway status -> expect empty
	status, err := store.Get()
	a.So(err, ShouldBeNil)
	a.So(status, ShouldNotBeNil)
	a.So(*status, ShouldResemble, pb_gateway.Status{})

	// Update -> expect no error
	statusMessage := &pb_gateway.Status{Description: "Fake Gateway"}
	err = store.Update(statusMessage)
	a.So(err, ShouldBeNil)

	// Get existing gateway status -> expect status
	status, err = store.Get()
	a.So(err, ShouldBeNil)
	a.So(status, ShouldNotBeNil)
	a.So(*status, ShouldResemble, *statusMessage)
}

// TODO: Test error cases

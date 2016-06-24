// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package gateway

import (
	"testing"

	pb_gateway "github.com/TheThingsNetwork/ttn/api/gateway"
	"github.com/TheThingsNetwork/ttn/core/types"
	. "github.com/smartystreets/assertions"

	"gopkg.in/redis.v3"
)

func getRedisClient() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
}

func TestNewGatewayStatusStore(t *testing.T) {
	a := New(t)
	client := getRedisClient()
	eui := types.GatewayEUI{0, 0, 0, 0, 0, 0, 0, 0}
	store := NewRedisStatusStore(client, eui)
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
	eui := types.GatewayEUI{0, 0, 0, 0, 0, 0, 0, 1}
	client := getRedisClient()
	store := NewRedisStatusStore(client, eui)

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

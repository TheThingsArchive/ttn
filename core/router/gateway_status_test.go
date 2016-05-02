package router

import (
	"testing"

	"github.com/TheThingsNetwork/ttn/api/gateway"
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
	store := NewGatewayStatusStore(client)
	a.So(store, ShouldNotBeNil)
}

func TestGatewayGetUpsert(t *testing.T) {
	a := New(t)
	eui := types.GatewayEUI{0, 0, 0, 0, 0, 0, 0, 1}
	client := getRedisClient()
	store := NewGatewayStatusStore(client)

	// Cleanup before and after
	client.Del(store.(*redisGatewayStore).key(eui))
	// defer client.Del(store.(*redisGatewayStore).key(eui))

	// Get non-existing gateway status -> expect empty
	status, err := store.Get(eui)
	a.So(err, ShouldBeNil)
	a.So(status, ShouldNotBeNil)
	a.So(*status, ShouldResemble, gateway.StatusMessage{})

	// Upsert -> expect no error
	statusMessage := &gateway.StatusMessage{Description: "Fake Gateway"}
	err = store.Upsert(eui, statusMessage)
	a.So(err, ShouldBeNil)

	// Get existing gateway status -> expect status
	status, err = store.Get(eui)
	a.So(err, ShouldBeNil)
	a.So(status, ShouldNotBeNil)
	a.So(*status, ShouldResemble, *statusMessage)
}

func TestHandleGatewayStatus(t *testing.T) {
	a := New(t)
	eui := types.GatewayEUI{0, 0, 0, 0, 0, 0, 0, 2}
	client := getRedisClient()
	store := NewGatewayStatusStore(client)
	router := &router{
		gatewayStatusStore: store,
	}

	// Handle
	statusMessage := &gateway.StatusMessage{Description: "Fake Gateway"}
	err := router.HandleGatewayStatus(eui, statusMessage)
	a.So(err, ShouldBeNil)

	// Check storage
	status, err := store.Get(eui)
	a.So(err, ShouldBeNil)
	a.So(status, ShouldNotBeNil)
	a.So(*status, ShouldResemble, *statusMessage)
}

// TODO: Test error cases

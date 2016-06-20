package application

import (
	"testing"

	"gopkg.in/redis.v3"

	"github.com/TheThingsNetwork/ttn/core/types"
	. "github.com/smartystreets/assertions"
)

func getRedisClient() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
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

		appEUI := types.AppEUI{0, 0, 0, 0, 0, 0, 0, 1}

		// Non-existing App
		err := s.Set(&Application{
			AppEUI:            appEUI,
			HandlerID:         "handler1ID",
			HandlerNetAddress: "handler1NetAddress:1234",
		})
		a.So(err, ShouldBeNil)

		// Existing App
		err = s.Set(&Application{
			AppEUI:            appEUI,
			HandlerID:         "handler1ID",
			HandlerNetAddress: "handler1NetAddress2:1234",
		})
		a.So(err, ShouldBeNil)

		app, err := s.Get(types.AppEUI{0, 0, 0, 0, 0, 0, 0, 2})
		a.So(err, ShouldNotBeNil)
		a.So(app, ShouldBeNil)

		app, err = s.Get(appEUI)
		a.So(err, ShouldBeNil)
		a.So(app, ShouldNotBeNil)
		a.So(app.HandlerNetAddress, ShouldEqual, "handler1NetAddress2:1234")

		// List
		apps, err := s.List()
		a.So(err, ShouldBeNil)
		a.So(apps, ShouldHaveLength, 1)

		err = s.Delete(appEUI)
		a.So(err, ShouldBeNil)

		app, err = s.Get(appEUI)
		a.So(err, ShouldNotBeNil)
		a.So(app, ShouldBeNil)
	}
}

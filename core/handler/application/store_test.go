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

		// Get non-existing
		app, err := s.Get(appEUI)
		a.So(err, ShouldNotBeNil)
		a.So(app, ShouldBeNil)

		// Create
		err = s.Set(&Application{
			AppEUI: appEUI,
		})
		a.So(err, ShouldBeNil)

		// Get existing
		app, err = s.Get(appEUI)
		a.So(err, ShouldBeNil)
		a.So(app, ShouldNotBeNil)

		// List
		apps, err := s.List()
		a.So(err, ShouldBeNil)
		a.So(apps, ShouldHaveLength, 1)

		// Delete
		err = s.Delete(appEUI)
		a.So(err, ShouldBeNil)

		// Get deleted
		app, err = s.Get(appEUI)
		a.So(err, ShouldNotBeNil)
		a.So(app, ShouldBeNil)
	}
}

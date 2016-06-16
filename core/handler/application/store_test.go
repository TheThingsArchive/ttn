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
		DB:       0,  // use default DB
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

		// Get non-existing
		dev, err := s.Get(types.AppEUI{0, 0, 0, 0, 0, 0, 0, 1})
		a.So(err, ShouldNotBeNil)
		a.So(dev, ShouldBeNil)

		// Create
		err = s.Set(&Application{
			AppEUI: types.AppEUI{0, 0, 0, 0, 0, 0, 0, 1},
		})
		a.So(err, ShouldBeNil)

		// Get existing
		dev, err = s.Get(types.AppEUI{0, 0, 0, 0, 0, 0, 0, 1})
		a.So(err, ShouldBeNil)
		a.So(dev, ShouldNotBeNil)

		// Delete
		err = s.Delete(types.AppEUI{0, 0, 0, 0, 0, 0, 0, 1})
		a.So(err, ShouldBeNil)

		// Get deleted
		dev, err = s.Get(types.AppEUI{0, 0, 0, 0, 0, 0, 0, 1})
		a.So(err, ShouldNotBeNil)
		a.So(dev, ShouldBeNil)
	}
}

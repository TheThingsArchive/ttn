// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package storage

import (
	"testing"

	. "github.com/smartystreets/assertions"
	redis "gopkg.in/redis.v5"
)

type oldStruct struct {
	FirstName string `redis:"first_name"`
	LastName  string `redis:"last_name"`
}

func (s *oldStruct) DBVersion() string {
	return ""
}

type newStruct struct {
	Name string `redis:"name"`
}

func (s *newStruct) DBVersion() string {
	return "1"
}

func TestRedisMapMigrate(t *testing.T) {
	a := New(t)
	c := getRedisClient()
	s := NewRedisMapStore(c, "test-redis-map-migrate")
	a.So(s, ShouldNotBeNil)

	defer func() {
		s.Delete("test")
	}()

	s.SetBase(&oldStruct{}, "")
	s.Create("test", &oldStruct{
		FirstName: "First",
		LastName:  "Last",
	})

	{
		oldI, _ := s.Get("test")
		old := oldI.(*oldStruct)
		a.So(old.FirstName, ShouldEqual, "First")
		a.So(old.LastName, ShouldEqual, "Last")
	}

	{
		err := s.Migrate("")
		a.So(err, ShouldBeNil)
	}

	{
		s.SetBase(&newStruct{}, "")
		newI, _ := s.Get("test")
		new := newI.(*newStruct)
		a.So(new.Name, ShouldBeEmpty)
	}

	{
		s.SetBase(&newStruct{}, "")
		s.AddMigration("", func(_ *redis.Client, key string, obj map[string]string) (string, map[string]string, error) {
			firstName, _ := obj["first_name"]
			delete(obj, "first_name")

			lastName, _ := obj["last_name"]
			delete(obj, "last_name")

			obj["name"] = firstName + " " + lastName

			return "1", obj, nil
		})
		newI, _ := s.Get("test")
		new := newI.(*newStruct)
		a.So(new.Name, ShouldEqual, "First Last")
	}

}

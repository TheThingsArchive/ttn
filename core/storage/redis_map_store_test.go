// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package storage

import (
	"fmt"
	"testing"
	"time"

	"github.com/TheThingsNetwork/ttn/utils/errors"
	. "github.com/smartystreets/assertions"
)

type testRedisStruct struct {
	unexported string
	Skipped    string             `redis:"-"`
	Name       string             `redis:"name,omitempty"`
	EmptyStr   string             `redis:"EmptyStr"`
	UpdatedAt  Time               `redis:"updated_at,omitempty"`
	Empty      *map[string]string `redis:"empty"`
	NotEmpty   *map[string]string `redis:"not_empty"`
	changed    []string
}

func (a testRedisStruct) ChangedFields() (changed []string) {
	return a.changed
}

func TestRedisMapStore(t *testing.T) {
	a := New(t)
	c := getRedisClient()
	s := NewRedisMapStore(c, "test-redis-map-store")
	a.So(s, ShouldNotBeNil)

	now := time.Now()
	notEmpty := map[string]string{"ab": "cd"}
	testRedisStructVal := testRedisStruct{
		Name:      "My Name",
		UpdatedAt: Time{now},
		NotEmpty:  &notEmpty,
		changed:   []string{"Name", "UpdatedAt", "NotEmpty"},
	}

	s.SetBase(testRedisStructVal, "")

	// Get non-existing
	{
		res, err := s.Get("test")
		a.So(err, ShouldNotBeNil)
		a.So(errors.GetErrType(err), ShouldEqual, errors.NotFound)
		a.So(res, ShouldBeNil)
	}

	// Create New
	{
		defer func() {
			c.Del("test-redis-map-store:test").Result()
		}()
		err := s.Create("test", &testRedisStructVal)
		a.So(err, ShouldBeNil)

		exists, err := c.Exists("test-redis-map-store:test").Result()
		a.So(err, ShouldBeNil)
		a.So(exists, ShouldBeTrue)
	}

	// Create Existing
	{
		err := s.Create("test", testRedisStructVal)
		a.So(err, ShouldNotBeNil)
	}

	// Get
	{
		res, err := s.Get("test")
		a.So(err, ShouldBeNil)
		a.So(res, ShouldNotBeNil)
		a.So(res.(testRedisStruct).Name, ShouldEqual, "My Name")
		a.So(res.(testRedisStruct).UpdatedAt.Nanosecond(), ShouldEqual, now.Nanosecond())
	}

	// GetFields
	{
		res, err := s.GetFields("test", "name")
		a.So(err, ShouldBeNil)
		a.So(res, ShouldNotBeNil)
		a.So(res.(testRedisStruct).Name, ShouldEqual, "My Name")
	}

	for i := 1; i < 10; i++ {
		// Create Extra
		{
			name := fmt.Sprintf("test-%d", i)
			defer func() {
				c.Del("test-redis-map-store:" + name).Result()
			}()
			s.Create(name, testRedisStruct{
				Name:    name,
				changed: []string{"Name"},
			})
		}
	}

	// GetAll
	{
		res, err := s.GetAll([]string{"test"}, nil)
		a.So(err, ShouldBeNil)
		a.So(res, ShouldHaveLength, 1)
		a.So(res[0].(testRedisStruct).Name, ShouldEqual, "My Name")
	}

	// List
	{
		res, err := s.List("", nil)
		a.So(err, ShouldBeNil)
		a.So(res, ShouldHaveLength, 10)
		a.So(res[0].(testRedisStruct).Name, ShouldEqual, "My Name")
	}

	// List With Options
	{
		res, _ := s.List("test-*", &ListOptions{Limit: 2})
		a.So(res, ShouldHaveLength, 2)
		a.So(res[0].(testRedisStruct).Name, ShouldEqual, "test-1")
		a.So(res[1].(testRedisStruct).Name, ShouldEqual, "test-2")

		res, _ = s.List("test-*", &ListOptions{Limit: 20})
		a.So(res, ShouldHaveLength, 9)

		res, _ = s.List("test-*", &ListOptions{Offset: 20})
		a.So(res, ShouldHaveLength, 0)

		res, _ = s.List("test-*", &ListOptions{Limit: 2, Offset: 1})
		a.So(res, ShouldHaveLength, 2)
		a.So(res[0].(testRedisStruct).Name, ShouldEqual, "test-2")
		a.So(res[1].(testRedisStruct).Name, ShouldEqual, "test-3")

		res, _ = s.List("test-*", &ListOptions{Limit: 20, Offset: 1})
		a.So(res, ShouldHaveLength, 8)
	}

	// Update Non-Existing
	{
		err := s.Update("not-there", &testRedisStructVal)
		a.So(err, ShouldNotBeNil)
	}

	// Create/Update/Set using ChangedFields
	{
		err := s.Create("test", &testRedisStruct{})
		a.So(err, ShouldBeNil) // Create not executed, so we don't know it's already there
		exists, _ := c.Exists("test-redis-map-store:new").Result()
		a.So(exists, ShouldBeFalse)

		err = s.Update("not-there", &testRedisStruct{})
		a.So(err, ShouldBeNil) // Update not executed, so we don't know it's not there
		exists, _ = c.Exists("test-redis-map-store:new").Result()
		a.So(exists, ShouldBeFalse)

		err = s.Set("test", &testRedisStruct{
			Name: "Not-changed Name",
		})
		a.So(err, ShouldBeNil)

		name, err := c.HGet("test-redis-map-store:test", "name").Result()
		a.So(err, ShouldBeNil)
		a.So(name, ShouldEqual, "My Name")
	}

	// Set
	{
		err := s.Set("test", &testRedisStruct{
			Name: "Other Name",
		}, "Name")
		a.So(err, ShouldBeNil)

		name, err := c.HGet("test-redis-map-store:test", "name").Result()
		a.So(err, ShouldBeNil)
		a.So(name, ShouldEqual, "Other Name")
	}

	// Update Existing
	{
		err := s.Update("test", &testRedisStruct{
			Name: "New Name",
		}, "Name")
		a.So(err, ShouldBeNil)

		name, err := c.HGet("test-redis-map-store:test", "name").Result()
		a.So(err, ShouldBeNil)
		a.So(name, ShouldEqual, "New Name")
	}

	// Delete
	{
		err := s.Delete("test")
		a.So(err, ShouldBeNil)

		exists, err := c.Exists("test-redis-map-store:test").Result()
		a.So(err, ShouldBeNil)
		a.So(exists, ShouldBeFalse)
	}

}

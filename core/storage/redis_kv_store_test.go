// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package storage

import (
	"fmt"
	"testing"

	"github.com/TheThingsNetwork/ttn/utils/errors"
	. "github.com/smartystreets/assertions"
)

func TestRedisKVStore(t *testing.T) {
	a := New(t)
	c := getRedisClient()
	s := NewRedisKVStore(c, "test-redis-kv-store")
	a.So(s, ShouldNotBeNil)

	// Get non-existing
	{
		_, err := s.Get("test")
		a.So(err, ShouldNotBeNil)
		a.So(errors.GetErrType(err), ShouldEqual, errors.NotFound)
	}

	// Create New
	{
		defer func() {
			c.Del("test-redis-kv-store:test").Result()
		}()
		err := s.Create("test", "value")
		a.So(err, ShouldBeNil)

		exists, err := c.Exists("test-redis-kv-store:test").Result()
		a.So(err, ShouldBeNil)
		a.So(exists, ShouldBeTrue)
	}

	// Create Existing
	{
		err := s.Create("test", "value")
		a.So(err, ShouldNotBeNil)
	}

	// Get
	{
		res, err := s.Get("test")
		a.So(err, ShouldBeNil)
		a.So(res, ShouldEqual, "value")
	}

	for i := 1; i < 10; i++ {
		// Create Extra
		{
			name := fmt.Sprintf("test-%d", i)
			defer func() {
				c.Del("test-redis-kv-store:" + name).Result()
			}()
			s.Create(name, name)
		}
	}

	// GetAll
	{
		res, err := s.GetAll([]string{"test"}, nil)
		a.So(err, ShouldBeNil)
		a.So(res, ShouldHaveLength, 1)
		a.So(res["test"], ShouldEqual, "value")
	}

	// List
	{
		res, err := s.List("", nil)
		a.So(err, ShouldBeNil)
		a.So(res, ShouldHaveLength, 10)
		a.So(res["test"], ShouldEqual, "value")
	}

	// List With Options
	{
		res, _ := s.List("test-*", &ListOptions{Limit: 2})
		a.So(res, ShouldHaveLength, 2)
		a.So(res["test-1"], ShouldEqual, "test-1")
		a.So(res["test-2"], ShouldEqual, "test-2")

		res, _ = s.List("test-*", &ListOptions{Limit: 20})
		a.So(res, ShouldHaveLength, 9)

		res, _ = s.List("test-*", &ListOptions{Offset: 20})
		a.So(res, ShouldHaveLength, 0)

		res, _ = s.List("test-*", &ListOptions{Limit: 2, Offset: 1})
		a.So(res, ShouldHaveLength, 2)
		a.So(res["test-2"], ShouldEqual, "test-2")
		a.So(res["test-3"], ShouldEqual, "test-3")

		res, _ = s.List("test-*", &ListOptions{Limit: 20, Offset: 1})
		a.So(res, ShouldHaveLength, 8)
	}

	// Update Non-Existing
	{
		err := s.Update("not-there", "value")
		a.So(err, ShouldNotBeNil)
	}

	// Set
	{
		err := s.Set("test", "other")
		a.So(err, ShouldBeNil)

		name, err := c.Get("test-redis-kv-store:test").Result()
		a.So(err, ShouldBeNil)
		a.So(name, ShouldEqual, "other")
	}

	// Update Existing
	{
		err := s.Update("test", "updated")
		a.So(err, ShouldBeNil)

		name, err := c.Get("test-redis-kv-store:test").Result()
		a.So(err, ShouldBeNil)
		a.So(name, ShouldEqual, "updated")
	}

	// Delete
	{
		err := s.Delete("test")
		a.So(err, ShouldBeNil)

		exists, err := c.Exists("test-redis-kv-store:test").Result()
		a.So(err, ShouldBeNil)
		a.So(exists, ShouldBeFalse)
	}

}

// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package storage

import (
	"fmt"
	"testing"

	"github.com/TheThingsNetwork/ttn/utils/errors"
	. "github.com/smartystreets/assertions"
)

func TestRedisSetStore(t *testing.T) {
	a := New(t)
	c := getRedisClient()
	s := NewRedisSetStore(c, "test-redis-set-store")
	a.So(s, ShouldNotBeNil)

	// Get non-existing
	{
		_, err := s.Get("test")
		a.So(err, ShouldNotBeNil)
		a.So(errors.GetErrType(err), ShouldEqual, errors.NotFound)

		contains, err := s.Contains("test", "value")
		a.So(err, ShouldBeNil)
		a.So(contains, ShouldBeFalse)
	}

	defer func() {
		c.Del("test-redis-set-store:test").Result()
	}()

	// Add
	{
		err := s.Add("test", "value")
		a.So(err, ShouldBeNil)

		contains, err := s.Contains("test", "value")
		a.So(err, ShouldBeNil)
		a.So(contains, ShouldBeTrue)

		contains, err = s.Contains("test", "not-there")
		a.So(err, ShouldBeNil)
		a.So(contains, ShouldBeFalse)
	}

	// Get
	{
		res, err := s.Get("test")
		a.So(err, ShouldBeNil)
		a.So(res, ShouldHaveLength, 1)

		count, err := s.Count("test")
		a.So(err, ShouldBeNil)
		a.So(count, ShouldEqual, 1)
	}

	// Get More
	{
		s.Add("test", "othervalue")

		res, err := s.Get("test")
		a.So(err, ShouldBeNil)
		a.So(res, ShouldHaveLength, 2)

		count, err := s.Count("test")
		a.So(err, ShouldBeNil)
		a.So(count, ShouldEqual, 2)
	}

	// Remove
	{
		s.Remove("test", "othervalue")

		res, err := s.Get("test")
		a.So(err, ShouldBeNil)
		a.So(res, ShouldHaveLength, 1)

		count, err := s.Count("test")
		a.So(err, ShouldBeNil)
		a.So(count, ShouldEqual, 1)
	}

	for i := 1; i < 10; i++ {
		// Create Extra
		{
			name := fmt.Sprintf("test-%d", i)
			defer func() {
				c.Del("test-redis-set-store:" + name).Result()
			}()
			s.Add(name, name)
		}
	}

	// GetAll
	{
		res, err := s.GetAll([]string{"test"}, nil)
		a.So(err, ShouldBeNil)
		a.So(res, ShouldHaveLength, 1)
		a.So(res["test"], ShouldHaveLength, 1)
	}

	// List
	{
		res, err := s.List("", nil)
		a.So(err, ShouldBeNil)
		a.So(res, ShouldHaveLength, 10)
		a.So(res["test"], ShouldHaveLength, 1)
	}

	// List With Options
	{
		res, _ := s.List("test-*", &ListOptions{Limit: 2})
		a.So(res, ShouldHaveLength, 2)
		a.So(res["test-1"], ShouldResemble, []string{"test-1"})
		a.So(res["test-2"], ShouldResemble, []string{"test-2"})

		res, _ = s.List("test-*", &ListOptions{Limit: 20})
		a.So(res, ShouldHaveLength, 9)

		res, _ = s.List("test-*", &ListOptions{Offset: 20})
		a.So(res, ShouldHaveLength, 0)

		res, _ = s.List("test-*", &ListOptions{Limit: 2, Offset: 1})
		a.So(res, ShouldHaveLength, 2)
		a.So(res["test-2"], ShouldResemble, []string{"test-2"})
		a.So(res["test-3"], ShouldResemble, []string{"test-3"})

		res, _ = s.List("test-*", &ListOptions{Limit: 20, Offset: 1})
		a.So(res, ShouldHaveLength, 8)
	}

	// Delete
	{
		err := s.Delete("test-1")
		a.So(err, ShouldBeNil)

		res, err := s.List("", nil)
		a.So(err, ShouldBeNil)
		a.So(res, ShouldHaveLength, 9)
	}

}

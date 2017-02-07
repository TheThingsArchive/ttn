// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package storage

import (
	"testing"

	. "github.com/smartystreets/assertions"
)

func TestRedisQueueStore(t *testing.T) {
	a := New(t)
	c := getRedisClient()
	s := NewRedisQueueStore(c, "test-redis-queue-store")
	a.So(s, ShouldNotBeNil)

	list, err := s.List("", nil)
	a.So(err, ShouldBeNil)
	a.So(list, ShouldBeEmpty)

	length, err := s.Length("test")
	a.So(err, ShouldBeNil)
	a.So(length, ShouldEqual, 0)

	res, err := s.Get("test")
	a.So(err, ShouldBeNil)
	a.So(res, ShouldBeEmpty)

	next, err := s.Next("test")
	a.So(err, ShouldBeNil)
	a.So(next, ShouldBeEmpty)

	defer func() {
		c.Del("test-redis-queue-store:test").Result()
	}()

	err = s.AddEnd("test", "value3", "value4")
	a.So(err, ShouldBeNil)

	err = s.AddFront("test", "value1", "value2")
	a.So(err, ShouldBeNil)

	list, err = s.List("", nil)
	a.So(err, ShouldBeNil)
	a.So(list, ShouldResemble, map[string][]string{
		"test": []string{"value2", "value1", "value3", "value4"},
	})

	length, err = s.Length("test")
	a.So(err, ShouldBeNil)
	a.So(length, ShouldEqual, 4)

	res, err = s.Get("test")
	a.So(err, ShouldBeNil)
	a.So(res, ShouldResemble, []string{"value2", "value1", "value3", "value4"})

	front, err := s.GetFront("test", 2)
	a.So(err, ShouldBeNil)
	a.So(front, ShouldResemble, []string{"value2", "value1"})

	end, err := s.GetEnd("test", 2)
	a.So(err, ShouldBeNil)
	a.So(end, ShouldResemble, []string{"value3", "value4"})

	next, err = s.Next("test")
	a.So(err, ShouldBeNil)
	a.So(next, ShouldEqual, "value2")

	length, err = s.Length("test")
	a.So(err, ShouldBeNil)
	a.So(length, ShouldEqual, 3)

	res, err = s.Get("test")
	a.So(err, ShouldBeNil)
	a.So(res, ShouldResemble, []string{"value1", "value3", "value4"})

	err = s.Trim("test", 2)
	a.So(err, ShouldBeNil)

	length, err = s.Length("test")
	a.So(err, ShouldBeNil)
	a.So(length, ShouldEqual, 2)

	res, err = s.Get("test")
	a.So(err, ShouldBeNil)
	a.So(res, ShouldResemble, []string{"value1", "value3"})

	err = s.Delete("test")
	a.So(err, ShouldBeNil)

	list, err = s.List("", nil)
	a.So(err, ShouldBeNil)
	a.So(list, ShouldBeEmpty)
}

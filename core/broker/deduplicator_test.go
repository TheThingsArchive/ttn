// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package broker

import (
	"sync"
	"testing"
	"time"

	. "github.com/smartystreets/assertions"
)

func TestCollectionAdd(t *testing.T) {
	a := New(t)
	c := newCollection()
	c.Add("item")
	a.So(c.values, ShouldContain, "item")
}

func TestCollectionGetAndClear(t *testing.T) {
	a := New(t)
	c := newCollection()
	a.So(c.GetAndClear(), ShouldBeEmpty)
	c.Add("item1")
	c.Add("item2")
	a.So(c.GetAndClear(), ShouldResemble, []interface{}{"item1", "item2"})
	a.So(c.GetAndClear(), ShouldBeEmpty)
}

func TestCollectionWaitDone(t *testing.T) {
	c := newCollection()
	c.done()
	c.wait()
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		c.wait()
		wg.Done()
	}()
	c.done()
	wg.Wait()
}

func TestDeduplicatorAdd(t *testing.T) {
	a := New(t)
	d := NewDeduplicator(5 * time.Millisecond).(*deduplicator)
	a.So(d.collections, ShouldBeEmpty)
	d.add("key", "item")
	a.So(d.collections, ShouldNotBeEmpty)
}

func TestDeduplicatorDeduplicate(t *testing.T) {
	a := New(t)
	d := NewDeduplicator(10 * time.Millisecond).(*deduplicator)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		res := d.Deduplicate("key", "value1")
		a.So(res, ShouldResemble, []interface{}{"value1", "value2", "value3"})
		wg.Done()
	}()

	<-time.After(5 * time.Millisecond)

	a.So(d.Deduplicate("key", "value2"), ShouldBeNil)
	a.So(d.Deduplicate("key", "value3"), ShouldBeNil)

	wg.Wait()
}

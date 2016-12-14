// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package broker

import (
	"sync"
	"time"
)

type collection struct {
	sync.Mutex
	ready  chan bool
	values []interface{}
}

func newCollection() *collection {
	return &collection{
		ready:  make(chan bool, 1),
		values: []interface{}{},
	}
}

func (c *collection) Add(value interface{}) {
	c.Lock()
	defer c.Unlock()
	c.values = append(c.values, value)
}

func (c *collection) GetAndClear() []interface{} {
	c.Lock()
	defer c.Unlock()
	values := c.values
	c.values = []interface{}{}
	return values
}

func (c *collection) done() {
	c.ready <- true
}

func (c *collection) wait() {
	<-c.ready
}

type Deduplicator interface {
	Deduplicate(key string, value interface{}) []interface{}
}

type deduplicator struct {
	sync.Mutex
	timeout     time.Duration
	collections map[string]*collection
}

func (d *deduplicator) add(key string, value interface{}) (c *collection, isFirst bool) {
	d.Lock()
	defer d.Unlock()
	var ok bool
	if c, ok = d.collections[key]; ok {
		c.Add(value)
	} else {
		isFirst = true
		c = newCollection()
		c.Add(value)
		d.collections[key] = c
	}
	return
}

func (d *deduplicator) Deduplicate(key string, value interface{}) (values []interface{}) {
	collection, isFirst := d.add(key, value)
	if isFirst {
		go func() {
			<-time.After(d.timeout)
			collection.done()
			<-time.After(d.timeout)
			d.Lock()
			defer d.Unlock()
			delete(d.collections, key)
		}()
		collection.wait()
		values = collection.GetAndClear()
	}
	return
}

func NewDeduplicator(timeout time.Duration) Deduplicator {
	return &deduplicator{
		timeout:     timeout,
		collections: map[string]*collection{},
	}
}

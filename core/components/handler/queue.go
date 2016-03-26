// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package handler

import (
	"reflect"
	"sync"
)

type pQueue interface {
	Put([]byte)
	Contains([]byte) bool
	Remove([]byte)
}

type pq struct {
	sync.RWMutex // Guars ids & values
	ids          []pQueueEntry
	values       map[[17]byte]pQueueEntry
	size         int
	ticket       int64
}

type pQueueEntry struct {
	Ticket int64
	ID     [17]byte
	Value  []byte
}

func newPQueue(size uint) pQueue {
	return &pq{
		size:   int(size),
		values: make(map[[17]byte]pQueueEntry),
	}
}

// Put Insert a new entry in the queue. If the queue is full, then the oldest entry is discarded
func (q *pq) Put(pid []byte) {
	q.Lock()
	id, v := q.fromPid(pid)
	q.ticket++
	entry := pQueueEntry{Ticket: q.ticket, Value: v, ID: id}
	q.values[id] = entry
	if len(q.values) > q.size {
		var i = 0
		for _, e := range q.ids {
			if v, ok := q.values[e.ID]; ok && v.ID == e.ID && v.Ticket == e.Ticket {
				delete(q.values, e.ID)
				q.ids = q.ids[i+1:]
				break
			}
			i++
		}
	}
	q.ids = append(q.ids, entry)
	q.Unlock()
}

// Remove discard an entry from the map
func (q *pq) Remove(pid []byte) {
	q.Lock()
	id, _ := q.fromPid(pid)
	delete(q.values, id)
	q.Unlock()
}

// Contains check whether a value exist for the given id
func (q *pq) Contains(pid []byte) bool {
	q.RLock()
	defer q.RUnlock()
	id, v := q.fromPid(pid)
	if e, ok := q.values[id]; ok {
		return reflect.DeepEqual(e.Value[:], v)
	}
	return false
}

func (q *pq) fromPid(pid []byte) ([17]byte, []byte) {
	var id [17]byte
	copy(id[:], pid[:17])
	return id, pid[17:]
}

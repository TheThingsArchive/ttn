package gateway

import (
	"container/heap"
	"sync"
	"time"

	router_pb "github.com/TheThingsNetwork/ttn/api/router"
)

type scheduledItem struct {
	id        string
	time      time.Time
	timestamp uint32
	length    uint32
	score     uint
	payload   *router_pb.DownlinkMessage
}

// A downlinkQueue implements heap.Interface and holds scheduledItems.
type downlinkQueue struct {
	sync.RWMutex
	items []*scheduledItem
}

func NewDownlinkQueue(items ...*scheduledItem) *downlinkQueue {
	dq := &downlinkQueue{
		items: items,
	}
	heap.Init(dq)
	return dq
}

// Len is used by heap.Interface
func (dq *downlinkQueue) Len() int { return len(dq.items) }

// Less is used by heap.Interface
func (dq *downlinkQueue) Less(i, j int) bool {
	return dq.items[i].time.Before(dq.items[j].time)
}

// Swap is used by heap.Interface.
func (dq *downlinkQueue) Swap(i, j int) {
	if len(dq.items) == 0 {
		return
	}
	dq.Lock()
	defer dq.Unlock()
	dq.items[i], dq.items[j] = dq.items[j], dq.items[i]
}

// Push is used by heap.Interface
func (dq *downlinkQueue) Push(x interface{}) {
	item := x.(*scheduledItem)
	dq.Lock()
	defer dq.Unlock()
	dq.items = append(dq.items, item)
}

// Pop is used by heap.Interface.
func (dq *downlinkQueue) Pop() interface{} {
	dq.Lock()
	defer dq.Unlock()
	n := len(dq.items)
	if n == 0 {
		return nil
	}
	item := dq.items[n-1]
	dq.items = dq.items[0 : n-1]
	return item
}

// Snapshot returns a snapshot of the downlinkQueue
func (dq *downlinkQueue) Snapshot() []*scheduledItem {
	dq.RLock()
	defer dq.RUnlock()
	return dq.items
}

// Peek returns the next item in the queue
func (dq *downlinkQueue) Peek() interface{} {
	snapshot := dq.Snapshot()
	n := len(snapshot)
	if n == 0 {
		return nil
	}
	item := snapshot[n-1]
	return item
}

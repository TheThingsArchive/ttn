package gateway

import (
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

// A downlinkQueue holds scheduledItems.
type downlinkQueue struct {
	sync.RWMutex
	items []*scheduledItem
}

// NewDownlinkQueue creates a new downlinkQueue
func NewDownlinkQueue(items ...*scheduledItem) *downlinkQueue {
	dq := &downlinkQueue{
		items: items,
	}
	return dq
}

// less is used for sorting
func (dq *downlinkQueue) less(i, j int) bool {
	return dq.items[i].time.Before(dq.items[j].time)
}

// swap is used for sorting
func (dq *downlinkQueue) swap(i, j int) {
	dq.items[i], dq.items[j] = dq.items[j], dq.items[i]
}

// Push an item to the queue
func (dq *downlinkQueue) Push(item *scheduledItem) {
	dq.items = append(dq.items, item)
	// TODO: Insertion sort is nice, but can be optimized for the use-case of TTN (LoRaWAN RX1 and RX2)
	for i := len(dq.items); i > 1; i-- {
		if dq.less(i-1, i-2) {
			dq.swap(i-1, i-2)
		} else {
			return
		}
	}
}

// Pop an item from the queue
func (dq *downlinkQueue) Pop() *scheduledItem {
	n := len(dq.items)
	if n == 0 {
		return nil
	}
	item := dq.items[0]
	dq.items = dq.items[1:]
	return item
}

// Snapshot returns a snapshot of the downlinkQueue
func (dq *downlinkQueue) Snapshot() []*scheduledItem {
	return dq.items
}

// Peek returns the next item in the queue
func (dq *downlinkQueue) Peek() *scheduledItem {
	n := len(dq.items)
	if n == 0 {
		return nil
	}
	return dq.items[0]
}

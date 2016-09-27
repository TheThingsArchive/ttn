// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package gateway

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	pb_lorawan "github.com/TheThingsNetwork/ttn/api/protocol/lorawan"
	router_pb "github.com/TheThingsNetwork/ttn/api/router"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	"github.com/TheThingsNetwork/ttn/utils/random"
	"github.com/TheThingsNetwork/ttn/utils/toa"
	"github.com/apex/log"
)

// Schedule is used to schedule downlink transmissions
type Schedule interface {
	fmt.GoStringer
	// Synchronize the schedule with the gateway timestamp (in microseconds)
	Sync(timestamp uint32)
	// Get an "option" on a transmission slot at timestamp for the maximum duration of length (both in microseconds)
	GetOption(timestamp uint32, length uint32) (id string, score uint)
	// Schedule a transmission on a slot
	Schedule(id string, downlink *router_pb.DownlinkMessage) error
	// Subscribe to downlink messages
	Subscribe() <-chan *router_pb.DownlinkMessage
	// Whether the gateway has active downlink
	IsActive() bool
	// Stop the subscription
	Stop()
}

// NewSchedule creates a new Schedule
func NewSchedule(ctx log.Interface) Schedule {
	s := &schedule{
		ctx:    ctx,
		items:  make(map[string]*scheduledItem),
		random: random.New(),
	}
	go func() {
		for {
			<-time.After(10 * time.Second)
			s.RLock()
			numItems := len(s.items)
			s.RUnlock()
			if numItems > 0 {
				s.Lock()
				for id, item := range s.items {
					// Delete the item if we are more than 2 seconds after the deadline
					if time.Now().After(item.deadlineAt.Add(2 * time.Second)) {
						delete(s.items, id)
					}
				}
				s.Unlock()
			}
		}
	}()
	return s
}

type scheduledItem struct {
	id         string
	deadlineAt time.Time
	timestamp  uint32
	length     uint32
	score      uint
	payload    *router_pb.DownlinkMessage
}

type schedule struct {
	sync.RWMutex
	random   *random.TTNRandom
	ctx      log.Interface
	offset   int64
	items    map[string]*scheduledItem
	downlink chan *router_pb.DownlinkMessage
}

func (s *schedule) GoString() (str string) {
	s.RLock()
	defer s.RUnlock()
	for _, item := range s.items {
		str += fmt.Sprintf("%s at %s\n", item.id, item.deadlineAt)
	}
	return
}

// Deadline for sending a downlink back to the gateway
// TODO: Make configurable
var Deadline = 200 * time.Millisecond

const uintmax = 1 << 32

// getConflicts walks over the schedule and returns the number of conflicts.
// Both timestamp and length are in microseconds
func (s *schedule) getConflicts(timestamp uint32, length uint32) (conflicts uint) {
	s.RLock()
	defer s.RUnlock()
	for _, item := range s.items {
		scheduledFrom := uint64(item.timestamp) % uintmax
		scheduledTo := scheduledFrom + uint64(item.length)
		from := uint64(timestamp)
		to := from + uint64(length)

		if scheduledTo > uintmax || to > uintmax {
			if scheduledTo-uintmax <= from || scheduledFrom >= to-uintmax {
				continue
			}
		} else if scheduledTo <= from || scheduledFrom >= to {
			continue
		}

		if item.payload == nil {
			conflicts++
		} else {
			conflicts += 100
		}
	}
	return
}

// realtime gets the synchronized time for a timestamp (in microseconds). Time
// should first be syncronized using func Sync()
func (s *schedule) realtime(timestamp uint32) (t time.Time) {
	offset := atomic.LoadInt64(&s.offset)
	t = time.Unix(0, 0)
	t = t.Add(time.Duration(int64(timestamp)*1000 + offset))
	if t.Before(time.Now()) {
		t = t.Add(time.Duration(int64(1<<32) * 1000))
	}
	return
}

// see interface
func (s *schedule) Sync(timestamp uint32) {
	atomic.StoreInt64(&s.offset, time.Now().UnixNano()-int64(timestamp)*1000)
}

// see interface
func (s *schedule) GetOption(timestamp uint32, length uint32) (id string, score uint) {
	id = s.random.String(32)
	score = s.getConflicts(timestamp, length)
	item := &scheduledItem{
		id:         id,
		deadlineAt: s.realtime(timestamp).Add(-1 * Deadline),
		timestamp:  timestamp,
		length:     length,
		score:      score,
	}
	s.Lock()
	defer s.Unlock()
	s.items[id] = item
	return id, score
}

// see interface
func (s *schedule) Schedule(id string, downlink *router_pb.DownlinkMessage) error {
	ctx := s.ctx.WithField("Identifier", id)

	s.Lock()
	defer s.Unlock()
	if item, ok := s.items[id]; ok {
		item.payload = downlink

		if lorawan := downlink.GetProtocolConfiguration().GetLorawan(); lorawan != nil {
			var time time.Duration
			if lorawan.Modulation == pb_lorawan.Modulation_LORA {
				// Calculate max ToA
				time, _ = toa.ComputeLoRa(
					uint(len(downlink.Payload)),
					lorawan.DataRate,
					lorawan.CodingRate,
				)
			}
			if lorawan.Modulation == pb_lorawan.Modulation_FSK {
				// Calculate max ToA
				time, _ = toa.ComputeFSK(
					uint(len(downlink.Payload)),
					int(lorawan.BitRate),
				)
			}
			item.length = uint32(time / 1000)
		}

		if time.Now().Before(item.deadlineAt) {
			// Schedule transmission before the Deadline
			go func() {
				waitTime := item.deadlineAt.Sub(time.Now())
				ctx.WithField("Remaining", waitTime).Info("Scheduled downlink")
				<-time.After(waitTime)
				if s.downlink != nil {
					s.downlink <- item.payload
				}
			}()
		} else if s.downlink != nil {
			overdue := time.Now().Sub(item.deadlineAt)
			if overdue < Deadline {
				// Immediately send it
				ctx.WithField("Overdue", overdue).Warn("Send Late Downlink")
				s.downlink <- item.payload
			} else {
				ctx.WithField("Overdue", overdue).Warn("Discard Late Downlink")
			}
		} else {
			ctx.Warn("Unable to send Downlink")
		}

		return nil
	}
	return errors.NewErrNotFound(id)
}

func (s *schedule) Stop() {
	close(s.downlink)
	s.downlink = nil
}

func (s *schedule) Subscribe() <-chan *router_pb.DownlinkMessage {
	if s.downlink != nil {
		return nil
	}
	s.downlink = make(chan *router_pb.DownlinkMessage)
	return s.downlink
}

func (s *schedule) IsActive() bool {
	return s.downlink != nil
}

package gateway

import (
	"errors"
	"sync"
	"sync/atomic"
	"time"

	router_pb "github.com/TheThingsNetwork/ttn/api/router"
	"github.com/TheThingsNetwork/ttn/utils/random"
)

// Schedule is used to schedule downlink transmissions
type Schedule interface {
	// Synchronize the schedule with the gateway timestamp (in microseconds)
	Sync(timestamp uint32)
	// Get an "option" on a transmission slot at timestamp for the maximum duration of length (both in microseconds)
	GetOption(timestamp uint32, length uint32) (id string, score uint)
	// Schedule a transmission on a slot
	Schedule(id string, downlink *router_pb.DownlinkMessage) error
	// TODO: Add some way to retrieve the next scheduled item (preferably at the right time)
}

// NewSchedule creates a new Schedule
func NewSchedule() Schedule {
	return &schedule{
		queue: NewDownlinkQueue(),
		byID:  make(map[string]*scheduledItem),
	}
}

type schedule struct {
	sync.RWMutex
	active         bool
	offset         int64
	queue          *downlinkQueue
	byID           map[string]*scheduledItem
	downlinkActive bool
	downlink       chan *router_pb.DownlinkMessage
}

const uintmax = 1 << 32

// getConflicts walks over the schedule and returns the number of conflicts.
// Both timestamp and length are in microseconds
func (s *schedule) getConflicts(timestamp uint32, length uint32) (conflicts uint) {
	s.RLock()
	snapshot := s.queue.Snapshot()
	s.RUnlock()
	for _, item := range snapshot {
		scheduledFrom := uint64(item.timestamp) % uintmax
		scheduledTo := scheduledFrom + uint64(item.length)
		from := uint64(timestamp)
		to := from + uint64(length)

		if scheduledTo > uintmax || to > uintmax {
			if scheduledTo-uintmax < from || scheduledFrom > to-uintmax {
				continue
			}
		} else if scheduledTo < from || scheduledFrom > to {
			continue
		}

		if item.payload == nil {
			conflicts++
		} else {
			conflicts += 10 // TODO: Configure this
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
	id = random.String(32)
	score = s.getConflicts(timestamp, length)
	item := &scheduledItem{
		id:        id,
		time:      s.realtime(timestamp),
		timestamp: timestamp,
		length:    length,
		score:     score,
	}
	s.Lock()
	defer s.Unlock()
	s.queue.Push(item)
	s.byID[id] = item
	return id, score
}

// see interface
func (s *schedule) Schedule(id string, downlink *router_pb.DownlinkMessage) error {
	s.RLock()
	defer s.RUnlock()
	if item, ok := s.byID[id]; ok {
		item.payload = downlink
		return nil
	}
	return errors.New("ID not found")
}

func (s *schedule) Stop() {
	s.downlinkActive = false
}

// TODO: Make configurable
const deadline = 50
const waitTime = 10 * time.Millisecond

func (s *schedule) Subscribe() <-chan *router_pb.DownlinkMessage {
	if s.downlinkActive {
		return nil
	}
	s.downlink = make(chan *router_pb.DownlinkMessage)
	s.downlinkActive = true
	go func() {
		for s.downlinkActive {
			<-time.After(waitTime)
			s.Lock()
			for {
				item := s.queue.Peek()
				if item != nil && time.Now().Add(-1*deadline*time.Millisecond).After(item.time) {
					s.queue.Pop()
					delete(s.byID, item.id)
					if item.payload != nil {
						s.downlink <- item.payload
					}
				} else {
					break
				}
			}
			s.Unlock()
		}
		close(s.downlink)
	}()
	return s.downlink
}

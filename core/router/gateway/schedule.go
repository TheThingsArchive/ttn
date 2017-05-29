// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package gateway

import (
	"sync"
	"time"

	ttnlog "github.com/TheThingsNetwork/go-utils/log"
	"github.com/TheThingsNetwork/go-utils/queue"
	"github.com/TheThingsNetwork/ttn/api/fields"
	router_pb "github.com/TheThingsNetwork/ttn/api/router"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	"github.com/TheThingsNetwork/ttn/utils/random"
	"github.com/TheThingsNetwork/ttn/utils/toa"
)

type scheduledItem struct {
	mu             sync.RWMutex
	id             string
	time           time.Time
	timestamp      uint64 // microseconds - fullTimestamp
	duration       uint32 // microseconds
	durationOffAir uint32 // microseconds
	payload        *router_pb.DownlinkMessage
}

func (i *scheduledItem) Time() time.Time {
	i.mu.RLock()
	defer i.mu.RUnlock()
	return i.time
}
func (i *scheduledItem) Timestamp() int64 {
	i.mu.RLock()
	defer i.mu.RUnlock()
	return int64(i.timestamp) * int64(time.Microsecond)
}
func (i *scheduledItem) Duration() time.Duration {
	i.mu.RLock()
	defer i.mu.RUnlock()
	return time.Duration(i.duration+i.durationOffAir) * time.Microsecond
}

const maxUint32 = 1 << 32

func getFullTimestamp(full uint64, lsb uint32) uint64 {
	if int64(lsb)-int64(full) > 0 {
		return uint64(lsb)
	}
	if uint32(full%maxUint32) <= lsb {
		return uint64(lsb) + (full/maxUint32)*maxUint32
	}
	return uint64(lsb) + ((full/maxUint32)+1)*maxUint32
}

// Schedule for a gateway
type Schedule struct {
	sync.RWMutex
	ctx      ttnlog.Interface
	items    map[string]*scheduledItem
	schedule queue.Schedule
	queue    queue.JIT

	gateway   *Gateway
	timestamp uint64
	offset    int64

	durationOffAirFunc func(*router_pb.DownlinkMessage) time.Duration

	downlink            chan *router_pb.DownlinkMessage
	downlinkSubscribers int
}

// ErrScheduleInactive is returned when attempting an operation on an inactive schedule
var ErrScheduleInactive = errors.New("gateway: schedule not active")

// ErrScheduleConflict is returned when an item can not be scheduled due to a conflicting item
var ErrScheduleConflict = errors.New("gateway: conflict in schedule")

// NewSchedule creates a new Schedule
func NewSchedule(ctx ttnlog.Interface, durationOffAirFunc func(*router_pb.DownlinkMessage) time.Duration) *Schedule {
	return &Schedule{ctx: ctx, durationOffAirFunc: durationOffAirFunc}
}

func (s *Schedule) init() {
	schedule := queue.NewSchedule()
	queue := queue.NewJIT()
	items := make(map[string]*scheduledItem)
	downlink := make(chan *router_pb.DownlinkMessage, 1)

	s.schedule = schedule
	s.queue = queue
	s.items = items
	s.downlink = downlink

	// Send downlink to downlink channel
	go func() {
		for {
			nextI := queue.Next()
			if nextI == nil {
				break
			}
			next := nextI.(*scheduledItem)
			select {
			case downlink <- next.payload:
			default:
			}
		}
		close(downlink)
		s.Lock()
		s.queue = nil
		s.items = nil
		s.Unlock()
	}()

	// Clean up expired items
	go func() {
		for {
			expiredI := schedule.Next()
			if expiredI == nil {
				break
			}
		}
		queue.Destroy()
		s.Lock()
		s.schedule = nil
		s.Unlock()
	}()
}

// Sync the gateway schedule
func (s *Schedule) Sync(timestamp uint32) {
	s.Lock()
	defer s.Unlock()
	if s.timestamp == 0 {
		s.timestamp = uint64(timestamp)
	} else {
		s.timestamp = getFullTimestamp(s.timestamp, timestamp)
	}
	s.offset = time.Duration(time.Now().UnixNano() - int64(s.timestamp*1000)).Nanoseconds()
}

func (s *Schedule) getFullTimestamp(lsb uint32) uint64 {
	return getFullTimestamp(s.timestamp, lsb)
}

func (s *Schedule) getRealtime(fullTimestamp uint64) time.Time {
	return time.Unix(0, int64(s.offset)+int64(fullTimestamp)*1000)
}

func (s *Schedule) getTimestamp(t time.Time) uint64 {
	return uint64(t.Add(-1*time.Duration(s.offset)).UnixNano() / 1000)
}

// DefaultGatewayRTT is the default gateway round-trip-time if none is reported by the gateway itself
var DefaultGatewayRTT = 300 * time.Millisecond

// DefaultGatewayBufferTime is the default time the gateway buffers downlink messages
var DefaultGatewayBufferTime = 500 * time.Millisecond // TODO: add this to gateway status message

func (s *Schedule) getGatewayRTT() (rtt time.Duration) {
	if s.gateway != nil {
		rtt = time.Duration(s.gateway.Status().Rtt) * time.Millisecond
	}
	if rtt == 0 {
		rtt = DefaultGatewayRTT
	}
	return
}

// GetOption returns a new schedule option
func (s *Schedule) GetOption(timestamp uint32, length uint32) (id string, numConflicts uint, err error) {
	id = random.String(32)
	s.Lock()
	defer s.Unlock()
	if !s.isActive() {
		return id, numConflicts, ErrScheduleInactive
	}
	item := &scheduledItem{
		id:        id,
		timestamp: s.getFullTimestamp(timestamp),
		duration:  length,
	}
	item.time = s.getRealtime(item.timestamp)
	s.items[id] = item
	conflicts := s.schedule.Add(item)
	for _, conflict := range conflicts {
		if i, ok := conflict.(*scheduledItem); ok {
			i.mu.RLock()
			isScheduled := i.payload != nil
			i.mu.RUnlock()
			if isScheduled {
				return id, numConflicts, ErrScheduleConflict
			}
			numConflicts++
		}
	}
	return
}

func (s *Schedule) ScheduleASAP(downlink *router_pb.DownlinkMessage) error {
	s.Lock()
	defer s.Unlock()
	if !s.isActive() {
		return ErrScheduleInactive
	}
	toa, _ := toa.Compute(downlink)
	time := s.schedule.ScheduleASAP(downlink, toa)
	copy := &scheduledItem{
		time:    time.Add(-1 * (s.getGatewayRTT() + DefaultGatewayBufferTime)),
		payload: downlink,
	}
	s.queue.Add(copy)
	return nil
}

// Schedule an option
func (s *Schedule) Schedule(id string, downlink *router_pb.DownlinkMessage) error {
	s.Lock()
	defer s.Unlock()
	if !s.isActive() {
		return ErrScheduleInactive
	}
	if item, ok := s.items[id]; ok {
		item.mu.Lock()
		item.payload = downlink
		toa, _ := toa.Compute(downlink)
		item.duration = uint32(toa / 1000)
		if s.durationOffAirFunc != nil {
			item.duration += uint32(s.durationOffAirFunc(downlink) / 1000)
		}
		copy := &scheduledItem{
			id:      item.id,
			time:    item.time.Add(-1 * (s.getGatewayRTT() + DefaultGatewayBufferTime)),
			payload: item.payload,
		}
		s.ctx.WithField("Identifier", item.id).WithFields(fields.Get(item.payload)).WithField("Remaining", time.Until(copy.time)).Debug("Scheduled Downlink")
		item.mu.Unlock()
		delete(s.items, id)
		s.queue.Add(copy)
		return nil
	}
	return errors.NewErrNotFound(id)
}

// Subscribe to the downlink schedule
func (s *Schedule) Subscribe() <-chan *router_pb.DownlinkMessage {
	s.Lock()
	defer s.Unlock()
	if !s.isActive() {
		s.init()
		s.downlinkSubscribers++
	}
	return s.downlink
}

// Stop the downlink schedule
func (s *Schedule) Stop() {
	s.Lock()
	defer s.Unlock()
	s.downlinkSubscribers--
	if s.downlinkSubscribers < 1 {
		s.schedule.Destroy()
	}
}

// IsActive returns whether the schedule is active
func (s *Schedule) IsActive() bool {
	s.RLock()
	defer s.RUnlock()
	return s.isActive()
}

func (s *Schedule) isActive() bool {
	return s.downlinkSubscribers > 0
}

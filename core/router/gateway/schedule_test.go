package gateway

import (
	"testing"
	"time"

	router_pb "github.com/TheThingsNetwork/ttn/api/router"
	. "github.com/smartystreets/assertions"
)

const almostEqual = time.Millisecond

func TestScheduleSync(t *testing.T) {
	a := New(t)
	s := &schedule{}
	s.Sync(0)
	a.So(s.offset, ShouldAlmostEqual, time.Now().UnixNano(), almostEqual)

	s.Sync(1000)
	a.So(s.offset, ShouldAlmostEqual, time.Now().UnixNano()-1000*1000, almostEqual)
}

func TestScheduleRealtime(t *testing.T) {
	a := New(t)
	s := &schedule{}
	s.Sync(0)
	tm := s.realtime(10)
	a.So(tm.UnixNano(), ShouldAlmostEqual, time.Now().UnixNano()+10*1000, almostEqual)

	s.Sync(1000)
	tm = s.realtime(1010)
	a.So(tm.UnixNano(), ShouldAlmostEqual, time.Now().UnixNano()+10*1000, almostEqual)

	// Don't go back in time when uint32 overflows
	s.Sync(uintmax - 1)
	tm = s.realtime(10)
	a.So(tm.UnixNano(), ShouldAlmostEqual, time.Now().UnixNano()+9*1000, almostEqual)
}

func TestScheduleGetConflicts(t *testing.T) {
	a := New(t)

	// Test without overflow
	s := &schedule{
		queue: NewDownlinkQueue(
			&scheduledItem{timestamp: 5, length: 15},
			&scheduledItem{timestamp: 25, length: 10},
			&scheduledItem{timestamp: 55, length: 5},
			&scheduledItem{timestamp: 70, length: 20},
			&scheduledItem{timestamp: 95, length: 5},
			&scheduledItem{timestamp: 105, length: 10},
		),
	}
	a.So(s.getConflicts(0, 10), ShouldEqual, 1)
	a.So(s.getConflicts(15, 15), ShouldEqual, 2)
	a.So(s.getConflicts(40, 5), ShouldEqual, 0)
	a.So(s.getConflicts(50, 15), ShouldEqual, 1)
	a.So(s.getConflicts(75, 5), ShouldEqual, 1)
	a.So(s.getConflicts(85, 25), ShouldEqual, 3)

	// Test with overflow (already scheduled)
	s = &schedule{
		queue: NewDownlinkQueue(
			&scheduledItem{timestamp: 1<<32 - 1, length: 20},
		),
	}
	a.So(s.getConflicts(0, 20), ShouldEqual, 1)
	a.So(s.getConflicts(25, 5), ShouldEqual, 0)

	// Test with overflow (to schedule)
	s = &schedule{
		queue: NewDownlinkQueue(
			&scheduledItem{timestamp: 10, length: 20},
		),
	}
	a.So(s.getConflicts(1<<32-1, 5), ShouldEqual, 0)
	a.So(s.getConflicts(1<<32-1, 20), ShouldEqual, 1)
}

func TestScheduleGetOption(t *testing.T) {
	a := New(t)
	s := NewSchedule().(*schedule)

	s.Sync(0)
	_, conflicts := s.GetOption(100, 100)
	a.So(conflicts, ShouldEqual, 0)
	_, conflicts = s.GetOption(50, 100)
	a.So(conflicts, ShouldEqual, 1)
}

func TestScheduleSchedule(t *testing.T) {
	a := New(t)
	s := NewSchedule().(*schedule)

	s.Sync(0)

	err := s.Schedule("random", &router_pb.DownlinkMessage{})
	a.So(err, ShouldNotBeNil)

	id, conflicts := s.GetOption(100, 100)
	err = s.Schedule(id, &router_pb.DownlinkMessage{})
	a.So(err, ShouldBeNil)

	_, conflicts = s.GetOption(50, 100)
	a.So(conflicts, ShouldEqual, 10)
}

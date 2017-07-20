// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package gateway

import (
	"fmt"
	"testing"
	"time"

	router_pb "github.com/TheThingsNetwork/api/router"
	. "github.com/TheThingsNetwork/ttn/utils/testing"
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

func buildItems(items ...*scheduledItem) map[string]*scheduledItem {
	m := make(map[string]*scheduledItem)
	for idx, item := range items {
		m[fmt.Sprintf("%d", idx)] = item
	}
	return m
}

func TestScheduleGetConflicts(t *testing.T) {
	a := New(t)

	// Test without overflow
	s := &schedule{
		items: buildItems(
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
		items: buildItems(
			&scheduledItem{timestamp: 1<<32 - 1, length: 20},
		),
	}
	a.So(s.getConflicts(0, 20), ShouldEqual, 1)
	a.So(s.getConflicts(25, 5), ShouldEqual, 0)

	// Test with overflow (to schedule)
	s = &schedule{
		items: buildItems(
			&scheduledItem{timestamp: 10, length: 20},
		),
	}
	a.So(s.getConflicts(1<<32-1, 5), ShouldEqual, 0)
	a.So(s.getConflicts(1<<32-1, 20), ShouldEqual, 1)
}

func TestScheduleGetOption(t *testing.T) {
	a := New(t)
	s := NewSchedule(nil).(*schedule)

	s.Sync(0)
	_, conflicts := s.GetOption(100, 100)
	a.So(conflicts, ShouldEqual, 0)
	_, conflicts = s.GetOption(50, 100)
	a.So(conflicts, ShouldEqual, 1)
}

func TestScheduleSchedule(t *testing.T) {
	a := New(t)
	s := NewSchedule(GetLogger(t, "TestScheduleSchedule")).(*schedule)

	s.Sync(0)

	err := s.Schedule("random", &router_pb.DownlinkMessage{})
	a.So(err, ShouldNotBeNil)

	id, conflicts := s.GetOption(100, 100)
	err = s.Schedule(id, &router_pb.DownlinkMessage{})
	a.So(err, ShouldBeNil)

	_, conflicts = s.GetOption(50, 100)
	a.So(conflicts, ShouldEqual, 100)
}

func TestScheduleSubscribe(t *testing.T) {
	a := New(t)
	s := NewSchedule(GetLogger(t, "TestScheduleSubscribe")).(*schedule)
	s.Sync(0)
	Deadline = 1 * time.Millisecond // Very short deadline

	downlink1 := &router_pb.DownlinkMessage{Payload: []byte{1}}
	downlink2 := &router_pb.DownlinkMessage{Payload: []byte{2}}
	downlink3 := &router_pb.DownlinkMessage{Payload: []byte{3}}

	go func() {
		var i int
		for out := range s.Subscribe("") {
			switch i {
			case 0:
				a.So(out, ShouldEqual, downlink2)
			case 1:
				a.So(out, ShouldEqual, downlink1)
			case 3:
				a.So(out, ShouldEqual, downlink3)
			}
			i++
		}
	}()

	id, _ := s.GetOption(30000, 50)
	s.Schedule(id, downlink1)
	id, _ = s.GetOption(20000, 50)
	s.Schedule(id, downlink2)
	id, _ = s.GetOption(40000, 50)
	s.Schedule(id, downlink3)

	go func() {
		<-time.After(400 * time.Millisecond)
		s.Stop("")
	}()

	<-time.After(500 * time.Millisecond)

}

// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package gateway

import (
	"fmt"
	"sync"
	"testing"
	"time"

	router_pb "github.com/TheThingsNetwork/ttn/api/router"
	. "github.com/TheThingsNetwork/ttn/utils/testing"
	. "github.com/smartystreets/assertions"
)

const almostEqual = time.Millisecond

func TestScheduleSync(t *testing.T) {
	a := New(t)
	s := &Schedule{}

	s.Sync(0)
	now := time.Now()
	a.So(s.getTimestamp(now), ShouldBeBetweenOrEqual, 0, 1050)
	a.So(s.getRealtime(0), ShouldHappenWithin, time.Millisecond, time.Now())
	a.So(s.timestamp, ShouldEqual, 0)
	a.So(s.offset, ShouldAlmostEqual, time.Now().UnixNano(), almostEqual)
	a.So(s.getFullTimestamp(10), ShouldEqual, 10)
	a.So(s.getRealtime(10000), ShouldHappenWithin, time.Millisecond, time.Now().Add(10*time.Millisecond))

	s.Sync(20000)
	now = time.Now()
	a.So(s.getTimestamp(now), ShouldBeBetweenOrEqual, 20000, 20050)
	a.So(s.getTimestamp(now.Add(5000)), ShouldBeBetweenOrEqual, 20005, 20055)

	a.So(s.getRealtime(20000), ShouldHappenWithin, time.Millisecond, time.Now())
	a.So(s.timestamp, ShouldEqual, 20000)
	a.So(s.offset, ShouldAlmostEqual, time.Now().UnixNano()-20000*1000, almostEqual)
	a.So(s.getRealtime(30000), ShouldHappenWithin, time.Millisecond, time.Now().Add(10*time.Millisecond))

	s.Sync(1<<32 - 10)
	a.So(s.timestamp, ShouldEqual, uint64(1<<32-10))
	a.So(s.getFullTimestamp(10), ShouldEqual, uint64(1<<32+10))

	s.Sync(1000)
	a.So(s.timestamp, ShouldEqual, uint64(1<<32+1000))
}

func TestScheduleGetOption(t *testing.T) {
	a := New(t)

	// Test without overflow
	s := NewSchedule(GetLogger(t, "TestScheduleGetOption"), nil)

	id, conflicts, err := s.GetOption(1, 10)
	a.So(err, ShouldNotBeNil)

	subCh := s.Subscribe()
	go func() {
		for down := range subCh {
			fmt.Println("Down:", down)
		}
	}()

	s.Sync(0)
	const base = 1000000 // one second from now

	id, conflicts, err = s.GetOption(base+5000, 15000)
	a.So(id, ShouldNotBeEmpty)
	a.So(conflicts, ShouldEqual, 0)
	a.So(err, ShouldBeNil)
	a.So(s.items[id].time, ShouldHappenWithin, time.Millisecond, time.Now().Add(1005*time.Millisecond))

	id, conflicts, err = s.GetOption(base+25000, 10000)
	a.So(conflicts, ShouldEqual, 0)

	id, conflicts, err = s.GetOption(base+55000, 5000)
	a.So(conflicts, ShouldEqual, 0)

	id, conflicts, err = s.GetOption(base+70000, 20000)
	a.So(conflicts, ShouldEqual, 0)

	id, conflicts, err = s.GetOption(base+95000, 5000)
	a.So(conflicts, ShouldEqual, 0)

	id, conflicts, err = s.GetOption(base+105000, 10000)
	a.So(conflicts, ShouldEqual, 0)

	id, conflicts, err = s.GetOption(base, 10000)
	a.So(conflicts, ShouldEqual, 1)

	id, conflicts, err = s.GetOption(base+15000, 15000)
	a.So(conflicts, ShouldEqual, 2)

	id, conflicts, err = s.GetOption(base+85000, 25000)
	a.So(conflicts, ShouldEqual, 3)

	// Test with overflow
	s.Sync(maxUint32 - 1000000) // 1 second before maxUint32

	s.GetOption(maxUint32-10000, 20000)

	id, conflicts, err = s.GetOption(0, 20000)
	a.So(conflicts, ShouldEqual, 1)

	id, conflicts, err = s.GetOption(25000, 4000)
	a.So(conflicts, ShouldEqual, 0)

	id, conflicts, err = s.GetOption(maxUint32-20000, 10000)
	a.So(conflicts, ShouldEqual, 1)

	id, conflicts, err = s.GetOption(maxUint32-20000, 30000)
	a.So(conflicts, ShouldEqual, 3)

	s.Stop()
}

func TestScheduleSchedule(t *testing.T) {
	a := New(t)
	s := NewSchedule(GetLogger(t, "TestScheduleSchedule"), nil)

	s.Sync(0)
	DefaultGatewayRTT = 20 * time.Millisecond
	DefaultGatewayBufferTime = 10 * time.Millisecond

	var wg sync.WaitGroup

	downlink := &router_pb.DownlinkMessage{Payload: []byte{1, 2, 3, 4}}

	var now time.Time
	subCh := s.Subscribe()
	wg.Add(1)
	go func() {
		a.So(<-subCh, ShouldEqual, downlink)
		a.So(time.Now(), ShouldHappenWithin, 10*time.Millisecond, now.Add(70*time.Millisecond))
		wg.Done()
	}()

	err := s.Schedule("random", downlink)
	a.So(err, ShouldNotBeNil)

	now = time.Now()
	id, _, _ := s.GetOption(100000, 100000)
	err = s.Schedule(id, downlink)
	a.So(err, ShouldBeNil)

	_, _, err = s.GetOption(50000, 100000)
	a.So(err, ShouldNotBeNil)

	wg.Wait()
}

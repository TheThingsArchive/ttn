package gateway

import (
	"testing"
	"time"

	. "github.com/smartystreets/assertions"
)

func TestScheduleDatastructure(t *testing.T) {
	a := New(t)
	dq := NewDownlinkQueue()

	a.So(dq.Peek(), ShouldBeNil)
	a.So(dq.Pop(), ShouldBeNil)

	now := time.Now()

	i1 := &scheduledItem{time: now.Add(300 * time.Millisecond)}
	i2 := &scheduledItem{time: now.Add(200 * time.Millisecond)}
	i3 := &scheduledItem{time: now.Add(100 * time.Millisecond)}
	i4 := &scheduledItem{time: now.Add(250 * time.Millisecond)}
	i5 := &scheduledItem{time: now.Add(50 * time.Millisecond)}

	dq.Push(i1)
	a.So(dq.Peek(), ShouldEqual, i1)
	dq.Push(i2)
	a.So(dq.Peek(), ShouldEqual, i2)
	dq.Push(i3)
	a.So(dq.Peek(), ShouldEqual, i3)
	dq.Push(i4)
	a.So(dq.Peek(), ShouldEqual, i3)
	a.So(dq.Pop(), ShouldEqual, i3)
	a.So(dq.Peek(), ShouldEqual, i2)
	dq.Push(i5)
	a.So(dq.Peek(), ShouldEqual, i5)

}

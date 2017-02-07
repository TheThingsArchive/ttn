// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package device

import (
	"testing"

	"github.com/TheThingsNetwork/ttn/core/types"
	. "github.com/TheThingsNetwork/ttn/utils/testing"
	. "github.com/smartystreets/assertions"
)

func TestDownlinkQueue(t *testing.T) {
	a := New(t)

	store := NewRedisDeviceStore(GetRedisClient(), "handler-test-downlink-queue")
	s, _ := store.DownlinkQueue("test", "test")

	defer func() {
		store.Delete("test", "test")
	}()

	{
		length, err := s.Length()
		a.So(err, ShouldBeNil)
		a.So(length, ShouldEqual, 0)
	}

	{
		next, err := s.Next()
		a.So(err, ShouldBeNil)
		a.So(next, ShouldBeNil)
	}

	{
		err := s.PushLast(&types.DownlinkMessage{
			PayloadRaw: []byte{0x12, 0x34},
		})
		a.So(err, ShouldBeNil)
	}

	{
		err := s.PushFirst(&types.DownlinkMessage{
			PayloadRaw: []byte{0xab, 0xcd},
		})
		a.So(err, ShouldBeNil)
	}

	{
		length, err := s.Length()
		a.So(err, ShouldBeNil)
		a.So(length, ShouldEqual, 2)
	}

	{
		next, err := s.Next()
		a.So(err, ShouldBeNil)
		a.So(next, ShouldNotBeNil)
		a.So(next.PayloadRaw, ShouldResemble, []byte{0xab, 0xcd})
	}

	{
		err := s.Replace(&types.DownlinkMessage{
			PayloadRaw: []byte{0xaa, 0xbc},
		})
		a.So(err, ShouldBeNil)
	}

	{
		length, err := s.Length()
		a.So(err, ShouldBeNil)
		a.So(length, ShouldEqual, 1)
	}

	{
		next, err := s.Next()
		a.So(err, ShouldBeNil)
		a.So(next, ShouldNotBeNil)
		a.So(next.PayloadRaw, ShouldResemble, []byte{0xaa, 0xbc})
	}

}

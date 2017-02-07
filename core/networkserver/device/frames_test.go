// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package device

import (
	"testing"

	"github.com/TheThingsNetwork/ttn/core/types"
	. "github.com/TheThingsNetwork/ttn/utils/testing"
	. "github.com/smartystreets/assertions"
)

func TestFramesStore(t *testing.T) {
	a := New(t)
	store := NewRedisDeviceStore(GetRedisClient(), "networkserver-test-frames-store")

	appEUI := types.AppEUI{0, 0, 0, 0, 0, 0, 0, 1}
	devEUI := types.DevEUI{0, 0, 0, 0, 0, 0, 0, 1}

	s, err := store.Frames(appEUI, devEUI)
	a.So(err, ShouldBeNil)

	defer s.Clear()

	{
		err := s.Push(&Frame{
			SNR:          -10.5,
			GatewayCount: 2,
		})
		a.So(err, ShouldBeNil)
	}

	{
		frames, err := s.Get()
		a.So(err, ShouldBeNil)
		a.So(frames, ShouldHaveLength, 1)
		a.So(frames[0].SNR, ShouldEqual, -10.5)
		a.So(frames[0].GatewayCount, ShouldEqual, 2)
	}

	{
		err := s.Clear()
		a.So(err, ShouldBeNil)
	}

	{
		frames, err := s.Get()
		a.So(err, ShouldBeNil)
		a.So(frames, ShouldBeEmpty)
	}

	{
		defer s.Clear()
		for i := 0; i < 25; i++ {
			s.Push(&Frame{
				GatewayCount: uint32(i + 1),
			})
		}
		{
			frames, err := s.Get()
			a.So(err, ShouldBeNil)
			a.So(frames, ShouldHaveLength, 20)
			a.So(frames[0].GatewayCount, ShouldEqual, 25)
		}

	}

}

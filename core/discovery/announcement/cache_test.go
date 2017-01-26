// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package announcement

import (
	"testing"

	"github.com/TheThingsNetwork/ttn/core/types"
	. "github.com/TheThingsNetwork/ttn/utils/testing"
	. "github.com/smartystreets/assertions"
)

func TestCachedAnnouncementStore(t *testing.T) {
	a := New(t)

	s := NewRedisAnnouncementStore(GetRedisClient(), "discovery-test-announcement-store")

	s = NewCachedAnnouncementStore(s, DefaultCacheOptions)

	// Get non-existing
	dev, err := s.Get("router", "router1")
	a.So(err, ShouldNotBeNil)
	a.So(dev, ShouldBeNil)

	// Create
	err = s.Set(&Announcement{
		ServiceName: "router",
		ID:          "router1",
	})
	a.So(err, ShouldBeNil)

	defer func() {
		s.Delete("router", "router1")
	}()

	// Get existing
	dev, err = s.Get("router", "router1")
	a.So(err, ShouldBeNil)
	a.So(dev, ShouldNotBeNil)

	// Create extra
	err = s.Set(&Announcement{
		ServiceName: "handler",
		ID:          "handler1",
	})
	a.So(err, ShouldBeNil)

	defer func() {
		s.Delete("handler", "handler1")
	}()

	err = s.Set(&Announcement{
		ServiceName: "handler",
		ID:          "handler2",
	})
	a.So(err, ShouldBeNil)

	defer func() {
		s.Delete("handler", "handler2")
	}()

	appEUI := types.AppEUI([8]byte{1, 2, 3, 4, 5, 6, 7, 8})

	err = s.AddMetadata("handler", "handler1",
		AppEUIMetadata{AppEUI: appEUI},
		AppIDMetadata{AppID: "AppID"},
	)
	a.So(err, ShouldBeNil)

	handler, err := s.GetForAppEUI(appEUI)
	a.So(err, ShouldBeNil)
	a.So(handler, ShouldNotBeNil)
	a.So(handler.ID, ShouldEqual, "handler1")

	handler, err = s.GetForAppID("AppID")
	a.So(err, ShouldBeNil)
	a.So(handler, ShouldNotBeNil)
	a.So(handler.ID, ShouldEqual, "handler1")

	err = s.AddMetadata("handler", "handler2",
		AppEUIMetadata{AppEUI: appEUI},
		AppIDMetadata{AppID: "AppID"},
		AppIDMetadata{AppID: "OtherAppID"},
	)
	a.So(err, ShouldBeNil)

	metadata, err := s.GetMetadata("handler", "handler2")
	a.So(err, ShouldBeNil)
	a.So(metadata, ShouldHaveLength, 3)

	err = s.AddMetadata("handler", "handler2",
		AppEUIMetadata{AppEUI: appEUI},
		AppIDMetadata{AppID: "AppID"},
	)
	a.So(err, ShouldBeNil)

	metadata, err = s.GetMetadata("handler", "handler2")
	a.So(err, ShouldBeNil)
	a.So(metadata, ShouldHaveLength, 3)

	handler, err = s.GetForAppEUI(appEUI)
	a.So(err, ShouldBeNil)
	a.So(handler, ShouldNotBeNil)
	a.So(handler.ID, ShouldEqual, "handler2")

	handler, err = s.GetForAppID("AppID")
	a.So(err, ShouldBeNil)
	a.So(handler, ShouldNotBeNil)
	a.So(handler.ID, ShouldEqual, "handler2")

	err = s.RemoveMetadata("handler", "handler1",
		AppEUIMetadata{AppEUI: appEUI},
		AppIDMetadata{AppID: "AppID"},
	)
	a.So(err, ShouldBeNil)

	err = s.RemoveMetadata("handler", "handler2",
		AppEUIMetadata{AppEUI: appEUI},
		AppIDMetadata{AppID: "AppID"},
	)
	a.So(err, ShouldBeNil)

	// List
	announcements, err := s.List(nil)
	a.So(err, ShouldBeNil)
	a.So(announcements, ShouldHaveLength, 3)

	// List
	announcements, err = s.ListService("router", nil)
	a.So(err, ShouldBeNil)
	a.So(announcements, ShouldHaveLength, 1)

	// Delete
	err = s.Delete("router", "router1")
	a.So(err, ShouldBeNil)

	// Get deleted
	dev, err = s.Get("router", "router1")
	a.So(err, ShouldNotBeNil)
	a.So(dev, ShouldBeNil)

	// Delete with Metadata
	err = s.Delete("handler", "handler2")
	a.So(err, ShouldBeNil)
}

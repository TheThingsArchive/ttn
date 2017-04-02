// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package application

import (
	"testing"

	. "github.com/TheThingsNetwork/ttn/utils/testing"
	. "github.com/smartystreets/assertions"
)

func TestApplicationStore(t *testing.T) {
	a := New(t)

	NewRedisApplicationStore(GetRedisClient(), "")

	s := NewRedisApplicationStore(GetRedisClient(), "handler-test-application-store")

	appID := "AppID-1"

	// Get non-existing
	app, err := s.Get(appID)
	a.So(err, ShouldNotBeNil)
	a.So(app, ShouldBeNil)

	// Create
	app = &Application{
		AppID:         appID,
		CustomEncoder: "custom encoder",
	}
	err = s.Set(app)
	defer func() {
		s.Delete(appID)
	}()
	a.So(err, ShouldBeNil)

	// Get existing
	app, err = s.Get(appID)
	a.So(err, ShouldBeNil)
	a.So(app, ShouldNotBeNil)
	a.So(app.CustomEncoder, ShouldEqual, "custom encoder")

	// Update
	err = s.Set(&Application{
		old:           app,
		AppID:         appID,
		CustomEncoder: "new encoder",
	})
	a.So(err, ShouldBeNil)

	// Get existing
	app, err = s.Get(appID)
	a.So(err, ShouldBeNil)
	a.So(app, ShouldNotBeNil)
	a.So(app.CustomEncoder, ShouldEqual, "new encoder")

	// List
	apps, err := s.List(nil)
	a.So(err, ShouldBeNil)
	a.So(apps, ShouldHaveLength, 1)

	// Delete
	err = s.Delete(appID)
	a.So(err, ShouldBeNil)

	// Get deleted
	app, err = s.Get(appID)
	a.So(err, ShouldNotBeNil)
	a.So(app, ShouldBeNil)
}

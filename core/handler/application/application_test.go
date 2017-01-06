// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package application

import (
	"testing"

	. "github.com/smartystreets/assertions"
)

func TestApplicationUpdate(t *testing.T) {
	a := New(t)
	application := &Application{
		AppID: "App",
	}
	application.StartUpdate()
	a.So(application.old.AppID, ShouldEqual, application.AppID)
}

func TestApplicationChangedFields(t *testing.T) {
	a := New(t)
	application := &Application{
		AppID: "Application",
	}
	application.StartUpdate()
	application.AppID = "NewAppID"

	a.So(application.ChangedFields(), ShouldHaveLength, 1)
	a.So(application.ChangedFields(), ShouldContain, "AppID")
}

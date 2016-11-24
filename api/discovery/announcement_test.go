// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package discovery

import (
	"testing"

	. "github.com/smartystreets/assertions"
)

func TestAnnouncementAddDeleteMetadata(t *testing.T) {
	a := New(t)
	announcement := new(Announcement)

	announcement.AddMetadata(Metadata_APP_ID, []byte("app-id"))
	a.So(announcement.Metadata, ShouldHaveLength, 1)
	a.So(announcement.Metadata[0], ShouldResemble, &Metadata{Key: Metadata_APP_ID, Value: []byte("app-id")})

	announcement.AddMetadata(Metadata_APP_ID, []byte("app-id"))
	a.So(announcement.Metadata, ShouldHaveLength, 1)

	announcement.AddMetadata(Metadata_APP_ID, []byte("other-app-id"))
	a.So(announcement.Metadata, ShouldHaveLength, 2)

	announcement.DeleteMetadata(Metadata_APP_ID, []byte("app-id"))
	a.So(announcement.Metadata, ShouldHaveLength, 1)

}

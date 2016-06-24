// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package discovery

import (
	"testing"

	. "github.com/smartystreets/assertions"
)

func getTestAnnouncement() (announcement *Announcement, dmap map[string]string) {
	return &Announcement{
			Id:             "abcdef",
			Token:          "ghijkl",
			Description:    "Test Description",
			ServiceName:    "router",
			ServiceVersion: "1.0-preview build abcdef",
			NetAddress:     "localhost:1234",
			Metadata: []*Metadata{
				&Metadata{
					Key:   Metadata_PREFIX,
					Value: []byte("38"),
				},
				&Metadata{
					Key:   Metadata_PREFIX,
					Value: []byte("39"),
				},
			},
		}, map[string]string{
			"id":              "abcdef",
			"token":           "ghijkl",
			"description":     "Test Description",
			"service_name":    "router",
			"service_version": "1.0-preview build abcdef",
			"net_address":     "localhost:1234",
			"metadata":        `[{"key":1,"value":"Mzg="},{"key":1,"value":"Mzk="}]`,
		}
}

func TestToStringMap(t *testing.T) {
	a := New(t)
	announcement, expected := getTestAnnouncement()
	dmap, err := announcement.ToStringStringMap(AnnouncementProperties...)
	a.So(err, ShouldBeNil)
	a.So(dmap, ShouldResemble, expected)
}

func TestFromStringMap(t *testing.T) {
	a := New(t)
	announcement := &Announcement{}
	expected, dmap := getTestAnnouncement()
	err := announcement.FromStringStringMap(dmap)
	a.So(err, ShouldBeNil)
	a.So(announcement, ShouldResemble, expected)
}

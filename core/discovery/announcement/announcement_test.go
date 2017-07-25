// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package announcement

import (
	"testing"

	pb "github.com/TheThingsNetwork/api/discovery"
	"github.com/TheThingsNetwork/ttn/core/types"
	. "github.com/smartystreets/assertions"
)

func TestMetadataTextMarshaling(t *testing.T) {
	a := New(t)
	subjects := map[string]Metadata{
		"AppEUI 0102030405060708": AppEUIMetadata{types.AppEUI([8]byte{1, 2, 3, 4, 5, 6, 7, 8})},
		"AppID AppID":             AppIDMetadata{"AppID"},
		"Prefix 00000000/0":       PrefixMetadata{types.DevAddrPrefix{}},
	}

	for str, obj := range subjects {
		marshaled, err := obj.MarshalText()
		a.So(err, ShouldBeNil)
		a.So(string(marshaled), ShouldEqual, str)
		unmarshaled := MetadataFromString(str)
		a.So(unmarshaled, ShouldResemble, obj)
	}
}

func TestAnnouncementUpdate(t *testing.T) {
	a := New(t)
	announcement := &Announcement{
		ID: "ID",
	}
	announcement.StartUpdate()
	a.So(announcement.old.ID, ShouldEqual, announcement.ID)
}

func TestAnnouncementChangedFields(t *testing.T) {
	a := New(t)
	announcement := &Announcement{
		ID: "ID",
	}
	announcement.StartUpdate()
	announcement.ID = "ID2"

	a.So(announcement.ChangedFields(), ShouldHaveLength, 1)
	a.So(announcement.ChangedFields(), ShouldContain, "ID")
}

func TestAnnouncementToProto(t *testing.T) {
	a := New(t)
	announcement := &Announcement{
		ID: "ID",
		Metadata: []Metadata{
			AppEUIMetadata{types.AppEUI([8]byte{1, 2, 3, 4, 5, 6, 7, 8})},
			AppIDMetadata{"AppID"},
			PrefixMetadata{types.DevAddrPrefix{}},
		},
	}
	proto := announcement.ToProto()
	a.So(proto.ID, ShouldEqual, announcement.ID)
	a.So(proto.Metadata, ShouldHaveLength, 3)
	a.So(proto.Metadata[0].GetAppEUI(), ShouldResemble, []byte{1, 2, 3, 4, 5, 6, 7, 8})
	a.So(proto.Metadata[1].GetAppID(), ShouldEqual, "AppID")
	a.So(proto.Metadata[2].GetDevAddrPrefix(), ShouldResemble, []byte{0, 0, 0, 0, 0})
}

func TestAnnouncementFromProto(t *testing.T) {
	a := New(t)
	proto := &pb.Announcement{
		ID: "ID",
		Metadata: []*pb.Metadata{
			&pb.Metadata{Metadata: &pb.Metadata_AppEUI{AppEUI: []byte{1, 2, 3, 4, 5, 6, 7, 8}}},
			&pb.Metadata{Metadata: &pb.Metadata_AppID{AppID: "AppID"}},
			&pb.Metadata{Metadata: &pb.Metadata_DevAddrPrefix{DevAddrPrefix: []byte{0, 0, 0, 0, 0}}},
		},
	}
	announcement := FromProto(proto)
	a.So(announcement.ID, ShouldEqual, proto.ID)
	a.So(announcement.Metadata, ShouldHaveLength, 3)
}

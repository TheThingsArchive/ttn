// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package announcement

import (
	"testing"

	pb "github.com/TheThingsNetwork/ttn/api/discovery"
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
			OtherMetadata{},
		},
	}
	proto := announcement.ToProto()
	a.So(proto.Id, ShouldEqual, announcement.ID)
	a.So(proto.Metadata, ShouldHaveLength, 4)
	a.So(proto.Metadata[0].Key, ShouldEqual, pb.Metadata_APP_EUI)
	a.So(proto.Metadata[1].Key, ShouldEqual, pb.Metadata_APP_ID)
	a.So(proto.Metadata[2].Key, ShouldEqual, pb.Metadata_PREFIX)
	a.So(proto.Metadata[3].Key, ShouldEqual, pb.Metadata_OTHER)
}

func TestAnnouncementFromProto(t *testing.T) {
	a := New(t)
	proto := &pb.Announcement{
		Id: "ID",
		Metadata: []*pb.Metadata{
			&pb.Metadata{Key: pb.Metadata_APP_EUI, Value: []byte{1, 2, 3, 4, 5, 6, 7, 8}},
			&pb.Metadata{Key: pb.Metadata_APP_ID, Value: []byte("AppID")},
			&pb.Metadata{Key: pb.Metadata_PREFIX, Value: []byte{0, 0, 0, 0, 0}},
			&pb.Metadata{Key: pb.Metadata_OTHER, Value: []byte{}},
		},
	}
	announcement := FromProto(proto)
	a.So(announcement.ID, ShouldEqual, proto.Id)
	a.So(announcement.Metadata, ShouldHaveLength, 4)
}

// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package discovery

import (
	"testing"

	pb "github.com/TheThingsNetwork/ttn/api/discovery"
	"github.com/TheThingsNetwork/ttn/core/discovery/announcement"
	"github.com/TheThingsNetwork/ttn/core/types"
	. "github.com/smartystreets/assertions"
)

func TestIntegrationBrokerDiscovery(t *testing.T) {
	a := New(t)

	port := randomPort()

	discoveryServer, s := buildTestDiscoveryServer(port)
	defer s.Stop()

	discoveryServer.services = announcement.NewAnnouncementStore()
	discoveryServer.services.Set(&pb.Announcement{
		Id:          "broker1",
		ServiceName: "broker",
		Metadata: []*pb.Metadata{
			&pb.Metadata{Key: pb.Metadata_PREFIX, Value: []byte{8, 0x01, 0x00, 0x00, 0x00}},
		},
	})
	discoveryServer.services.Set(&pb.Announcement{
		Id:          "broker2",
		ServiceName: "broker",
		Metadata: []*pb.Metadata{
			&pb.Metadata{Key: pb.Metadata_PREFIX, Value: []byte{8, 0x02, 0x00, 0x00, 0x00}},
		},
	})

	discoveryClient := buildTestBrokerDiscoveryClient(port)

	brokers, err := discoveryClient.Discover(types.DevAddr{0x01, 0x02, 0x03, 0x04})
	a.So(err, ShouldBeNil)
	a.So(brokers, ShouldHaveLength, 1)
}

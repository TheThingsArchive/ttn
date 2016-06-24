// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package discovery

import (
	"testing"

	pb "github.com/TheThingsNetwork/ttn/api/discovery"
	"github.com/TheThingsNetwork/ttn/core/types"
	. "github.com/smartystreets/assertions"
)

func TestIntegrationBrokerDiscovery(t *testing.T) {
	a := New(t)

	port := randomPort()

	discoveryServer, s := buildTestDiscoveryServer(port)
	defer s.Stop()

	discoveryServer.services = map[string]map[string]*pb.Announcement{
		"broker": map[string]*pb.Announcement{
			"broker1": &pb.Announcement{Metadata: []*pb.Metadata{
				&pb.Metadata{Key: pb.Metadata_PREFIX, Value: []byte{8, 0x01, 0x00, 0x00, 0x00}},
			}},
			"broker2": &pb.Announcement{Metadata: []*pb.Metadata{
				&pb.Metadata{Key: pb.Metadata_PREFIX, Value: []byte{8, 0x02, 0x00, 0x00, 0x00}},
			}},
		},
		"other": map[string]*pb.Announcement{
			"other": &pb.Announcement{},
		},
	}

	discoveryClient := buildTestBrokerDiscoveryClient(port)

	brokers, err := discoveryClient.Discover(types.DevAddr{0x01, 0x02, 0x03, 0x04})
	a.So(err, ShouldBeNil)
	a.So(brokers, ShouldHaveLength, 1)
}

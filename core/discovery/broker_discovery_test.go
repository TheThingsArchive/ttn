// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package discovery

import (
	"fmt"
	"testing"
	"time"

	"google.golang.org/grpc"

	"github.com/TheThingsNetwork/ttn/api"
	pb "github.com/TheThingsNetwork/ttn/api/discovery"
	"github.com/TheThingsNetwork/ttn/core"
	"github.com/TheThingsNetwork/ttn/core/types"
	. "github.com/smartystreets/assertions"
)

func buildTestBrokerDiscoveryClient(port uint) *brokerDiscovery {
	conn, err := grpc.Dial(fmt.Sprintf("localhost:%d", port), append(api.DialOptions, grpc.WithInsecure(), grpc.WithBlock())...)
	if err != nil {
		panic(err)
	}
	client := pb.NewDiscoveryClient(conn)
	discovery := NewBrokerDiscovery(&core.Component{Discovery: client}).(*brokerDiscovery)
	discovery.refreshCache()
	return discovery
}

func TestBrokerDiscovery(t *testing.T) {
	a := New(t)

	// Broker1 has a prefix with all DevAddrs
	broker1 := &pb.Announcement{ServiceName: "broker", NetAddress: "localhost1:1881",
		Metadata: []*pb.Metadata{
			&pb.Metadata{Key: pb.Metadata_PREFIX, Value: []byte{0, 0x00, 0x00, 0x00, 0x00}},
		},
	}

	// Broker2 has one DevAddr prefix
	broker2 := &pb.Announcement{ServiceName: "broker", NetAddress: "localhost2:1881",
		Metadata: []*pb.Metadata{
			&pb.Metadata{Key: pb.Metadata_PREFIX, Value: []byte{8, 0x01, 0x00, 0x00, 0x00}},
		},
	}

	// Broker3 has multiple DevAddr prefixes
	broker3 := &pb.Announcement{ServiceName: "broker", NetAddress: "localhost3:1881",
		Metadata: []*pb.Metadata{
			&pb.Metadata{Key: pb.Metadata_PREFIX, Value: []byte{16, 0x02, 0x03, 0x00, 0x00}},
		},
	}

	d := &brokerDiscovery{
		cacheValidUntil: time.Now().Add(10 * time.Minute),
		cache:           []*pb.Announcement{broker1, broker2, broker3},
	}

	results, err := d.All()
	a.So(err, ShouldBeNil)
	a.So(results, ShouldNotBeEmpty)
	a.So(results, ShouldContain, broker1)
	a.So(results, ShouldContain, broker2)
	a.So(results, ShouldContain, broker3)

	results, err = d.Discover(types.DevAddr{0x01, 0x02, 0x03, 0x04})
	a.So(err, ShouldBeNil)
	a.So(results, ShouldNotBeEmpty)
	a.So(results, ShouldContain, broker1)
	a.So(results, ShouldContain, broker2)
	a.So(results, ShouldNotContain, broker3)

	results, err = d.Discover(types.DevAddr{0x02, 0x03, 0x04, 0x05})
	a.So(err, ShouldBeNil)
	a.So(results, ShouldNotBeEmpty)
	a.So(results, ShouldContain, broker1)
	a.So(results, ShouldNotContain, broker2)
	a.So(results, ShouldContain, broker3)

	results, err = d.Discover(types.DevAddr{0x04, 0x05, 0x06, 0x07})
	a.So(err, ShouldBeNil)
	a.So(results, ShouldNotBeEmpty)
	a.So(results, ShouldContain, broker1)
	a.So(results, ShouldNotContain, broker2)
	a.So(results, ShouldNotContain, broker3)
}

func TestBrokerDiscoveryCache(t *testing.T) {
	a := New(t)

	port := randomPort()

	discoveryServer, _ := buildMockDiscoveryServer(port)

	broker := &pb.Announcement{ServiceName: "broker", NetAddress: "localhost1:1881",
		Metadata: []*pb.Metadata{&pb.Metadata{Key: pb.Metadata_PREFIX, Value: []byte{0x00, 0x00, 0x00, 0x00, 0x00}}},
	}

	conn, err := grpc.Dial(fmt.Sprintf("localhost:%d", port), append(api.DialOptions, grpc.WithInsecure(), grpc.WithBlock())...)
	if err != nil {
		panic(err)
	}
	client := pb.NewDiscoveryClient(conn)

	d := &brokerDiscovery{
		component: &core.Component{
			Discovery: client,
		},
		cacheValidUntil: time.Now().Add(-1 * time.Minute),
		cache:           []*pb.Announcement{broker},
	}

	// It should return the cached broker initially
	results, err := d.Discover(types.DevAddr{0x01, 0x02, 0x03, 0x04})
	a.So(err, ShouldBeNil)
	a.So(results, ShouldContain, broker)

	// It should still return the cached broker
	results, err = d.Discover(types.DevAddr{0x01, 0x02, 0x03, 0x04})
	a.So(err, ShouldBeNil)
	a.So(results, ShouldContain, broker)

	<-time.After(20 * time.Millisecond)

	// It should return the refreshed (empty) broker list
	results, err = d.Discover(types.DevAddr{0x01, 0x02, 0x03, 0x04})
	a.So(err, ShouldBeNil)
	a.So(results, ShouldBeEmpty)

	a.So(discoveryServer.discover, ShouldEqual, 1)
}

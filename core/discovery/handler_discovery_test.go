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
	. "github.com/smartystreets/assertions"
)

func buildTestHandlerDiscoveryClient(port uint) *handlerDiscovery {
	conn, err := grpc.Dial(fmt.Sprintf("localhost:%d", port), append(api.DialOptions, grpc.WithBlock())...)
	if err != nil {
		panic(err)
	}
	client := pb.NewDiscoveryClient(conn)
	discovery := NewHandlerDiscovery(&core.Component{Discovery: client}).(*handlerDiscovery)
	discovery.refreshCache()
	return discovery
}

func TestHandlerDiscovery(t *testing.T) {
	a := New(t)

	// Handler1 owns one AppEUI
	handler1 := &pb.Announcement{ServiceName: "handler", Id: "handler1", NetAddress: "localhost1:1881",
		Metadata: []*pb.Metadata{
			&pb.Metadata{Key: pb.Metadata_APP_ID, Value: []byte("AppID-2")},
		},
	}

	// Handler2 has two AppEUIs
	handler2 := &pb.Announcement{ServiceName: "handler", Id: "handler2", NetAddress: "localhost2:1881",
		Metadata: []*pb.Metadata{
			&pb.Metadata{Key: pb.Metadata_APP_ID, Value: []byte("AppID-1")},
		},
	}

	d := &handlerDiscovery{
		cacheValidUntil: time.Now().Add(10 * time.Minute),
		cache:           []*pb.Announcement{handler1, handler2},
	}
	d.updateLookups()

	announcement, err := d.Get("handler1")
	a.So(err, ShouldBeNil)
	a.So(announcement, ShouldEqual, handler1)

	results, err := d.All()
	a.So(err, ShouldBeNil)
	a.So(results, ShouldNotBeEmpty)
	a.So(results, ShouldContain, handler1)
	a.So(results, ShouldContain, handler2)

	results, err = d.ForAppID("AppID-1")
	a.So(err, ShouldBeNil)
	a.So(results, ShouldNotBeEmpty)
	a.So(results, ShouldNotContain, handler1)
	a.So(results, ShouldContain, handler2)

	results, err = d.ForAppID("AppID-2")
	a.So(err, ShouldBeNil)
	a.So(results, ShouldNotBeEmpty)
	a.So(results, ShouldContain, handler1)
	a.So(results, ShouldNotContain, handler2)

	results, err = d.ForAppID("AppID-3")
	a.So(err, ShouldBeNil)
	a.So(results, ShouldBeEmpty)
}

func TestHandlerDiscoveryCache(t *testing.T) {
	a := New(t)

	port := randomPort()

	discoveryServer, _ := buildMockDiscoveryServer(port)

	handler := &pb.Announcement{ServiceName: "handler", NetAddress: "localhost1:1881",
		Metadata: []*pb.Metadata{&pb.Metadata{Key: pb.Metadata_APP_ID, Value: []byte("AppID-1")}},
	}

	conn, err := grpc.Dial(fmt.Sprintf("localhost:%d", port), append(api.DialOptions, grpc.WithBlock())...)
	if err != nil {
		panic(err)
	}
	client := pb.NewDiscoveryClient(conn)

	d := &handlerDiscovery{
		component: &core.Component{
			Discovery: client,
		},
		cacheValidUntil: time.Now().Add(-1 * time.Minute),
		cache:           []*pb.Announcement{handler},
	}
	d.updateLookups()

	// It should return the cached handler initially
	results, err := d.ForAppID("AppID-1")
	a.So(err, ShouldBeNil)
	a.So(results, ShouldContain, handler)

	// It should still return the cached handler
	results, err = d.ForAppID("AppID-1")
	a.So(err, ShouldBeNil)
	a.So(results, ShouldContain, handler)

	<-time.After(20 * time.Millisecond)

	// It should return the refreshed (empty) handler list
	results, err = d.ForAppID("AppID-1")
	a.So(err, ShouldBeNil)
	a.So(results, ShouldBeEmpty)

	a.So(discoveryServer.discover, ShouldEqual, 1)
}

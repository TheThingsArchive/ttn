// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package broker

import (
	"fmt"
	"math/rand"
	"net"
	"testing"
	"time"

	"github.com/TheThingsNetwork/ttn/api"
	pb "github.com/TheThingsNetwork/ttn/api/broker"
	"github.com/TheThingsNetwork/ttn/core"
	"github.com/TheThingsNetwork/ttn/core/types"
	. "github.com/TheThingsNetwork/ttn/utils/testing"
	. "github.com/smartystreets/assertions"
	"golang.org/x/net/context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func randomPort() uint {
	rand.Seed(time.Now().UnixNano())
	port := rand.Intn(5000) + 5000
	return uint(port)
}

func buildTestBrokerServer(t *testing.T, port uint) (*broker, *grpc.Server) {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		panic(err)
	}
	b := &broker{
		Component: &core.Component{
			Ctx: GetLogger(t, "TestBrokerServer"),
		},
		routers:                make(map[string]chan *pb.DownlinkMessage),
		handlers:               make(map[string]chan *pb.DeduplicatedUplinkMessage),
		ns:                     &mockNetworkServer{},
		uplinkDeduplicator:     NewDeduplicator(300 * time.Millisecond),
		activationDeduplicator: NewDeduplicator(1000 * time.Millisecond),
	}
	s := grpc.NewServer()
	b.RegisterRPC(s)
	go s.Serve(lis)
	return b, s
}

func TestAssociateRPC(t *testing.T) {
	a := New(t)

	port := randomPort()
	b, s := buildTestBrokerServer(t, port)
	defer s.Stop()

	conn, err := grpc.Dial(fmt.Sprintf("localhost:%d", port), api.DialOptions...)
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	client := pb.NewBrokerClient(conn)
	md := metadata.Pairs(
		"token", "token",
		"id", "RouterID",
	)
	ctx := metadata.NewContext(context.Background(), md)

	stream, err := client.Associate(ctx)
	a.So(err, ShouldBeNil)

	<-time.After(5 * time.Millisecond)

	a.So(b.routers, ShouldNotBeEmpty)

	err = stream.CloseSend()
	a.So(err, ShouldBeNil)

	<-time.After(5 * time.Millisecond)

	a.So(b.routers, ShouldBeEmpty)

}

func TestSubscribeRPC(t *testing.T) {
	a := New(t)

	port := randomPort()
	b, s := buildTestBrokerServer(t, port)
	defer s.Stop()

	conn, err := grpc.Dial(fmt.Sprintf("localhost:%d", port), api.DialOptions...)
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	client := pb.NewBrokerClient(conn)
	md := metadata.Pairs(
		"token", "token",
		"id", "HandlerID",
	)
	ctx := metadata.NewContext(context.Background(), md)

	stream, err := client.Subscribe(ctx, &pb.SubscribeRequest{})
	a.So(err, ShouldBeNil)

	<-time.After(5 * time.Millisecond)

	a.So(b.handlers, ShouldNotBeEmpty)

	err = stream.CloseSend()
	a.So(err, ShouldBeNil)

	err = conn.Close()
	a.So(err, ShouldBeNil)

	<-time.After(5 * time.Millisecond)

	a.So(b.handlers, ShouldBeEmpty)

}

func TestPublishRPC(t *testing.T) {
	a := New(t)

	appEUI := types.AppEUI{0, 1, 2, 3, 4, 5, 6, 7}
	devEUI := types.DevEUI{0, 1, 2, 3, 4, 5, 6, 7}

	port := randomPort()
	b, s := buildTestBrokerServer(t, port)
	defer s.Stop()

	conn, err := grpc.Dial(fmt.Sprintf("localhost:%d", port), api.DialOptions...)
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	client := pb.NewBrokerClient(conn)
	md := metadata.Pairs(
		"token", "token",
		"id", "HandlerID",
	)
	ctx := metadata.NewContext(context.Background(), md)

	dlch := make(chan *pb.DownlinkMessage, 2)
	b.routers["routerID"] = dlch

	stream, _ := client.Publish(ctx)
	stream.Send(&pb.DownlinkMessage{
		DevEui: &devEUI,
		AppEui: &appEUI,
		DownlinkOption: &pb.DownlinkOption{
			Identifier: "routerID:scheduleID",
		},
	})
	ack, err := stream.CloseAndRecv()
	a.So(err, ShouldBeNil)
	a.So(ack, ShouldNotBeNil)

	<-time.After(10 * time.Millisecond)

	a.So(len(dlch), ShouldEqual, 1)
}

func TestActivate(t *testing.T) {
	// TODO
}

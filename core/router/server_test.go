// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package router

import (
	"fmt"
	"math/rand"
	"net"
	"sync"
	"testing"
	"time"

	"golang.org/x/net/context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"github.com/TheThingsNetwork/ttn/api"
	pb_gateway "github.com/TheThingsNetwork/ttn/api/gateway"
	pb "github.com/TheThingsNetwork/ttn/api/router"
	"github.com/TheThingsNetwork/ttn/core"
	"github.com/TheThingsNetwork/ttn/core/router/gateway"
	"github.com/TheThingsNetwork/ttn/core/types"
	. "github.com/TheThingsNetwork/ttn/utils/testing"
	. "github.com/smartystreets/assertions"
)

func randomPort() uint {
	rand.Seed(time.Now().UnixNano())
	port := rand.Intn(5000) + 5000
	return uint(port)
}

func buildTestRouterServer(t *testing.T, port uint) (*router, *grpc.Server) {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		panic(err)
	}
	r := &router{
		Component: &core.Component{
			Ctx: GetLogger(t, "TestRouterServer"),
		},
		gateways:        map[types.GatewayEUI]*gateway.Gateway{},
		brokerDiscovery: &mockBrokerDiscovery{},
	}
	s := grpc.NewServer()
	r.RegisterRPC(s)
	go s.Serve(lis)
	return r, s
}

func TestGatewayStatusRPC(t *testing.T) {
	a := New(t)

	port := randomPort()
	r, s := buildTestRouterServer(t, port)
	defer s.Stop()

	eui := types.GatewayEUI{1, 2, 3, 4, 5, 6, 7, 8}

	conn, err := grpc.Dial(fmt.Sprintf("localhost:%d", port), api.DialOptions...)
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	client := pb.NewRouterClient(conn)
	md := metadata.Pairs(
		"token", "token",
		"gateway_eui", eui.String(),
	)
	ctx := metadata.NewContext(context.Background(), md)
	stream, err := client.GatewayStatus(ctx)
	if err != nil {
		panic(err)
	}
	statusMessage := &pb_gateway.Status{Description: "Fake Gateway"}
	stream.Send(statusMessage)
	ack, err := stream.CloseAndRecv()
	a.So(err, ShouldBeNil)
	a.So(ack, ShouldResemble, &api.Ack{})

	<-time.After(5 * time.Millisecond)

	status, err := r.getGateway(eui).Status.Get()
	a.So(err, ShouldBeNil)
	a.So(status, ShouldNotBeNil)
	a.So(*status, ShouldResemble, *statusMessage)
}

func TestUplinkRPC(t *testing.T) {
	a := New(t)

	port := randomPort()
	r, s := buildTestRouterServer(t, port)
	defer s.Stop()

	eui := types.GatewayEUI{1, 2, 3, 4, 5, 6, 7, 8}

	conn, err := grpc.Dial(fmt.Sprintf("localhost:%d", port), api.DialOptions...)
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	client := pb.NewRouterClient(conn)
	md := metadata.Pairs(
		"token", "token",
		"gateway_eui", eui.String(),
	)
	ctx := metadata.NewContext(context.Background(), md)
	stream, err := client.Uplink(ctx)
	if err != nil {
		panic(err)
	}
	stream.Send(newReferenceUplink())
	ack, err := stream.CloseAndRecv()
	a.So(err, ShouldBeNil)
	a.So(ack, ShouldResemble, &api.Ack{})

	<-time.After(5 * time.Millisecond)

	utilization := r.getGateway(eui).Utilization
	utilization.Tick()
	rx, _ := utilization.Get()
	a.So(rx, ShouldBeGreaterThan, 0)
}

func TestSubscribeRPC(t *testing.T) {
	a := New(t)

	port := randomPort()
	r, s := buildTestRouterServer(t, port)
	defer s.Stop()

	eui := types.GatewayEUI{1, 2, 3, 4, 5, 6, 7, 8}

	conn, err := grpc.Dial(fmt.Sprintf("localhost:%d", port), api.DialOptions...)
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	client := pb.NewRouterClient(conn)
	md := metadata.Pairs(
		"token", "token",
		"gateway_eui", eui.String(),
	)
	ctx := metadata.NewContext(context.Background(), md)

	stream, err := client.Subscribe(ctx, &pb.SubscribeRequest{})
	a.So(err, ShouldBeNil)

	downlink := &pb.DownlinkMessage{Payload: []byte{1}}

	var wg sync.WaitGroup
	go func() {
		dl, err := stream.Recv()
		a.So(err, ShouldBeNil)
		a.So(*dl, ShouldResemble, *downlink)
		wg.Done()
	}()

	wg.Add(1)
	schedule := r.getGateway(eui).Schedule
	gateway.Deadline = 1 // Extremely short deadline
	schedule.Sync(0)
	id, _ := schedule.GetOption(300, 50)
	schedule.Schedule(id, downlink)

	wg.Wait()
}

func TestActivateRPC(t *testing.T) {
	a := New(t)

	port := randomPort()
	r, s := buildTestRouterServer(t, port)
	defer s.Stop()

	eui := types.GatewayEUI{1, 2, 3, 4, 5, 6, 7, 8}
	appEUI := types.AppEUI{1, 2, 3, 4, 5, 6, 7, 8}
	devEUI := types.DevEUI{1, 2, 3, 4, 5, 6, 7, 8}

	conn, err := grpc.Dial(fmt.Sprintf("localhost:%d", port), api.DialOptions...)
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	client := pb.NewRouterClient(conn)
	md := metadata.Pairs(
		"token", "token",
		"gateway_eui", eui.String(),
	)
	ctx := metadata.NewContext(context.Background(), md)
	uplink := newReferenceUplink()
	activation := &pb.DeviceActivationRequest{
		Payload:          []byte{},
		ProtocolMetadata: uplink.ProtocolMetadata,
		GatewayMetadata:  uplink.GatewayMetadata,
		AppEui:           &appEUI,
		DevEui:           &devEUI,
	}
	res, err := client.Activate(ctx, activation)
	a.So(res, ShouldBeNil)
	a.So(err, ShouldNotBeNil)

	<-time.After(5 * time.Millisecond)

	utilization := r.getGateway(eui).Utilization
	utilization.Tick()
	rx, _ := utilization.Get()
	a.So(rx, ShouldBeGreaterThan, 0)
}

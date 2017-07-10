// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package monitorclient

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/TheThingsNetwork/go-utils/grpc/auth"
	"github.com/TheThingsNetwork/go-utils/grpc/ttnctx"
	"github.com/TheThingsNetwork/go-utils/log"
	"github.com/TheThingsNetwork/ttn/api/broker"
	"github.com/TheThingsNetwork/ttn/api/gateway"
	"github.com/TheThingsNetwork/ttn/api/handler"
	"github.com/TheThingsNetwork/ttn/api/monitor"
	"github.com/TheThingsNetwork/ttn/api/monitor/monitorserver"
	"github.com/TheThingsNetwork/ttn/api/networkserver"
	"github.com/TheThingsNetwork/ttn/api/router"
	"github.com/htdvisser/grpc-testing/test"
	. "github.com/smartystreets/assertions"
	"google.golang.org/grpc"
)

func TestClient(t *testing.T) {
	waitTime := 20 * time.Millisecond

	a := New(t)

	testLogger := test.NewLogger()
	log.Set(testLogger)
	defer testLogger.Print(t)

	lis, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatalf("Failed to listen: %v", err)
	}
	s := grpc.NewServer()
	server := monitorserver.NewReferenceMonitorServer(10)

	monitor.RegisterMonitorServer(s, server)
	go s.Serve(lis)

	conn, err := grpc.Dial(lis.Addr().String(), grpc.WithInsecure())
	if err != nil {
		t.Fatalf("Failed to dial: %v", err)
	}

	cli := NewMonitorClient(WithConn("test", conn))

	// Gateway
	{
		auth := auth.WithStaticToken("token").WithInsecure()
		c := cli.GatewayClient(ttnctx.OutgoingContextWithID(context.Background(), "test"), grpc.PerRPCCredentials(auth))
		c.Send(&gateway.Status{})
		c.Send(&router.UplinkMessage{})
		c.Send(&router.DeviceActivationRequest{})
		c.Send(&router.DownlinkMessage{})
		time.Sleep(waitTime)
		a.So(server.Metrics.GatewayStatuses, ShouldEqual, 1)
		a.So(server.Metrics.UplinkMessages, ShouldEqual, 2)
		a.So(server.Metrics.DownlinkMessages, ShouldEqual, 1)
		c.Close()
	}

	// Router
	{
		auth := auth.WithStaticToken("token").WithInsecure()
		c := cli.RouterClient(ttnctx.OutgoingContextWithID(context.Background(), "test"), grpc.PerRPCCredentials(auth))
		c.Send(&router.Status{})
		time.Sleep(waitTime)
		a.So(server.Metrics.RouterStatuses, ShouldEqual, 1)
		c.Close()
	}

	// Broker
	{
		auth := auth.WithStaticToken("token").WithInsecure()
		c := cli.BrokerClient(ttnctx.OutgoingContextWithID(context.Background(), "test"), grpc.PerRPCCredentials(auth))
		c.Send(&broker.Status{})
		c.Send(&broker.DeduplicatedUplinkMessage{})
		c.Send(&broker.DeduplicatedDeviceActivationRequest{})
		c.Send(&broker.DownlinkMessage{})
		time.Sleep(waitTime)
		a.So(server.Metrics.BrokerStatuses, ShouldEqual, 1)
		a.So(server.Metrics.BrokerUplinkMessages, ShouldEqual, 2)
		a.So(server.Metrics.BrokerDownlinkMessages, ShouldEqual, 1)
		c.Close()
	}

	// Networkserver
	{
		auth := auth.WithStaticToken("token").WithInsecure()
		c := cli.NetworkServerClient(ttnctx.OutgoingContextWithID(context.Background(), "test"), grpc.PerRPCCredentials(auth))
		c.Send(&networkserver.Status{})
		time.Sleep(waitTime)
		a.So(server.Metrics.NetworkServerStatuses, ShouldEqual, 1)
		c.Close()
	}

	// Handler
	{
		auth := auth.WithStaticToken("token").WithInsecure()
		c := cli.HandlerClient(ttnctx.OutgoingContextWithID(context.Background(), "test"), grpc.PerRPCCredentials(auth))
		c.Send(&handler.Status{})
		c.Send(&broker.DeduplicatedUplinkMessage{})
		c.Send(&broker.DeduplicatedDeviceActivationRequest{})
		c.Send(&broker.DownlinkMessage{})
		time.Sleep(waitTime)
		a.So(server.Metrics.HandlerStatuses, ShouldEqual, 1)
		a.So(server.Metrics.HandlerUplinkMessages, ShouldEqual, 2)
		a.So(server.Metrics.HandlerDownlinkMessages, ShouldEqual, 1)
		c.Close()
	}

}

// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package monitor

import (
	"net"
	"testing"
	"time"

	"github.com/TheThingsNetwork/go-utils/log"
	"github.com/TheThingsNetwork/ttn/api/broker"
	"github.com/TheThingsNetwork/ttn/api/gateway"
	"github.com/TheThingsNetwork/ttn/api/handler"
	"github.com/TheThingsNetwork/ttn/api/networkserver"
	"github.com/TheThingsNetwork/ttn/api/pool"
	"github.com/TheThingsNetwork/ttn/api/router"
	"github.com/htdvisser/grpc-testing/test"
	. "github.com/smartystreets/assertions"
	"google.golang.org/grpc"
)

func TestMonitor(t *testing.T) {
	waitTime := 10 * time.Millisecond

	a := New(t)

	testLogger := test.NewLogger()
	log.Set(testLogger)
	defer testLogger.Print(t)

	lis, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatalf("Failed to listen: %v", err)
	}
	s := grpc.NewServer()
	server := NewReferenceMonitorServer(10)

	RegisterMonitorServer(s, server)
	go s.Serve(lis)

	cli := NewClient(DefaultClientConfig)

	cli.AddServer("tls-without-tls", lis.Addr().String())

	time.Sleep(50 * time.Millisecond)
	testLogger.Print(t)
	pool.Global.Close(lis.Addr().String())

	cli.AddServer("test", lis.Addr().String())
	time.Sleep(waitTime)
	defer func() {
		cli.Close()
		time.Sleep(waitTime)
		s.Stop()
	}()

	testLogger.Print(t)

	gtw := cli.NewGatewayStreams("test", "token")
	time.Sleep(waitTime)
	for i := 0; i < 20; i++ {
		gtw.Send(&router.UplinkMessage{})
		gtw.Send(&router.DeviceActivationRequest{})
		gtw.Send(&router.DownlinkMessage{})
		gtw.Send(&gateway.Status{})
		time.Sleep(time.Millisecond)
	}
	time.Sleep(waitTime)
	gtw.Close()
	time.Sleep(waitTime)

	a.So(server.metrics.uplinkMessages, ShouldEqual, 40)
	a.So(server.metrics.downlinkMessages, ShouldEqual, 20)
	a.So(server.metrics.gatewayStatuses, ShouldEqual, 20)

	testLogger.Print(t)

	rtr := cli.NewRouterStreams("test", "token")
	time.Sleep(waitTime)
	for i := 0; i < 20; i++ {
		rtr.Send(&router.Status{})
		time.Sleep(time.Millisecond)
	}
	time.Sleep(waitTime)
	rtr.Close()
	time.Sleep(waitTime)

	a.So(server.metrics.routerStatuses, ShouldEqual, 20)

	testLogger.Print(t)

	brk := cli.NewBrokerStreams("test", "token")
	time.Sleep(waitTime)
	brk.Send(&broker.DeduplicatedUplinkMessage{})
	brk.Send(&broker.DeduplicatedDeviceActivationRequest{})
	brk.Send(&broker.DownlinkMessage{})
	brk.Send(&broker.Status{})
	time.Sleep(waitTime)
	brk.Close()
	time.Sleep(waitTime)

	a.So(server.metrics.brokerUplinkMessages, ShouldEqual, 2)
	a.So(server.metrics.brokerDownlinkMessages, ShouldEqual, 1)

	testLogger.Print(t)

	ns := cli.NewNetworkServerStreams("test", "token")
	time.Sleep(waitTime)
	for i := 0; i < 20; i++ {
		ns.Send(&networkserver.Status{})
		time.Sleep(time.Millisecond)
	}
	time.Sleep(waitTime)
	ns.Close()
	time.Sleep(waitTime)

	a.So(server.metrics.routerStatuses, ShouldEqual, 20)

	testLogger.Print(t)

	hdl := cli.NewHandlerStreams("test", "token")
	time.Sleep(waitTime)
	hdl.Send(&broker.DeduplicatedUplinkMessage{})
	hdl.Send(&broker.DeduplicatedDeviceActivationRequest{})
	hdl.Send(&broker.DownlinkMessage{})
	hdl.Send(&handler.Status{})
	time.Sleep(waitTime)
	hdl.Close()
	time.Sleep(waitTime)

	a.So(server.metrics.handlerUplinkMessages, ShouldEqual, 2)
	a.So(server.metrics.handlerDownlinkMessages, ShouldEqual, 1)

	server.metrics.brokerUplinkMessages = 0

	cli.AddConn("test2", cli.serverConns[1].conn)

	brk = cli.NewBrokerStreams("test", "token")
	time.Sleep(waitTime)
	brk.Send(&broker.DeduplicatedUplinkMessage{})
	time.Sleep(waitTime)
	brk.Close()
	time.Sleep(waitTime)

	a.So(server.metrics.brokerUplinkMessages, ShouldEqual, 2)

	testLogger.Print(t)
}

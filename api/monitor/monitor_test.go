package monitor

import (
	"net"
	"testing"

	"time"

	"github.com/TheThingsNetwork/go-utils/log"
	"github.com/TheThingsNetwork/ttn/api/broker"
	"github.com/TheThingsNetwork/ttn/api/gateway"
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

	log.Get().Info("Expect err about grpc.WithInsecure()")
	cli.AddServer("invalid-config", lis.Addr().String())

	cli.AddServer("test", lis.Addr().String(), grpc.WithInsecure())
	time.Sleep(waitTime)
	defer func() {
		cli.Close()
		time.Sleep(waitTime)
		s.Stop()
	}()

	gtw := cli.NewGatewayStreams("test", "token")
	time.Sleep(waitTime)
	for i := 0; i < 20; i++ {
		gtw.Send(&router.UplinkMessage{})
		gtw.Send(&router.DownlinkMessage{})
		gtw.Send(&gateway.Status{})
		time.Sleep(time.Millisecond)
	}
	time.Sleep(waitTime)
	gtw.Close()
	time.Sleep(waitTime)

	a.So(server.metrics.uplinkMessages, ShouldEqual, 20)
	a.So(server.metrics.downlinkMessages, ShouldEqual, 20)
	a.So(server.metrics.gatewayStatuses, ShouldEqual, 20)

	brk := cli.NewBrokerStreams("test", "token")
	time.Sleep(waitTime)
	brk.Send(&broker.DeduplicatedUplinkMessage{})
	brk.Send(&broker.DownlinkMessage{})
	time.Sleep(waitTime)
	brk.Close()
	time.Sleep(waitTime)

	a.So(server.metrics.brokerUplinkMessages, ShouldEqual, 1)
	a.So(server.metrics.brokerDownlinkMessages, ShouldEqual, 1)

}

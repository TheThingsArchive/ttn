// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package router

import (
	"net"
	"testing"
	"time"

	"github.com/TheThingsNetwork/go-utils/log"
	"github.com/TheThingsNetwork/ttn/api/gateway"
	"github.com/TheThingsNetwork/ttn/api/pool"
	"github.com/htdvisser/grpc-testing/test"
	. "github.com/smartystreets/assertions"
	"google.golang.org/grpc"
)

func TestRouter(t *testing.T) {
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
	server := NewReferenceRouterServer(10)

	RegisterRouterServer(s, server)
	go s.Serve(lis)

	cli := NewClient(DefaultClientConfig)

	conn, err := pool.Global.DialInsecure(lis.Addr().String())
	if err != nil {
		t.Fatalf("Failed to connect to server: %v", err)
	}

	cli.AddServer("test", conn)
	time.Sleep(waitTime)
	defer func() {
		cli.Close()
		time.Sleep(waitTime)
		s.Stop()
	}()

	testLogger.Print(t)

	gtw := cli.NewGatewayStreams("test", "token", true)
	time.Sleep(waitTime)
	for i := 0; i < 20; i++ {
		gtw.Uplink(&UplinkMessage{})
		gtw.Status(&gateway.Status{})
		time.Sleep(time.Millisecond)
	}
	time.Sleep(waitTime)

	a.So(server.metrics.uplinkMessages, ShouldEqual, 20)
	a.So(server.metrics.gatewayStatuses, ShouldEqual, 20)

	testLogger.Print(t)

	downlink, _ := gtw.Downlink()
	recvDownlink := []*DownlinkMessage{}
	var downlinkClosed bool
	go func() {
		for msg := range downlink {
			recvDownlink = append(recvDownlink, msg)
		}
		downlinkClosed = true
	}()

	server.downlink["test"].ch <- &DownlinkMessage{}

	time.Sleep(waitTime)
	gtw.Close()
	time.Sleep(waitTime)

	a.So(recvDownlink, ShouldHaveLength, 1)
	a.So(downlinkClosed, ShouldBeTrue)

	testLogger.Print(t)
}

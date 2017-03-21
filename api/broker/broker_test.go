// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package broker

import (
	"net"
	"testing"
	"time"

	"github.com/TheThingsNetwork/go-utils/log"
	"github.com/TheThingsNetwork/ttn/api/pool"
	"github.com/htdvisser/grpc-testing/test"
	. "github.com/smartystreets/assertions"
	"google.golang.org/grpc"
)

func TestRouterBroker(t *testing.T) {
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
	server := NewReferenceBrokerServer(10)

	RegisterBrokerServer(s, server)
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

	rtr := cli.NewRouterStreams("test", "token")
	time.Sleep(waitTime)
	for i := 0; i < 20; i++ {
		rtr.Uplink(&UplinkMessage{})
		time.Sleep(time.Millisecond)
	}
	time.Sleep(waitTime)

	a.So(server.metrics.uplinkIn, ShouldEqual, 20)

	testLogger.Print(t)

	downlink := rtr.Downlink()
	recvDownlink := []*DownlinkMessage{}
	var downlinkClosed bool
	go func() {
		for msg := range downlink {
			recvDownlink = append(recvDownlink, msg)
		}
		downlinkClosed = true
	}()

	server.downlinkOut["test"].ch <- &DownlinkMessage{}

	time.Sleep(waitTime)
	rtr.Close()
	time.Sleep(waitTime)

	a.So(recvDownlink, ShouldHaveLength, 1)
	a.So(downlinkClosed, ShouldBeTrue)

	testLogger.Print(t)
}

func TestHandlerBroker(t *testing.T) {
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
	server := NewReferenceBrokerServer(10)

	RegisterBrokerServer(s, server)
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

	hdl := cli.NewHandlerStreams("test", "token")
	time.Sleep(waitTime)
	for i := 0; i < 20; i++ {
		hdl.Downlink(&DownlinkMessage{})
		time.Sleep(time.Millisecond)
	}
	time.Sleep(waitTime)

	a.So(server.metrics.downlinkIn, ShouldEqual, 20)

	testLogger.Print(t)

	uplink := hdl.Uplink()
	recvUplink := []*DeduplicatedUplinkMessage{}
	var uplinkClosed bool
	go func() {
		for msg := range uplink {
			recvUplink = append(recvUplink, msg)
		}
		uplinkClosed = true
	}()

	server.uplinkOut["test"].ch <- &DeduplicatedUplinkMessage{}

	time.Sleep(waitTime)
	hdl.Close()
	time.Sleep(waitTime)

	a.So(recvUplink, ShouldHaveLength, 1)
	a.So(uplinkClosed, ShouldBeTrue)

	testLogger.Print(t)
}

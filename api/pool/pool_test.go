// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package pool

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/TheThingsNetwork/go-utils/log"
	"github.com/TheThingsNetwork/ttn/api/health"
	"github.com/htdvisser/grpc-testing/test"
	. "github.com/smartystreets/assertions"
	"google.golang.org/grpc"
)

func TestPool(t *testing.T) {
	a := New(t)

	testLogger := test.NewLogger()
	log.Set(testLogger)
	defer testLogger.Print(t)

	lis, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatalf("Failed to listen: %v", err)
	}
	s := grpc.NewServer()
	health.RegisterServer(s)
	go s.Serve(lis)

	addr := lis.Addr().String()

	pool := NewPool(context.Background(), grpc.WithBlock())

	conn1, err := pool.DialInsecure(addr)
	a.So(err, ShouldBeNil)
	a.So(conn1, ShouldNotBeNil)

	conn2, err := pool.DialInsecure(addr)
	a.So(err, ShouldBeNil)
	a.So(conn2, ShouldEqual, conn1)

	{
		ok, err := health.Check(conn1)
		a.So(err, ShouldBeNil)
		a.So(ok, ShouldBeTrue)
	}

	{
		ok, err := health.Check(conn2)
		a.So(err, ShouldBeNil)
		a.So(ok, ShouldBeTrue)
	}

	s.Stop()

	time.Sleep(200 * time.Millisecond)

	{
		ok, err := health.Check(conn1)
		a.So(err, ShouldNotBeNil)
		a.So(ok, ShouldBeFalse)
	}

	lis, err = net.Listen("tcp", addr)
	if err != nil {
		t.Fatalf("Failed to listen: %v", err)
	}
	s = grpc.NewServer()
	health.RegisterServer(s)
	go s.Serve(lis)

	time.Sleep(time.Second)

	{
		ok, err := health.Check(conn1)
		a.So(err, ShouldBeNil)
		a.So(ok, ShouldBeTrue)
	}

	pool.Close(addr)
	pool.Close(addr)

	conn3, err := pool.DialInsecure(addr)
	a.So(err, ShouldBeNil)
	a.So(conn3, ShouldNotEqual, conn1) // the connection was closed, because there were no more users

	pool.Close()

	pool = NewPool(context.Background(), grpc.WithInsecure()) // Without the grpc.WithBlock()

	conn4, err := pool.DialInsecure(addr)
	a.So(err, ShouldBeNil)
	a.So(conn4, ShouldNotBeNil)

	{
		ok, err := health.Check(conn4)
		a.So(err, ShouldBeNil)
		a.So(ok, ShouldBeTrue)
	}

	pool.CloseConn(conn4)
	a.So(pool.conns, ShouldBeEmpty)
}

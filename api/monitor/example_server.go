// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package monitor

import (
	"fmt"
	"io"
	"net"

	"github.com/TheThingsNetwork/ttn/api/gateway"
	"github.com/TheThingsNetwork/ttn/api/router"
	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

var errNotImplemented = grpc.Errorf(codes.Unimplemented, "That's not implemented yet")
var errBufferFull = grpc.Errorf(codes.ResourceExhausted, "Take it easy, dude! My buffers are full")

func newExampleServer(channelSize int) *exampleServer {
	return &exampleServer{
		gatewayStatuses:  make(chan *gateway.Status, channelSize),
		uplinkMessages:   make(chan *router.UplinkMessage, channelSize),
		downlinkMessages: make(chan *router.DownlinkMessage, channelSize),
	}
}

type exampleServer struct {
	gatewayStatuses  chan *gateway.Status
	uplinkMessages   chan *router.UplinkMessage
	downlinkMessages chan *router.DownlinkMessage
}

func (s *exampleServer) GatewayStatus(stream Monitor_GatewayStatusServer) error {
	for {
		status, err := stream.Recv()
		if err == io.EOF {
			return stream.SendAndClose(&empty.Empty{})
		}
		if err != nil {
			return err
		}
		select {
		case s.gatewayStatuses <- status:
			fmt.Println("Saving gateway status to database and doing something cool")
		default:
			fmt.Println("Warning: Dropping gateway status [full buffer]")
			return errBufferFull
		}
	}
}

func (s *exampleServer) GatewayUplink(stream Monitor_GatewayUplinkServer) error {
	for {
		uplink, err := stream.Recv()
		if err == io.EOF {
			return stream.SendAndClose(&empty.Empty{})
		}
		if err != nil {
			return err
		}
		select {
		case s.uplinkMessages <- uplink:
			fmt.Println("Saving uplink to database and doing something cool")
		default:
			fmt.Println("Warning: Dropping uplink [full buffer]")
			return errBufferFull
		}
	}
}

func (s *exampleServer) GatewayDownlink(stream Monitor_GatewayDownlinkServer) error {
	for {
		downlink, err := stream.Recv()
		if err == io.EOF {
			return stream.SendAndClose(&empty.Empty{})
		}
		if err != nil {
			return err
		}
		select {
		case s.downlinkMessages <- downlink:
			fmt.Println("Saving downlink to database and doing something cool")
		default:
			fmt.Println("Warning: Dropping downlink [full buffer]")
			return errBufferFull
		}
	}
}

func (s *exampleServer) RouterStatus(stream Monitor_RouterStatusServer) error {
	return errNotImplemented
}

func (s *exampleServer) BrokerStatus(stream Monitor_BrokerStatusServer) error {
	return errNotImplemented
}

func (s *exampleServer) HandlerStatus(stream Monitor_HandlerStatusServer) error {
	return errNotImplemented
}

func (s *exampleServer) Serve(port int) {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		panic(err)
	}
	srv := grpc.NewServer()
	RegisterMonitorServer(srv, s)
	srv.Serve(lis)
}

func startExampleServer(channelSize, port int) {
	s := newExampleServer(channelSize)
	s.Serve(port)
}

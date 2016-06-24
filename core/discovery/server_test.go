// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package discovery

import (
	"fmt"
	"math/rand"
	"net"
	"time"

	"github.com/TheThingsNetwork/ttn/api"
	pb "github.com/TheThingsNetwork/ttn/api/discovery"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

func randomPort() uint {
	rand.Seed(time.Now().UnixNano())
	port := rand.Intn(5000) + 5000
	return uint(port)
}

func buildTestDiscoveryServer(port uint) (*discovery, *grpc.Server) {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		panic(err)
	}
	d := &discovery{}
	s := grpc.NewServer()
	d.RegisterRPC(s)
	go s.Serve(lis)

	return d, s
}

func buildMockDiscoveryServer(port uint) (*mockDiscoveryServer, *grpc.Server) {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		panic(err)
	}
	d := &mockDiscoveryServer{}
	s := grpc.NewServer()
	pb.RegisterDiscoveryServer(s, d)
	go s.Serve(lis)
	return d, s
}

type mockDiscoveryServer struct {
	announce uint
	discover uint
	get      uint
}

func (d *mockDiscoveryServer) Announce(ctx context.Context, announcement *pb.Announcement) (*api.Ack, error) {
	d.announce++
	<-time.After(5 * time.Millisecond)
	return &api.Ack{}, nil
}
func (d *mockDiscoveryServer) Discover(ctx context.Context, req *pb.DiscoverRequest) (*pb.DiscoverResponse, error) {
	d.discover++
	<-time.After(5 * time.Millisecond)
	return &pb.DiscoverResponse{
		Services: []*pb.Announcement{},
	}, nil
}
func (d *mockDiscoveryServer) Get(ctx context.Context, req *pb.GetRequest) (*pb.DiscoverResponse, error) {
	d.get++
	<-time.After(5 * time.Millisecond)
	return &pb.DiscoverResponse{
		Services: []*pb.Announcement{},
	}, nil
}

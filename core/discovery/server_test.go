// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package discovery

import (
	"fmt"
	"math/rand"
	"net"
	"time"

	pb "github.com/TheThingsNetwork/ttn/api/discovery"
	"github.com/golang/protobuf/ptypes/empty"
	"golang.org/x/net/context" // See https://github.com/grpc/grpc-go/issues/711"
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

func (d *mockDiscoveryServer) Announce(ctx context.Context, announcement *pb.Announcement) (*empty.Empty, error) {
	d.announce++
	<-time.After(5 * time.Millisecond)
	return &empty.Empty{}, nil
}
func (d *mockDiscoveryServer) GetAll(ctx context.Context, req *pb.GetServiceRequest) (*pb.AnnouncementsResponse, error) {
	d.discover++
	<-time.After(5 * time.Millisecond)
	return &pb.AnnouncementsResponse{
		Services: []*pb.Announcement{},
	}, nil
}
func (d *mockDiscoveryServer) Get(ctx context.Context, req *pb.GetRequest) (*pb.Announcement, error) {
	d.get++
	<-time.After(5 * time.Millisecond)
	return &pb.Announcement{}, nil
}
func (d *mockDiscoveryServer) AddMetadata(ctx context.Context, in *pb.MetadataRequest) (*empty.Empty, error) {
	<-time.After(5 * time.Millisecond)
	return &empty.Empty{}, nil
}
func (d *mockDiscoveryServer) DeleteMetadata(ctx context.Context, in *pb.MetadataRequest) (*empty.Empty, error) {
	<-time.After(5 * time.Millisecond)
	return &empty.Empty{}, nil
}
func (d *mockDiscoveryServer) GetByAppID(ctx context.Context, req *pb.GetByAppIDRequest) (*pb.Announcement, error) {
	<-time.After(5 * time.Millisecond)
	return &pb.Announcement{}, nil
}
func (d *mockDiscoveryServer) GetByAppEUI(ctx context.Context, req *pb.GetByAppEUIRequest) (*pb.Announcement, error) {
	<-time.After(5 * time.Millisecond)
	return &pb.Announcement{}, nil
}

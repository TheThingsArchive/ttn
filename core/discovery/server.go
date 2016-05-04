package discovery

import (
	"github.com/TheThingsNetwork/ttn/api"
	pb "github.com/TheThingsNetwork/ttn/api/discovery"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

type discoveryServer struct {
	discovery Discovery
}

func (d *discoveryServer) Announce(ctx context.Context, announcement *pb.Announcement) (*api.Ack, error) {
	err := d.discovery.Announce(announcement)
	if err != nil {
		return nil, err
	}
	return &api.Ack{}, nil
}

func (d *discoveryServer) Discover(ctx context.Context, req *pb.DiscoverRequest) (*pb.DiscoverResponse, error) {
	services, err := d.discovery.Discover(req.ServiceName)
	if err != nil {
		return nil, err
	}
	return &pb.DiscoverResponse{
		Services: services,
	}, nil
}

func (d *discovery) RegisterDiscoveryServer(s *grpc.Server) {
	server := &discoveryServer{d}
	pb.RegisterDiscoveryServer(s, server)
}

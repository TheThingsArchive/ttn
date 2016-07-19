// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package discovery

import (
	"errors"

	"github.com/TheThingsNetwork/ttn/api"
	pb "github.com/TheThingsNetwork/ttn/api/discovery"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

type discoveryServer struct {
	discovery Discovery
}

func (d *discoveryServer) Announce(ctx context.Context, announcement *pb.Announcement) (*api.Ack, error) {
	claims, err := d.discovery.ValidateContext(ctx)
	if err != nil {
		return nil, err
	}
	_ = claims // TODO: Check the claims for the announced Component ID
	announcementCopy := *announcement
	announcement.Metadata = []*pb.Metadata{} // This will be taken from existing announcement
	err = d.discovery.Announce(&announcementCopy)
	if err != nil {
		return nil, err
	}
	return &api.Ack{}, nil
}

func (d *discoveryServer) AddMetadata(ctx context.Context, in *pb.MetadataRequest) (*api.Ack, error) {
	claims, err := d.discovery.ValidateContext(ctx)
	if err != nil {
		return nil, err
	}
	switch in.Metadata.Key {
	case pb.Metadata_PREFIX:
		// Allow announcing any PREFIX
	case pb.Metadata_APP_EUI:
		// TODO: Check the claims for the announced APP_EUI
		return nil, errors.New("ttn/discovery: Can not announce AppEUIs at this time")
	case pb.Metadata_APP_ID:
		if !claims.CanEditApp(string(in.Metadata.Value)) {
			return nil, errors.New("ttn/discovery: No access to this application")
		}
	}
	err = d.discovery.AddMetadata(in.ServiceName, in.Id, in.Metadata)
	if err != nil {
		return nil, err
	}
	return &api.Ack{}, nil
}

func (d *discoveryServer) DeleteMetadata(ctx context.Context, in *pb.MetadataRequest) (*api.Ack, error) {
	claims, err := d.discovery.ValidateContext(ctx)
	if err != nil {
		return nil, err
	}
	_ = claims // TODO: Check the claims for the announced Component ID
	err = d.discovery.DeleteMetadata(in.ServiceName, in.Id, in.Metadata)
	if err != nil {
		return nil, err
	}
	return &api.Ack{}, nil
}

func (d *discoveryServer) GetAll(ctx context.Context, req *pb.GetAllRequest) (*pb.AnnouncementsResponse, error) {
	services, err := d.discovery.GetAll(req.ServiceName)
	if err != nil {
		return nil, err
	}
	return &pb.AnnouncementsResponse{
		Services: services,
	}, nil
}

func (d *discoveryServer) Get(ctx context.Context, req *pb.GetRequest) (*pb.Announcement, error) {
	service, err := d.discovery.Get(req.ServiceName, req.Id)
	if err != nil {
		return nil, err
	}
	return service, nil
}

// RegisterRPC registers the local discovery with a gRPC server
func (d *discovery) RegisterRPC(s *grpc.Server) {
	server := &discoveryServer{d}
	pb.RegisterDiscoveryServer(s, server)
}

// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package discovery

import (
	pb "github.com/TheThingsNetwork/ttn/api/discovery"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	"github.com/golang/protobuf/ptypes/empty"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

type discoveryServer struct {
	discovery *discovery
}

var grpcErrf = grpc.Errorf // To make go vet stop complaining

func errPermissionDeniedf(format string, args ...string) error {
	return grpcErrf(codes.PermissionDenied, "Discovery:"+format, args)
}

func (d *discoveryServer) checkMetadataEditRights(ctx context.Context, in *pb.MetadataRequest) error {
	claims, err := d.discovery.ValidateTTNAuthContext(ctx)
	if err != nil {
		return err
	}
	switch in.Metadata.Key {
	case pb.Metadata_PREFIX:
		if in.ServiceName != "broker" {
			return errPermissionDeniedf("Announcement service type should be \"broker\"")
		}
		// Only allow prefix announcements if token is issued by the official ttn account server (or if in dev mode)
		if claims.Issuer != "ttn-account" && d.discovery.Component.Identity.Id != "dev" {
			return errPermissionDeniedf("Token issuer \"%s\" should be \"ttn-account\"", claims.Issuer)
		}
		if claims.Type != in.ServiceName {
			return errPermissionDeniedf("Token type %s does not correspond with announcement service type %s", claims.Type, in.ServiceName)
		}
		if claims.Subject != in.Id {
			return errPermissionDeniedf("Token subject %s does not correspond with announcement id %s", claims.Subject, in.Id)
		}
		// TODO: Check if this PREFIX can be announced
	case pb.Metadata_APP_EUI:
		if in.ServiceName != "handler" {
			return errPermissionDeniedf("Announcement service type should be \"handler\"")
		}
		// Only allow eui announcements if token is issued by the official ttn account server (or if in dev mode)
		if claims.Issuer != "ttn-account" && d.discovery.Component.Identity.Id != "dev" {
			return errPermissionDeniedf("Token issuer %s should be ttn-account", claims.Issuer)
		}
		if claims.Type != in.ServiceName {
			return errPermissionDeniedf("Token type %s does not correspond with announcement service type %s", claims.Type, in.ServiceName)
		}
		if claims.Subject != in.Id {
			return errPermissionDeniedf("Token subject %s does not correspond with announcement id %s", claims.Subject, in.Id)
		}
		// TODO: Check if this APP_EUI can be announced
		return errPermissionDeniedf("Can not announce AppEUIs at this time")
	case pb.Metadata_APP_ID:
		if in.ServiceName != "handler" {
			return errPermissionDeniedf("Announcement service type should be \"handler\"")
		}
		// Allow APP_ID announcements from all trusted auth servers
		// When announcing APP_ID, token is user token that contains apps
		if !claims.CanEditApp(string(in.Metadata.Value)) {
			return errPermissionDeniedf("No access to this application")
		}
	}
	return nil
}

func (d *discoveryServer) Announce(ctx context.Context, announcement *pb.Announcement) (*empty.Empty, error) {
	claims, err := d.discovery.ValidateTTNAuthContext(ctx)
	if err != nil {
		return nil, err
	}
	// Only allow announcements if token is issued by the official ttn account server (or if in dev mode)
	if claims.Issuer != "ttn-account" && d.discovery.Component.Identity.Id != "dev" {
		return nil, errPermissionDeniedf("Token issuer %s should be ttn-account", claims.Issuer)
	}
	if claims.Subject != announcement.Id {
		return nil, errPermissionDeniedf("Token subject %s does not correspond with announcement ID %s", claims.Subject, announcement.Id)
	}
	if claims.Type != announcement.ServiceName {
		return nil, errPermissionDeniedf("Token type %s does not correspond with announcement service type %s", claims.Type, announcement.ServiceName)
	}
	announcementCopy := *announcement
	announcement.Metadata = []*pb.Metadata{} // This will be taken from existing announcement
	err = d.discovery.Announce(&announcementCopy)
	if err != nil {
		return nil, errors.BuildGRPCError(err)
	}
	return &empty.Empty{}, nil
}

func (d *discoveryServer) AddMetadata(ctx context.Context, in *pb.MetadataRequest) (*empty.Empty, error) {
	err := d.checkMetadataEditRights(ctx, in)
	if err != nil {
		return nil, err
	}
	err = d.discovery.AddMetadata(in.ServiceName, in.Id, in.Metadata)
	if err != nil {
		return nil, errors.BuildGRPCError(err)
	}
	return &empty.Empty{}, nil
}

func (d *discoveryServer) DeleteMetadata(ctx context.Context, in *pb.MetadataRequest) (*empty.Empty, error) {
	err := d.checkMetadataEditRights(ctx, in)
	if err != nil {
		return nil, err
	}
	err = d.discovery.DeleteMetadata(in.ServiceName, in.Id, in.Metadata)
	if err != nil {
		return nil, errors.BuildGRPCError(err)
	}
	return &empty.Empty{}, nil
}

func (d *discoveryServer) GetAll(ctx context.Context, req *pb.GetAllRequest) (*pb.AnnouncementsResponse, error) {
	services, err := d.discovery.GetAll(req.ServiceName)
	if err != nil {
		return nil, errors.BuildGRPCError(err)
	}
	return &pb.AnnouncementsResponse{
		Services: services,
	}, nil
}

func (d *discoveryServer) Get(ctx context.Context, req *pb.GetRequest) (*pb.Announcement, error) {
	service, err := d.discovery.Get(req.ServiceName, req.Id)
	if err != nil {
		return nil, errors.BuildGRPCError(err)
	}
	return service, nil
}

// RegisterRPC registers the local discovery with a gRPC server
func (d *discovery) RegisterRPC(s *grpc.Server) {
	server := &discoveryServer{d}
	pb.RegisterDiscoveryServer(s, server)
}

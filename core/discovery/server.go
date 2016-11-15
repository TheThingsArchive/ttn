// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package discovery

import (
	"fmt"

	"github.com/TheThingsNetwork/go-account-lib/rights"
	pb "github.com/TheThingsNetwork/ttn/api/discovery"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	"github.com/golang/protobuf/ptypes/empty"
	"golang.org/x/net/context" // See https://github.com/grpc/grpc-go/issues/711"
	"google.golang.org/grpc"
)

type discoveryServer struct {
	discovery *discovery
}

func errPermissionDeniedf(format string, args ...string) error {
	return errors.BuildGRPCError(errors.NewErrPermissionDenied(fmt.Sprintf("Discovery:"+format, args)))
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
		// Only allow prefix announcements if token is issued by a master auth server (or if in dev mode)
		if d.discovery.Component.Identity.Id != "dev" && !d.discovery.IsMasterAuthServer(claims.Issuer) {
			return errPermissionDeniedf("Token issuer \"%s\" is not allowed to make changes to the network settings", claims.Issuer)
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
		// Only allow eui announcements if token is issued by a master auth server (or if in dev mode)
		if d.discovery.Component.Identity.Id != "dev" && !d.discovery.IsMasterAuthServer(claims.Issuer) {
			return errPermissionDeniedf("Token issuer %s is not allowed to make changes to the network settings", claims.Issuer)
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
		if !claims.AppRight(string(in.Metadata.Value), rights.AppSettings) {
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

	// If not in development mode
	if d.discovery.Component.Identity.Id != "dev" {
		if !d.discovery.IsMasterAuthServer(claims.Issuer) {
			return nil, errPermissionDeniedf("Token issuer %s is not allowed to make changes to the network settings", claims.Issuer)
		}

		// Can't announce development components
		if claims.Subject == "dev" {
			return nil, errPermissionDeniedf("Can't announce development components to production networks")
		}
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

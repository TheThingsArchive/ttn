// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package discovery

import (
	"fmt"

	"github.com/TheThingsNetwork/go-account-lib/rights"
	"github.com/TheThingsNetwork/go-utils/grpc/ttnctx"
	pb "github.com/TheThingsNetwork/ttn/api/discovery"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	"github.com/golang/protobuf/ptypes/empty"
	"golang.org/x/net/context" // See https://github.com/grpc/grpc-go/issues/711"
	"google.golang.org/grpc"
)

type discoveryServer struct {
	discovery *discovery
}

func errPermissionDeniedf(format string, args ...interface{}) error {
	return errors.NewErrPermissionDenied(fmt.Sprintf("Discovery:"+format, args...))
}

func (d *discoveryServer) checkMetadataEditRights(ctx context.Context, in *pb.MetadataRequest) error {
	claims, err := d.discovery.ValidateTTNAuthContext(ctx)
	if err != nil {
		return err
	}

	appEUI := in.Metadata.GetAppEui()
	appID := in.Metadata.GetAppId()
	prefix := in.Metadata.GetDevAddrPrefix()

	if appEUI == nil && appID == "" && prefix == nil {
		return errPermissionDeniedf("Unknown Metadata type")
	}

	// AppEUI and AppID can only be added to Handlers
	if (appEUI != nil || appID != "") && in.ServiceName != "handler" {
		return errPermissionDeniedf("Announcement service type should be \"handler\"")
	}

	// DevAddrPrefix can only be added to Brokers
	if prefix != nil && in.ServiceName != "broker" {
		return errPermissionDeniedf("Announcement service type should be \"broker\"")
	}

	// DevAddrPrefix and AppEUI are network level changes
	if prefix != nil || appEUI != nil {

		// If not in develop mode
		if d.discovery.Component.Identity.Id != "dev" {

			// We require a signature from a master auth server
			if !d.discovery.IsMasterAuthServer(claims.Issuer) {
				return errPermissionDeniedf("Token issuer \"%s\" is not allowed to make changes to the network settings", claims.Issuer)
			}

			// TODO: Check if claims allow DevAddrPrefix to be announced

			// AppEUI can not be announced yet
			if appEUI != nil {
				return errPermissionDeniedf("Can not announce AppEUIs at this time")
			}
		}

		// Can only be announced to "self"
		if claims.Type != in.ServiceName {
			return errPermissionDeniedf("Token type %s does not correspond with announcement service type %s", claims.Type, in.ServiceName)
		}
		if claims.Subject != in.Id {
			return errPermissionDeniedf("Token subject %s does not correspond with announcement id %s", claims.Subject, in.Id)
		}
	}

	// Check claims for AppID
	if appID != "" {
		if !claims.AppRight(appID, rights.AppDelete) {
			return errPermissionDeniedf(`No "%s" rights to Application "%s"`, rights.AppDelete, appID)
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
		return nil, err
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
		return nil, err
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
		return nil, err
	}
	return &empty.Empty{}, nil
}

func (d *discoveryServer) GetAll(ctx context.Context, req *pb.GetServiceRequest) (*pb.AnnouncementsResponse, error) {
	limit, offset, err := ttnctx.LimitAndOffsetFromIncomingContext(ctx)
	if err != nil {
		return nil, err
	}
	services, err := d.discovery.GetAll(req.ServiceName, limit, offset)
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

func (d *discoveryServer) GetByAppID(ctx context.Context, req *pb.GetByAppIDRequest) (*pb.Announcement, error) {
	service, err := d.discovery.GetByAppID(req.AppId)
	if err != nil {
		return nil, err
	}
	return service, nil
}

func (d *discoveryServer) GetByAppEUI(ctx context.Context, req *pb.GetByAppEUIRequest) (*pb.Announcement, error) {
	if req.AppEui == nil {
		return nil, errors.NewErrInvalidArgument("AppEUI", "empty")
	}
	service, err := d.discovery.GetByAppEUI(*req.AppEui)
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

// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package discovery

import (
	"fmt"

	pb "github.com/TheThingsNetwork/api/discovery"
	"github.com/TheThingsNetwork/go-account-lib/rights"
	"github.com/TheThingsNetwork/go-utils/grpc/ttnctx"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	"github.com/gogo/protobuf/types"
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

	appEUI := in.Metadata.GetAppEUI()
	appID := in.Metadata.GetAppID()
	prefix := in.Metadata.GetDevAddrPrefix()
	gatewayID := in.Metadata.GetGatewayID()

	if appEUI == nil && appID == "" && prefix == nil && gatewayID == "" {
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

	// GatewayID can only be added to Routers
	if gatewayID != "" && in.ServiceName != "router" {
		return errPermissionDeniedf("Announcement service type should be \"router\"")
	}

	// DevAddrPrefix and AppEUI are network level changes
	if prefix != nil || appEUI != nil {

		// If not in develop mode
		if d.discovery.Component.Identity.ID != "dev" {

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
		if claims.Subject != in.ID {
			return errPermissionDeniedf("Token subject %s does not correspond with announcement id %s", claims.Subject, in.ID)
		}
	}

	// Check claims for AppID
	if appID != "" {
		if !claims.AppRight(appID, rights.AppDelete) {
			return errPermissionDeniedf(`No "%s" rights to Application "%s"`, rights.AppDelete, appID)
		}
	}

	// Check claims for GatewayID
	if gatewayID != "" {
		if !claims.GatewayRight(gatewayID, rights.GatewayDelete) && !(claims.Type == "gateway" && claims.Subject == gatewayID) {
			return errPermissionDeniedf(`No "%s" rights to Gateway "%s"`, rights.GatewayDelete, gatewayID)
		}
	}
	return nil
}

func (d *discoveryServer) Announce(ctx context.Context, announcement *pb.Announcement) (*types.Empty, error) {
	claims, err := d.discovery.ValidateTTNAuthContext(ctx)
	if err != nil {
		return nil, err
	}

	// If not in development mode
	if d.discovery.Component.Identity.ID != "dev" {
		if !d.discovery.IsMasterAuthServer(claims.Issuer) {
			return nil, errPermissionDeniedf("Token issuer %s is not allowed to make changes to the network settings", claims.Issuer)
		}

		// Can't announce development components
		if claims.Subject == "dev" {
			return nil, errPermissionDeniedf("Can't announce development components to production networks")
		}
	}

	if claims.Subject != announcement.ID {
		return nil, errPermissionDeniedf("Token subject %s does not correspond with announcement ID %s", claims.Subject, announcement.ID)
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
	return &types.Empty{}, nil
}

func (d *discoveryServer) AddMetadata(ctx context.Context, in *pb.MetadataRequest) (*types.Empty, error) {
	err := d.checkMetadataEditRights(ctx, in)
	if err != nil {
		return nil, err
	}
	err = d.discovery.AddMetadata(in.ServiceName, in.ID, &in.Metadata)
	if err != nil {
		return nil, err
	}
	return &types.Empty{}, nil
}

func (d *discoveryServer) DeleteMetadata(ctx context.Context, in *pb.MetadataRequest) (*types.Empty, error) {
	err := d.checkMetadataEditRights(ctx, in)
	if err != nil {
		return nil, err
	}
	err = d.discovery.DeleteMetadata(in.ServiceName, in.ID, &in.Metadata)
	if err != nil {
		return nil, err
	}
	return &types.Empty{}, nil
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
	service, err := d.discovery.Get(req.ServiceName, req.ID)
	if err != nil {
		return nil, err
	}
	return service, nil
}

func (d *discoveryServer) GetByAppID(ctx context.Context, req *pb.GetByAppIDRequest) (*pb.Announcement, error) {
	service, err := d.discovery.GetByAppID(req.AppID)
	if err != nil {
		return nil, err
	}
	return service, nil
}

func (d *discoveryServer) GetByGatewayID(ctx context.Context, req *pb.GetByGatewayIDRequest) (*pb.Announcement, error) {
	service, err := d.discovery.GetByGatewayID(req.GatewayID)
	if err != nil {
		return nil, err
	}
	return service, nil
}

func (d *discoveryServer) GetByAppEUI(ctx context.Context, req *pb.GetByAppEUIRequest) (*pb.Announcement, error) {
	if req.AppEUI.IsEmpty() {
		return nil, errors.NewErrInvalidArgument("AppEUI", "empty")
	}
	service, err := d.discovery.GetByAppEUI(req.AppEUI)
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

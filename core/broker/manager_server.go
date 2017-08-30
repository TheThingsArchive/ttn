// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package broker

import (
	"fmt"
	"time"

	pb "github.com/TheThingsNetwork/api/broker"
	"github.com/TheThingsNetwork/api/discovery"
	"github.com/TheThingsNetwork/api/protocol/lorawan"
	pb_lorawan "github.com/TheThingsNetwork/api/protocol/lorawan"
	"github.com/TheThingsNetwork/go-account-lib/claims"
	"github.com/TheThingsNetwork/go-account-lib/rights"
	"github.com/TheThingsNetwork/go-utils/grpc/ttnctx"
	"github.com/TheThingsNetwork/ttn/api/ratelimit"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	"github.com/gogo/protobuf/types"
	"golang.org/x/net/context" // See https://github.com/grpc/grpc-go/issues/711"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

type brokerManager struct {
	broker         *broker
	deviceManager  pb_lorawan.DeviceManagerClient
	devAddrManager pb_lorawan.DevAddrManagerClient
	clientRate     *ratelimit.Registry
}

func (b *brokerManager) validateClient(ctx context.Context) (*claims.Claims, error) {
	claims, err := b.broker.ValidateTTNAuthContext(ctx)
	if err != nil {
		return nil, err
	}
	if b.clientRate.Limit(claims.Subject) {
		return claims, grpc.Errorf(codes.ResourceExhausted, "Rate limit for client reached")
	}
	return claims, nil
}

func (b *brokerManager) GetDevice(ctx context.Context, in *lorawan.DeviceIdentifier) (*lorawan.Device, error) {
	if _, err := b.validateClient(ctx); err != nil {
		return nil, err
	}
	token, _ := ttnctx.TokenFromIncomingContext(ctx)
	res, err := b.deviceManager.GetDevice(ttnctx.OutgoingContextWithToken(ctx, token), in)
	if err != nil {
		return nil, errors.Wrap(errors.FromGRPCError(err), "NetworkServer did not return device")
	}
	return res, nil
}

func (b *brokerManager) SetDevice(ctx context.Context, in *lorawan.Device) (*types.Empty, error) {
	if _, err := b.validateClient(ctx); err != nil {
		return nil, err
	}
	token, _ := ttnctx.TokenFromIncomingContext(ctx)
	res, err := b.deviceManager.SetDevice(ttnctx.OutgoingContextWithToken(ctx, token), in)
	if err != nil {
		return nil, errors.Wrap(errors.FromGRPCError(err), "NetworkServer did not set device")
	}
	return res, nil
}

func (b *brokerManager) DeleteDevice(ctx context.Context, in *lorawan.DeviceIdentifier) (*types.Empty, error) {
	if _, err := b.validateClient(ctx); err != nil {
		return nil, err
	}
	token, _ := ttnctx.TokenFromIncomingContext(ctx)
	res, err := b.deviceManager.DeleteDevice(ttnctx.OutgoingContextWithToken(ctx, token), in)
	if err != nil {
		return nil, errors.Wrap(errors.FromGRPCError(err), "NetworkServer did not delete device")
	}
	return res, nil
}

func (b *brokerManager) RegisterApplicationHandler(ctx context.Context, in *pb.ApplicationHandlerRegistration) (*types.Empty, error) {
	claims, err := b.broker.Component.ValidateTTNAuthContext(ctx)
	if err != nil {
		return nil, err
	}
	if err := in.Validate(); err != nil {
		return nil, errors.Wrap(err, "Invalid Application Handler Registration")
	}
	if !claims.AppRight(in.AppID, rights.AppSettings) {
		return nil, errors.NewErrPermissionDenied("No access to this application")
	}
	// Add Handler in local cache
	handler, err := b.broker.Discovery.Get("handler", in.HandlerID)
	if err != nil {
		return nil, errors.NewErrInternal("Could not get Handler Announcement")
	}
	handler.Metadata = append(handler.Metadata, &discovery.Metadata{Metadata: &discovery.Metadata_AppID{
		AppID: in.AppID,
	}})
	return &types.Empty{}, nil
}

func (b *brokerManager) GetPrefixes(ctx context.Context, in *lorawan.PrefixesRequest) (*lorawan.PrefixesResponse, error) {
	res, err := b.devAddrManager.GetPrefixes(ctx, in)
	if err != nil {
		return nil, errors.Wrap(errors.FromGRPCError(err), "NetworkServer did not return prefixes")
	}
	return res, nil
}

func (b *brokerManager) GetDevAddr(ctx context.Context, in *lorawan.DevAddrRequest) (*lorawan.DevAddrResponse, error) {
	res, err := b.devAddrManager.GetDevAddr(ctx, in)
	if err != nil {
		return nil, errors.Wrap(errors.FromGRPCError(err), "NetworkServer did not return DevAddr")
	}
	return res, nil
}

func (b *brokerManager) GetStatus(ctx context.Context, in *pb.StatusRequest) (*pb.Status, error) {
	if b.broker.Identity.ID != "dev" {
		claims, err := b.broker.ValidateTTNAuthContext(ctx)
		if err != nil {
			return nil, errors.Wrap(err, "No access")
		}
		if !claims.ComponentAccess(b.broker.Identity.ID) {
			return nil, errors.NewErrPermissionDenied(fmt.Sprintf("Claims do not grant access to %s", b.broker.Identity.ID))
		}
	}
	status := b.broker.GetStatus()
	if status == nil {
		return new(pb.Status), nil
	}
	return status, nil
}

func (b *broker) RegisterManager(s *grpc.Server) {
	server := &brokerManager{
		broker:         b,
		deviceManager:  pb_lorawan.NewDeviceManagerClient(b.nsConn),
		devAddrManager: pb_lorawan.NewDevAddrManagerClient(b.nsConn),
	}

	server.clientRate = ratelimit.NewRegistry(5000, time.Hour)

	pb.RegisterBrokerManagerServer(s, server)
	lorawan.RegisterDeviceManagerServer(s, server)
	lorawan.RegisterDevAddrManagerServer(s, server)
}

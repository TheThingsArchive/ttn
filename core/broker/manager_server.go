// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package broker

import (
	"github.com/TheThingsNetwork/go-account-lib/rights"
	pb "github.com/TheThingsNetwork/ttn/api/broker"
	"github.com/TheThingsNetwork/ttn/api/discovery"
	"github.com/TheThingsNetwork/ttn/api/protocol/lorawan"
	pb_lorawan "github.com/TheThingsNetwork/ttn/api/protocol/lorawan"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	"github.com/golang/protobuf/ptypes/empty"
	"golang.org/x/net/context" // See https://github.com/grpc/grpc-go/issues/711"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

type brokerManager struct {
	broker         *broker
	deviceManager  pb_lorawan.DeviceManagerClient
	devAddrManager pb_lorawan.DevAddrManagerClient
}

func (b *brokerManager) GetDevice(ctx context.Context, in *lorawan.DeviceIdentifier) (*lorawan.Device, error) {
	res, err := b.deviceManager.GetDevice(ctx, in)
	if err != nil {
		return nil, errors.BuildGRPCError(errors.Wrap(errors.FromGRPCError(err), "NetworkServer did not return device"))
	}
	return res, nil
}

func (b *brokerManager) SetDevice(ctx context.Context, in *lorawan.Device) (*empty.Empty, error) {
	res, err := b.deviceManager.SetDevice(ctx, in)
	if err != nil {
		return nil, errors.BuildGRPCError(errors.Wrap(errors.FromGRPCError(err), "NetworkServer did not set device"))
	}
	return res, nil
}

func (b *brokerManager) DeleteDevice(ctx context.Context, in *lorawan.DeviceIdentifier) (*empty.Empty, error) {
	res, err := b.deviceManager.DeleteDevice(ctx, in)
	if err != nil {
		return nil, errors.BuildGRPCError(errors.Wrap(errors.FromGRPCError(err), "NetworkServer did not delete device"))
	}
	return res, nil
}

func (b *brokerManager) RegisterApplicationHandler(ctx context.Context, in *pb.ApplicationHandlerRegistration) (*empty.Empty, error) {
	claims, err := b.broker.Component.ValidateTTNAuthContext(ctx)
	if err != nil {
		return nil, errors.BuildGRPCError(errors.FromGRPCError(err))
	}
	if !in.Validate() {
		return nil, grpcErrf(codes.InvalidArgument, "Invalid Application Handler Registration")
	}
	if !claims.AppRight(in.AppId, rights.AppSettings) {
		return nil, grpcErrf(codes.PermissionDenied, "No access to this application")
	}
	// Add Handler in local cache
	handler, err := b.broker.Discovery.Get("handler", in.HandlerId)
	if err != nil {
		return nil, grpcErrf(codes.Internal, "Could not get Handler Announcement")
	}
	handler.AddMetadata(discovery.Metadata_APP_ID, []byte(in.AppId))
	return &empty.Empty{}, nil
}

func (b *brokerManager) GetPrefixes(ctx context.Context, in *lorawan.PrefixesRequest) (*lorawan.PrefixesResponse, error) {
	res, err := b.devAddrManager.GetPrefixes(ctx, in)
	if err != nil {
		return nil, errors.BuildGRPCError(errors.Wrap(errors.FromGRPCError(err), "NetworkServer did not return prefixes"))
	}
	return res, nil
}

func (b *brokerManager) GetDevAddr(ctx context.Context, in *lorawan.DevAddrRequest) (*lorawan.DevAddrResponse, error) {
	res, err := b.devAddrManager.GetDevAddr(ctx, in)
	if err != nil {
		return nil, errors.BuildGRPCError(errors.Wrap(errors.FromGRPCError(err), "NetworkServer did not return DevAddr"))
	}
	return res, nil
}

func (b *brokerManager) GetStatus(ctx context.Context, in *pb.StatusRequest) (*pb.Status, error) {
	return nil, grpcErrf(codes.Unimplemented, "Not Implemented")
}

func (b *broker) RegisterManager(s *grpc.Server) {
	server := &brokerManager{
		broker:         b,
		deviceManager:  pb_lorawan.NewDeviceManagerClient(b.nsConn),
		devAddrManager: pb_lorawan.NewDevAddrManagerClient(b.nsConn),
	}
	pb.RegisterBrokerManagerServer(s, server)
	lorawan.RegisterDeviceManagerServer(s, server)
	lorawan.RegisterDevAddrManagerServer(s, server)
}

// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package broker

import (
	"errors"

	"github.com/TheThingsNetwork/ttn/api"
	pb "github.com/TheThingsNetwork/ttn/api/broker"
	"github.com/TheThingsNetwork/ttn/api/protocol/lorawan"
	pb_lorawan "github.com/TheThingsNetwork/ttn/api/protocol/lorawan"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

type brokerManager struct {
	broker         *broker
	deviceManager  pb_lorawan.DeviceManagerClient
	devAddrManager pb_lorawan.DevAddrManagerClient
}

var errf = grpc.Errorf

func (b *brokerManager) GetDevice(ctx context.Context, in *lorawan.DeviceIdentifier) (*lorawan.Device, error) {
	return b.deviceManager.GetDevice(ctx, in)
}

func (b *brokerManager) SetDevice(ctx context.Context, in *lorawan.Device) (*api.Ack, error) {
	return b.deviceManager.SetDevice(ctx, in)
}

func (b *brokerManager) DeleteDevice(ctx context.Context, in *lorawan.DeviceIdentifier) (*api.Ack, error) {
	return b.deviceManager.DeleteDevice(ctx, in)
}

func (b *brokerManager) RegisterApplicationHandler(ctx context.Context, in *pb.ApplicationHandlerRegistration) (*api.Ack, error) {
	claims, err := b.broker.Component.ValidateTTNAuthContext(ctx)
	if err != nil {
		return nil, err
	}
	if !in.Validate() {
		return nil, grpcErrf(codes.InvalidArgument, "Invalid Application Handler Registration")
	}
	if !claims.CanEditApp(in.AppId) {
		return nil, errf(codes.Unauthenticated, "No access to this application")
	}
	err = b.broker.handlerDiscovery.AddAppID(in.HandlerId, in.AppId)
	if err != nil {
		return nil, err
	}
	return &api.Ack{}, nil
}

func (b *brokerManager) GetPrefixes(ctx context.Context, in *lorawan.PrefixesRequest) (*lorawan.PrefixesResponse, error) {
	return b.devAddrManager.GetPrefixes(ctx, in)
}

func (b *brokerManager) GetDevAddr(ctx context.Context, in *lorawan.DevAddrRequest) (*lorawan.DevAddrResponse, error) {
	return b.devAddrManager.GetDevAddr(ctx, in)
}

func (b *brokerManager) GetStatus(ctx context.Context, in *pb.StatusRequest) (*pb.Status, error) {
	return nil, errors.New("Not Implemented")
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

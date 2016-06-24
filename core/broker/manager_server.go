// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package broker

import (
	"errors"

	"github.com/TheThingsNetwork/ttn/api"
	pb "github.com/TheThingsNetwork/ttn/api/broker"
	"github.com/TheThingsNetwork/ttn/api/protocol/lorawan"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

type brokerManager struct {
	*broker
}

var errf = grpc.Errorf

func (b *brokerManager) GetDevice(ctx context.Context, in *lorawan.DeviceIdentifier) (*lorawan.Device, error) {
	return b.nsManager.GetDevice(ctx, in)
}

func (b *brokerManager) SetDevice(ctx context.Context, in *lorawan.Device) (*api.Ack, error) {
	return b.nsManager.SetDevice(ctx, in)
}

func (b *brokerManager) DeleteDevice(ctx context.Context, in *lorawan.DeviceIdentifier) (*api.Ack, error) {
	return b.nsManager.DeleteDevice(ctx, in)
}

func (b *brokerManager) RegisterApplicationHandler(ctx context.Context, in *pb.ApplicationHandlerRegistration) (*api.Ack, error) {
	claims, err := b.Component.ValidateContext(ctx)
	if err != nil {
		return nil, err
	}
	if !claims.CanEditApp(in.AppId) {
		return nil, errf(codes.Unauthenticated, "No access to this application")
	}
	err = b.handlerDiscovery.AddAppID(in.HandlerId, in.AppId)
	if err != nil {
		return nil, err
	}
	return &api.Ack{}, nil
}

func (b *brokerManager) GetStatus(ctx context.Context, in *pb.StatusRequest) (*pb.Status, error) {
	return nil, errors.New("Not Implemented")
}

func (b *broker) RegisterManager(s *grpc.Server) {
	server := &brokerManager{b}
	pb.RegisterBrokerManagerServer(s, server)
	lorawan.RegisterDeviceManagerServer(s, server)
}

// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package networkserver

import (
	"errors"

	"github.com/TheThingsNetwork/ttn/api/broker"
	"github.com/TheThingsNetwork/ttn/api/handler"
	pb "github.com/TheThingsNetwork/ttn/api/networkserver"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type networkServerRPC struct {
	networkServer NetworkServer
}

func validateBrokerFromMetadata(ctx context.Context) (err error) {
	md, ok := metadata.FromContext(ctx)
	// TODO: Check OK
	id, ok := md["id"]
	if !ok || len(id) < 1 {
		err = errors.New("ttn/networkserver: Broker did not provide \"id\" in context")
		return
	}
	if err != nil {
		return
	}
	token, ok := md["token"]
	if !ok || len(token) < 1 {
		err = errors.New("ttn/networkserver: Broker did not provide \"token\" in context")
		return
	}
	if token[0] != "token" {
		// TODO: Validate Token
		err = errors.New("ttn/networkserver: Broker not authorized")
		return
	}

	return
}

func (s *networkServerRPC) GetDevices(ctx context.Context, req *pb.DevicesRequest) (*pb.DevicesResponse, error) {
	err := validateBrokerFromMetadata(ctx)
	if err != nil {
		return nil, err
	}
	return s.networkServer.HandleGetDevices(req)
}

func (s *networkServerRPC) PrepareActivation(ctx context.Context, activation *broker.DeduplicatedDeviceActivationRequest) (*broker.DeduplicatedDeviceActivationRequest, error) {
	err := validateBrokerFromMetadata(ctx)
	if err != nil {
		return nil, err
	}
	return s.networkServer.HandlePrepareActivation(activation)
}

func (s *networkServerRPC) Activate(ctx context.Context, activation *handler.DeviceActivationResponse) (*handler.DeviceActivationResponse, error) {
	err := validateBrokerFromMetadata(ctx)
	if err != nil {
		return nil, err
	}
	return s.networkServer.HandleActivate(activation)
}

func (s *networkServerRPC) Uplink(ctx context.Context, message *broker.DeduplicatedUplinkMessage) (*broker.DeduplicatedUplinkMessage, error) {
	err := validateBrokerFromMetadata(ctx)
	if err != nil {
		return nil, err
	}
	return s.networkServer.HandleUplink(message)
}

func (s *networkServerRPC) Downlink(ctx context.Context, message *broker.DownlinkMessage) (*broker.DownlinkMessage, error) {
	err := validateBrokerFromMetadata(ctx)
	if err != nil {
		return nil, err
	}
	return s.networkServer.HandleDownlink(message)
}

// RegisterRPC registers this networkserver as a NetworkServerServer (github.com/TheThingsNetwork/ttn/api/networkserver)
func (n *networkServer) RegisterRPC(s *grpc.Server) {
	server := &networkServerRPC{n}
	pb.RegisterNetworkServerServer(s, server)
}

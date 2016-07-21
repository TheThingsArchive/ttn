// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package networkserver

import (
	"errors"

	"github.com/TheThingsNetwork/ttn/api/broker"
	"github.com/TheThingsNetwork/ttn/api/handler"
	pb "github.com/TheThingsNetwork/ttn/api/networkserver"
	"github.com/TheThingsNetwork/ttn/utils/security"
	"github.com/dgrijalva/jwt-go"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
)

type networkServerRPC struct {
	networkServer NetworkServer
}

var grpcErrf = grpc.Errorf // To make go vet stop complaining

func (s *networkServerRPC) ValidateContext(ctx context.Context) error {
	md, ok := metadata.FromContext(ctx)
	if !ok {
		return errors.New("ttn: Could not get metadata")
	}
	var id, token string
	if ids, ok := md["id"]; ok && len(ids) == 1 {
		id = ids[0]
	}
	if id == "" {
		return errors.New("ttn: Could not get id")
	}
	if tokens, ok := md["token"]; ok && len(tokens) == 1 {
		token = tokens[0]
	}
	if token == "" {
		return errors.New("ttn: Could not get token")
	}
	var claims *jwt.StandardClaims
	claims, err := security.ValidateJWT(token, []byte(s.networkServer.(*networkServer).Identity.PublicKey))
	if err != nil {
		return err
	}
	if claims.Subject != id {
		return errors.New("The token was issued for a different component ID")
	}
	return nil
}

func (s *networkServerRPC) GetDevices(ctx context.Context, req *pb.DevicesRequest) (*pb.DevicesResponse, error) {
	if err := s.ValidateContext(ctx); err != nil {
		return nil, err
	}
	if !req.Validate() {
		return nil, grpcErrf(codes.InvalidArgument, "Invalid Devices Request")
	}
	return s.networkServer.HandleGetDevices(req)
}

func (s *networkServerRPC) PrepareActivation(ctx context.Context, activation *broker.DeduplicatedDeviceActivationRequest) (*broker.DeduplicatedDeviceActivationRequest, error) {
	if err := s.ValidateContext(ctx); err != nil {
		return nil, err
	}
	if !activation.Validate() {
		return nil, grpcErrf(codes.InvalidArgument, "Invalid Activation Request")
	}
	return s.networkServer.HandlePrepareActivation(activation)
}

func (s *networkServerRPC) Activate(ctx context.Context, activation *handler.DeviceActivationResponse) (*handler.DeviceActivationResponse, error) {
	if err := s.ValidateContext(ctx); err != nil {
		return nil, err
	}
	if !activation.Validate() {
		return nil, grpcErrf(codes.InvalidArgument, "Invalid Activation Request")
	}
	return s.networkServer.HandleActivate(activation)
}

func (s *networkServerRPC) Uplink(ctx context.Context, message *broker.DeduplicatedUplinkMessage) (*broker.DeduplicatedUplinkMessage, error) {
	if err := s.ValidateContext(ctx); err != nil {
		return nil, err
	}
	if !message.Validate() {
		return nil, grpcErrf(codes.InvalidArgument, "Invalid Uplink")
	}
	return s.networkServer.HandleUplink(message)
}

func (s *networkServerRPC) Downlink(ctx context.Context, message *broker.DownlinkMessage) (*broker.DownlinkMessage, error) {
	if err := s.ValidateContext(ctx); err != nil {
		return nil, err
	}
	if !message.Validate() {
		return nil, grpcErrf(codes.InvalidArgument, "Invalid Downlink")
	}
	return s.networkServer.HandleDownlink(message)
}

// RegisterRPC registers this networkserver as a NetworkServerServer (github.com/TheThingsNetwork/ttn/api/networkserver)
func (n *networkServer) RegisterRPC(s *grpc.Server) {
	server := &networkServerRPC{n}
	pb.RegisterNetworkServerServer(s, server)
}

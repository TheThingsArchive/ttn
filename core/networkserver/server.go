// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package networkserver

import (
	"github.com/TheThingsNetwork/ttn/api/broker"
	"github.com/TheThingsNetwork/ttn/api/handler"
	pb "github.com/TheThingsNetwork/ttn/api/networkserver"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	"github.com/TheThingsNetwork/ttn/utils/security"
	"github.com/dgrijalva/jwt-go"
	"golang.org/x/net/context" // See https://github.com/grpc/grpc-go/issues/711"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type networkServerRPC struct {
	networkServer NetworkServer
}

func (s *networkServerRPC) ValidateContext(ctx context.Context) error {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return errors.NewErrInternal("Could not get metadata from context")
	}
	var id, token string
	if ids, ok := md["id"]; ok && len(ids) == 1 {
		id = ids[0]
	}
	if id == "" {
		return errors.NewErrInvalidArgument("Metadata", "id missing")
	}
	if tokens, ok := md["token"]; ok && len(tokens) == 1 {
		token = tokens[0]
	}
	if token == "" {
		return errors.NewErrInvalidArgument("Metadata", "token missing")
	}
	var claims *jwt.StandardClaims
	claims, err := security.ValidateJWT(token, []byte(s.networkServer.(*networkServer).Identity.PublicKey))
	if err != nil {
		return err
	}
	if claims.Subject != id {
		return errors.NewErrInvalidArgument("Metadata", "token was issued for a different component id")
	}
	return nil
}

func (s *networkServerRPC) GetDevices(ctx context.Context, req *pb.DevicesRequest) (*pb.DevicesResponse, error) {
	if err := s.ValidateContext(ctx); err != nil {
		return nil, err
	}
	if err := req.Validate(); err != nil {
		return nil, errors.Wrap(err, "Invalid Devices Request")
	}
	res, err := s.networkServer.HandleGetDevices(req)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (s *networkServerRPC) PrepareActivation(ctx context.Context, activation *broker.DeduplicatedDeviceActivationRequest) (*broker.DeduplicatedDeviceActivationRequest, error) {
	if err := s.ValidateContext(ctx); err != nil {
		return nil, err
	}
	if err := activation.Validate(); err != nil {
		return nil, errors.Wrap(err, "Invalid Activation Request")
	}
	res, err := s.networkServer.HandlePrepareActivation(activation)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (s *networkServerRPC) Activate(ctx context.Context, activation *handler.DeviceActivationResponse) (*handler.DeviceActivationResponse, error) {
	if err := s.ValidateContext(ctx); err != nil {
		return nil, err
	}
	if err := activation.Validate(); err != nil {
		return nil, errors.Wrap(err, "Invalid Activation Request")
	}
	res, err := s.networkServer.HandleActivate(activation)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (s *networkServerRPC) Uplink(ctx context.Context, message *broker.DeduplicatedUplinkMessage) (*broker.DeduplicatedUplinkMessage, error) {
	if err := s.ValidateContext(ctx); err != nil {
		return nil, err
	}
	if err := message.Validate(); err != nil {
		return nil, errors.Wrap(err, "Invalid Uplink")
	}
	res, err := s.networkServer.HandleUplink(message)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (s *networkServerRPC) Downlink(ctx context.Context, message *broker.DownlinkMessage) (*broker.DownlinkMessage, error) {
	if err := s.ValidateContext(ctx); err != nil {
		return nil, err
	}
	if err := message.Validate(); err != nil {
		return nil, errors.Wrap(err, "Invalid Downlink")
	}
	res, err := s.networkServer.HandleDownlink(message)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// RegisterRPC registers this networkserver as a NetworkServerServer (github.com/TheThingsNetwork/ttn/api/networkserver)
func (n *networkServer) RegisterRPC(s *grpc.Server) {
	server := &networkServerRPC{n}
	pb.RegisterNetworkServerServer(s, server)
}

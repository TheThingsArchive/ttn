// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package handler

import (
	pb_broker "github.com/TheThingsNetwork/ttn/api/broker"
	pb "github.com/TheThingsNetwork/ttn/api/handler"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

type handlerRPC struct {
	handler Handler
}

var grpcErrf = grpc.Errorf // To make go vet stop complaining

func (h *handlerRPC) ActivationChallenge(ctx context.Context, challenge *pb_broker.ActivationChallengeRequest) (*pb_broker.ActivationChallengeResponse, error) {
	_, err := h.handler.ValidateNetworkContext(ctx)
	if err != nil {
		return nil, errors.BuildGRPCError(err)
	}
	if !challenge.Validate() {
		return nil, grpcErrf(codes.InvalidArgument, "Invalid Activation Request")
	}
	res, err := h.handler.HandleActivationChallenge(challenge)
	if err != nil {
		return nil, errors.BuildGRPCError(err)
	}
	return res, nil
}

func (h *handlerRPC) Activate(ctx context.Context, activation *pb_broker.DeduplicatedDeviceActivationRequest) (*pb.DeviceActivationResponse, error) {
	_, err := h.handler.ValidateNetworkContext(ctx)
	if err != nil {
		return nil, errors.BuildGRPCError(err)
	}
	if !activation.Validate() {
		return nil, grpcErrf(codes.InvalidArgument, "Invalid Activation Request")
	}
	res, err := h.handler.HandleActivation(activation)
	if err != nil {
		return nil, errors.BuildGRPCError(err)
	}
	return res, nil
}

// RegisterRPC registers this handler as a HandlerServer (github.com/TheThingsNetwork/ttn/api/handler)
func (h *handler) RegisterRPC(s *grpc.Server) {
	server := &handlerRPC{h}
	pb.RegisterHandlerServer(s, server)
}

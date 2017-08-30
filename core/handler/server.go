// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package handler

import (
	pb_broker "github.com/TheThingsNetwork/api/broker"
	pb "github.com/TheThingsNetwork/api/handler"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	"golang.org/x/net/context" // See https://github.com/grpc/grpc-go/issues/711"
	"google.golang.org/grpc"
)

type handlerRPC struct {
	handler Handler
}

func (h *handlerRPC) ActivationChallenge(ctx context.Context, challenge *pb_broker.ActivationChallengeRequest) (*pb_broker.ActivationChallengeResponse, error) {
	_, err := h.handler.ValidateNetworkContext(ctx)
	if err != nil {
		return nil, err
	}
	if err := challenge.Validate(); err != nil {
		return nil, errors.Wrap(err, "Invalid Activation Challenge Request")
	}
	res, err := h.handler.HandleActivationChallenge(challenge)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (h *handlerRPC) Activate(ctx context.Context, activation *pb_broker.DeduplicatedDeviceActivationRequest) (*pb.DeviceActivationResponse, error) {
	_, err := h.handler.ValidateNetworkContext(ctx)
	if err != nil {
		return nil, err
	}
	if err := activation.Validate(); err != nil {
		return nil, errors.Wrap(err, "Invalid Activation Request")
	}
	res, err := h.handler.HandleActivation(activation)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// RegisterRPC registers this handler as a HandlerServer (github.com/TheThingsNetwork/api/handler)
func (h *handler) RegisterRPC(s *grpc.Server) {
	server := &handlerRPC{h}
	pb.RegisterHandlerServer(s, server)
}

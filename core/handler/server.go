// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package handler

import (
	pb_broker "github.com/TheThingsNetwork/ttn/api/broker"
	pb "github.com/TheThingsNetwork/ttn/api/handler"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

type handlerRPC struct {
	handler Handler
}

var grpcErrf = grpc.Errorf // To make go vet stop complaining

func (h *handlerRPC) Activate(ctx context.Context, activation *pb_broker.DeduplicatedDeviceActivationRequest) (*pb.DeviceActivationResponse, error) {
	_, err := h.handler.ValidateNetworkContext(ctx)
	if err != nil {
		return nil, err
	}
	if !activation.Validate() {
		return nil, grpcErrf(codes.InvalidArgument, "Invalid Activation Request")
	}
	return h.handler.HandleActivation(activation)
}

// RegisterRPC registers this handler as a HandlerServer (github.com/TheThingsNetwork/ttn/api/handler)
func (r *handler) RegisterRPC(s *grpc.Server) {
	server := &handlerRPC{r}
	pb.RegisterHandlerServer(s, server)
}

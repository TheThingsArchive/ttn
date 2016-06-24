// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package handler

import (
	"errors"

	pb_broker "github.com/TheThingsNetwork/ttn/api/broker"
	pb "github.com/TheThingsNetwork/ttn/api/handler"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type handlerRPC struct {
	handler Handler
}

func validateBrokerFromMetadata(ctx context.Context) (err error) {
	md, ok := metadata.FromContext(ctx)
	// TODO: Check OK
	id, ok := md["id"]
	if !ok || len(id) < 1 {
		err = errors.New("ttn/handler: Broker did not provide \"id\" in context")
		return
	}
	if err != nil {
		return
	}
	token, ok := md["token"]
	if !ok || len(token) < 1 {
		err = errors.New("ttn/handler: Broker did not provide \"token\" in context")
		return
	}
	if token[0] != "token" {
		// TODO: Validate Token
		err = errors.New("ttn/handler: Broker not authorized")
		return
	}

	return
}

func (h *handlerRPC) Activate(ctx context.Context, activation *pb_broker.DeduplicatedDeviceActivationRequest) (*pb.DeviceActivationResponse, error) {
	err := validateBrokerFromMetadata(ctx)
	if err != nil {
		return nil, err
	}
	return h.handler.HandleActivation(activation)
}

// RegisterRPC registers this handler as a HandlerServer (github.com/TheThingsNetwork/ttn/api/handler)
func (r *handler) RegisterRPC(s *grpc.Server) {
	server := &handlerRPC{r}
	pb.RegisterHandlerServer(s, server)
}

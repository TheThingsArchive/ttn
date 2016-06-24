// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package router

import (
	"errors"

	pb "github.com/TheThingsNetwork/ttn/api/router"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

type routerManager struct {
	*router
}

var errf = grpc.Errorf

func (r *routerManager) GatewayStatus(ctx context.Context, in *pb.GatewayStatusRequest) (*pb.GatewayStatusResponse, error) {
	if in.GatewayEui == nil {
		return nil, errf(codes.InvalidArgument, "GatewayEUI is required")
	}
	_, err := r.ValidateContext(ctx)
	if err != nil {
		return nil, errf(codes.Unauthenticated, "No access")
	}
	gtw := r.getGateway(*in.GatewayEui)
	status, err := gtw.Status.Get()
	if err != nil {
		return nil, err
	}
	return &pb.GatewayStatusResponse{
		LastSeen: gtw.LastSeen.UnixNano(),
		Status:   status,
	}, nil
}

func (r *routerManager) GetStatus(ctx context.Context, in *pb.StatusRequest) (*pb.Status, error) {
	return nil, errors.New("Not Implemented")
}

// RegisterManager registers this router as a RouterManagerServer (github.com/TheThingsNetwork/ttn/api/router)
func (r *router) RegisterManager(s *grpc.Server) {
	server := &routerManager{r}
	pb.RegisterRouterManagerServer(s, server)
}

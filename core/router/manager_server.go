// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package router

import (
	"fmt"

	pb "github.com/TheThingsNetwork/ttn/api/router"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	"golang.org/x/net/context" // See https://github.com/grpc/grpc-go/issues/711"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

type routerManager struct {
	router *router
}

func (r *routerManager) GatewayStatus(ctx context.Context, in *pb.GatewayStatusRequest) (*pb.GatewayStatusResponse, error) {
	if in.GatewayId == "" {
		return nil, errors.NewErrInvalidArgument("Gateway Status Request", "ID is required")
	}
	_, err := r.router.ValidateTTNAuthContext(ctx)
	if err != nil {
		return nil, errors.NewErrPermissionDenied("No access")
	}
	r.router.gatewaysLock.RLock()
	gtw, ok := r.router.gateways[in.GatewayId]
	r.router.gatewaysLock.RUnlock()
	if !ok {
		return nil, errors.NewErrNotFound(fmt.Sprintf("Gateway %s", in.GatewayId))
	}
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
	return nil, grpc.Errorf(codes.Unimplemented, "Not Implemented")
}

// RegisterManager registers this router as a RouterManagerServer (github.com/TheThingsNetwork/ttn/api/router)
func (r *router) RegisterManager(s *grpc.Server) {
	server := &routerManager{r}
	pb.RegisterRouterManagerServer(s, server)
}

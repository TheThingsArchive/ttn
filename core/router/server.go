// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package router

import (
	"fmt"
	"io"

	"github.com/TheThingsNetwork/go-account-lib/claims"
	"github.com/TheThingsNetwork/ttn/api"
	pb "github.com/TheThingsNetwork/ttn/api/router"
	"github.com/TheThingsNetwork/ttn/core/router/gateway"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/spf13/viper"
	context "golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

type routerRPC struct {
	router *router
}

var grpcErrf = grpc.Errorf // To make go vet stop complaining

func (r *routerRPC) gatewayFromContext(ctx context.Context) (gtw *gateway.Gateway, err error) {
	md, err := api.MetadataFromContext(ctx)
	if err != nil {
		return nil, err
	}

	gatewayID, err := api.IDFromMetadata(md)
	if err != nil {
		return nil, err
	}

	token, _ := api.TokenFromMetadata(md)

	if !viper.GetBool("router.skip-verify-gateway-token") {
		if token == "" {
			return nil, errors.NewErrPermissionDenied("No gateway token supplied")
		}
		if r.router.TokenKeyProvider == nil {
			return nil, errors.NewErrInternal("No token provider configured")
		}
		claims, err := claims.FromToken(r.router.TokenKeyProvider, token)
		if err != nil {
			return nil, errors.NewErrPermissionDenied(fmt.Sprintf("Gateway token invalid: %s", err.Error()))
		}
		if claims.Type != "gateway" || claims.Subject != gatewayID {
			return nil, errors.NewErrPermissionDenied("Gateway token not consistent")
		}
	}

	gtw = r.router.getGateway(gatewayID)
	gtw.SetToken(token)

	return gtw, nil
}

// GatewayStatus implements RouterServer interface (github.com/TheThingsNetwork/ttn/api/router)
func (r *routerRPC) GatewayStatus(stream pb.Router_GatewayStatusServer) error {
	gateway, err := r.gatewayFromContext(stream.Context())
	if err != nil {
		return errors.BuildGRPCError(err)
	}

	for {
		status, err := stream.Recv()
		if err == io.EOF {
			return stream.SendAndClose(&empty.Empty{})
		}
		if err != nil {
			return err
		}
		if !status.Validate() {
			return grpcErrf(codes.InvalidArgument, "Invalid Gateway Status")
		}
		go r.router.HandleGatewayStatus(gateway.ID, status)
	}
}

// Uplink implements RouterServer interface (github.com/TheThingsNetwork/ttn/api/router)
func (r *routerRPC) Uplink(stream pb.Router_UplinkServer) error {
	gateway, err := r.gatewayFromContext(stream.Context())
	if err != nil {
		return errors.BuildGRPCError(err)
	}

	for {
		uplink, err := stream.Recv()
		if err == io.EOF {
			return stream.SendAndClose(&empty.Empty{})
		}
		if err != nil {
			return err
		}
		if !uplink.Validate() {
			return grpcErrf(codes.InvalidArgument, "Invalid Uplink")
		}
		go r.router.HandleUplink(gateway.ID, uplink)
	}
}

// Subscribe implements RouterServer interface (github.com/TheThingsNetwork/ttn/api/router)
func (r *routerRPC) Subscribe(req *pb.SubscribeRequest, stream pb.Router_SubscribeServer) error {
	gateway, err := r.gatewayFromContext(stream.Context())
	if err != nil {
		return errors.BuildGRPCError(err)
	}

	downlinkChannel, err := r.router.SubscribeDownlink(gateway.ID)
	if err != nil {
		return errors.BuildGRPCError(err)
	}
	defer r.router.UnsubscribeDownlink(gateway.ID)

	for {
		if downlinkChannel == nil {
			return nil
		}
		select {
		case <-stream.Context().Done():
			return stream.Context().Err()
		case downlink := <-downlinkChannel:
			if err := stream.Send(downlink); err != nil {
				return err
			}
		}
	}
}

// Activate implements RouterServer interface (github.com/TheThingsNetwork/ttn/api/router)
func (r *routerRPC) Activate(ctx context.Context, req *pb.DeviceActivationRequest) (*pb.DeviceActivationResponse, error) {
	gateway, err := r.gatewayFromContext(ctx)
	if err != nil {
		return nil, errors.BuildGRPCError(err)
	}

	if !req.Validate() {
		return nil, grpcErrf(codes.InvalidArgument, "Invalid Activation Request")
	}
	return r.router.HandleActivation(gateway.ID, req)
}

// RegisterRPC registers this router as a RouterServer (github.com/TheThingsNetwork/ttn/api/router)
func (r *router) RegisterRPC(s *grpc.Server) {
	server := &routerRPC{r}
	pb.RegisterRouterServer(s, server)
}

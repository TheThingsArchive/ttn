// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package router

import (
	"fmt"

	"github.com/TheThingsNetwork/go-account-lib/claims"
	"github.com/TheThingsNetwork/ttn/api"
	pb_gateway "github.com/TheThingsNetwork/ttn/api/gateway"
	pb "github.com/TheThingsNetwork/ttn/api/router"
	"github.com/TheThingsNetwork/ttn/core/router/gateway"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	"github.com/spf13/viper"
	"golang.org/x/net/context" // See https://github.com/grpc/grpc-go/issues/711"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type routerRPC struct {
	router *router
	pb.RouterStreamServer
}

func (r *routerRPC) gatewayFromMetadata(md metadata.MD) (gtw *gateway.Gateway, err error) {
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

func (r *routerRPC) gatewayFromContext(ctx context.Context) (gtw *gateway.Gateway, err error) {
	md, err := api.MetadataFromContext(ctx)
	if err != nil {
		return nil, err
	}
	return r.gatewayFromMetadata(md)
}

func (r *routerRPC) getUplink(md metadata.MD) (ch chan *pb.UplinkMessage, err error) {
	gateway, err := r.gatewayFromMetadata(md)
	if err != nil {
		return nil, err
	}
	ch = make(chan *pb.UplinkMessage)
	go func() {
		for uplink := range ch {
			r.router.HandleUplink(gateway.ID, uplink)
		}
	}()
	return
}

func (r *routerRPC) getGatewayStatus(md metadata.MD) (ch chan *pb_gateway.Status, err error) {
	gateway, err := r.gatewayFromMetadata(md)
	if err != nil {
		return nil, err
	}
	ch = make(chan *pb_gateway.Status)
	go func() {
		for status := range ch {
			r.router.HandleGatewayStatus(gateway.ID, status)
		}
	}()
	return
}

func (r *routerRPC) getDownlink(md metadata.MD) (ch <-chan *pb.DownlinkMessage, cancel func(), err error) {
	gateway, err := r.gatewayFromMetadata(md)
	if err != nil {
		return nil, nil, err
	}
	ch = make(chan *pb.DownlinkMessage)
	cancel = func() {
		r.router.UnsubscribeDownlink(gateway.ID)
	}
	downlinkChannel, err := r.router.SubscribeDownlink(gateway.ID)
	if err != nil {
		return nil, nil, err
	}
	return downlinkChannel, cancel, nil
}

// Activate implements RouterServer interface (github.com/TheThingsNetwork/ttn/api/router)
func (r *routerRPC) Activate(ctx context.Context, req *pb.DeviceActivationRequest) (*pb.DeviceActivationResponse, error) {
	gateway, err := r.gatewayFromContext(ctx)
	if err != nil {
		return nil, errors.BuildGRPCError(err)
	}
	if err := req.Validate(); err != nil {
		return nil, errors.BuildGRPCError(errors.Wrap(err, "Invalid Activation Request"))
	}
	return r.router.HandleActivation(gateway.ID, req)
}

// RegisterRPC registers this router as a RouterServer (github.com/TheThingsNetwork/ttn/api/router)
func (r *router) RegisterRPC(s *grpc.Server) {
	server := &routerRPC{router: r}
	server.SetLogger(api.Apex(r.Ctx))
	server.UplinkChanFunc = server.getUplink
	server.DownlinkChanFunc = server.getDownlink
	server.GatewayStatusChanFunc = server.getGatewayStatus
	pb.RegisterRouterServer(s, server)
}

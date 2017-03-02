// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package router

import (
	"fmt"
	"time"

	"github.com/TheThingsNetwork/go-account-lib/claims"
	"github.com/TheThingsNetwork/ttn/api"
	pb_gateway "github.com/TheThingsNetwork/ttn/api/gateway"
	"github.com/TheThingsNetwork/ttn/api/ratelimit"
	pb "github.com/TheThingsNetwork/ttn/api/router"
	"github.com/TheThingsNetwork/ttn/core/router/gateway"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	"github.com/TheThingsNetwork/ttn/utils/random"
	"github.com/spf13/viper"
	"golang.org/x/net/context" // See https://github.com/grpc/grpc-go/issues/711"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
)

type routerRPC struct {
	router *router
	pb.RouterStreamServer

	uplinkRate *ratelimit.Registry
	statusRate *ratelimit.Registry
}

func (r *routerRPC) gatewayFromMetadata(md metadata.MD) (gtw *gateway.Gateway, err error) {
	gatewayID, err := api.IDFromMetadata(md)
	if err != nil {
		return nil, err
	}

	authErr := errors.NewErrPermissionDenied("Gateway not authenticated")
	authenticated := false
	token, _ := api.TokenFromMetadata(md)

	if token != "" {
		if r.router.TokenKeyProvider == nil {
			return nil, errors.NewErrInternal("No token provider configured")
		}
		claims, err := claims.FromGatewayToken(r.router.TokenKeyProvider, token)
		if err != nil {
			authErr = errors.NewErrPermissionDenied(fmt.Sprintf("Gateway token invalid: %s", err))
		} else {
			if claims.Subject != gatewayID {
				authErr = errors.NewErrPermissionDenied(fmt.Sprintf("Token subject \"%s\" not consistent with gateway ID \"%s\"", claims.Subject, gatewayID))
			} else {
				authErr = nil
				authenticated = true
			}
		}
	}

	if authErr != nil && !viper.GetBool("router.skip-verify-gateway-token") {
		return nil, authErr
	}

	gtw = r.router.getGateway(gatewayID)
	gtw.SetAuth(token, authenticated)

	return gtw, nil
}

func (r *routerRPC) gatewayFromContext(ctx context.Context) (gtw *gateway.Gateway, err error) {
	md := api.MetadataFromContext(ctx)
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
			if waitTime := r.uplinkRate.Wait(gateway.ID); waitTime != 0 {
				r.router.Ctx.WithField("GatewayID", gateway.ID).WithField("Wait", waitTime).Warn("Gateway reached uplink rate limit")
				time.Sleep(waitTime)
			}
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
			if waitTime := r.statusRate.Wait(gateway.ID); waitTime != 0 {
				r.router.Ctx.WithField("GatewayID", gateway.ID).WithField("Wait", waitTime).Warn("Gateway reached status rate limit")
				time.Sleep(waitTime)
			}
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
	subscriptionID := random.String(10)
	ch = make(chan *pb.DownlinkMessage)
	cancel = func() {
		r.router.UnsubscribeDownlink(gateway.ID, subscriptionID)
	}
	downlinkChannel, err := r.router.SubscribeDownlink(gateway.ID, subscriptionID)
	if err != nil {
		return nil, nil, err
	}
	return downlinkChannel, cancel, nil
}

// Activate implements RouterServer interface (github.com/TheThingsNetwork/ttn/api/router)
func (r *routerRPC) Activate(ctx context.Context, req *pb.DeviceActivationRequest) (*pb.DeviceActivationResponse, error) {
	gateway, err := r.gatewayFromContext(ctx)
	if err != nil {
		return nil, err
	}
	if err := req.Validate(); err != nil {
		return nil, errors.Wrap(err, "Invalid Activation Request")
	}
	if r.uplinkRate.Limit(gateway.ID) {
		return nil, grpc.Errorf(codes.ResourceExhausted, "Gateway reached uplink rate limit")
	}
	return r.router.HandleActivation(gateway.ID, req)
}

// RegisterRPC registers this router as a RouterServer (github.com/TheThingsNetwork/ttn/api/router)
func (r *router) RegisterRPC(s *grpc.Server) {
	server := &routerRPC{router: r}
	server.SetLogger(r.Ctx)
	server.UplinkChanFunc = server.getUplink
	server.DownlinkChanFunc = server.getDownlink
	server.GatewayStatusChanFunc = server.getGatewayStatus

	// TODO: Monitor actual rates and configure sensible limits
	//
	// The current values are based on the following:
	// - 20 byte messages on all 6 orthogonal SFs at the same time -> ~1500 msgs/minute
	// - 8 channels at 5% utilization: 600 msgs/minute
	// - let's double that and round it to 1500/minute

	server.uplinkRate = ratelimit.NewRegistry(1500, time.Minute) // includes activations
	server.statusRate = ratelimit.NewRegistry(10, time.Minute)   // 10 per minute (pkt fwd default is 2 per minute)

	pb.RegisterRouterServer(s, server)
}

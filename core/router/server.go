// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package router

import (
	"fmt"
	"io"
	"time"

	pb "github.com/TheThingsNetwork/api/router"
	"github.com/TheThingsNetwork/api/trace"
	"github.com/TheThingsNetwork/go-account-lib/claims"
	"github.com/TheThingsNetwork/go-utils/grpc/ttnctx"
	"github.com/TheThingsNetwork/ttn/api/ratelimit"
	"github.com/TheThingsNetwork/ttn/core/router/gateway"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	"github.com/TheThingsNetwork/ttn/utils/random"
	"github.com/gogo/protobuf/types"
	"github.com/spf13/viper"
	"golang.org/x/net/context" // See https://github.com/grpc/grpc-go/issues/711"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
)

type routerRPC struct {
	router *router

	uplinkRate *ratelimit.Registry
	statusRate *ratelimit.Registry
}

func (r *routerRPC) gatewayFromMetadata(md metadata.MD) (gtw *gateway.Gateway, err error) {
	gatewayID, err := ttnctx.IDFromMetadata(md)
	if err != nil {
		return nil, err
	}

	authErr := errors.NewErrPermissionDenied("Gateway not authenticated")
	authenticated := false
	token, _ := ttnctx.TokenFromMetadata(md)

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
	if authenticated {
		gtw.SetAuth(token, authenticated)
	}

	return gtw, nil
}

func (r *routerRPC) gatewayFromContext(ctx context.Context) (gtw *gateway.Gateway, err error) {
	md := ttnctx.MetadataFromIncomingContext(ctx)
	return r.gatewayFromMetadata(md)
}

// Uplink handles uplink streams
func (r *routerRPC) Uplink(stream pb.Router_UplinkServer) error {
	ctx := stream.Context()
	gateway, err := r.gatewayFromContext(ctx)
	if err != nil {
		return err
	}
	logger := r.router.Ctx.WithField("GatewayID", gateway.ID)
	if err = stream.SendHeader(metadata.MD{}); err != nil {
		return err
	}
	for {
		uplink, err := stream.Recv()
		if err == io.EOF {
			return stream.SendAndClose(&types.Empty{})
		}
		if err != nil {
			return err
		}
		if err := uplink.Validate(); err != nil {
			logger.WithError(err).Warn("Invalid Uplink")
			continue
		}
		if err := uplink.UnmarshalPayload(); err != nil {
			logger.WithError(err).Warn("Could not unmarshal Uplink payload")
		}
		if waitTime := r.uplinkRate.Wait(gateway.ID); waitTime != 0 {
			logger.WithField("Wait", waitTime).Warn("Gateway reached uplink rate limit")
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(waitTime):
			}
		}
		if err := r.router.HandleUplink(gateway.ID, uplink); err != nil {
			logger.WithError(err).Warn("Failed to handle uplink")
		}
	}
}

// GatewayStatus handles gateway status streams
func (r *routerRPC) GatewayStatus(stream pb.Router_GatewayStatusServer) error {
	ctx := stream.Context()
	gateway, err := r.gatewayFromContext(ctx)
	if err != nil {
		return err
	}
	logger := r.router.Ctx.WithField("GatewayID", gateway.ID)
	if err := stream.SendHeader(metadata.MD{}); err != nil {
		return err
	}
	for {
		status, err := stream.Recv()
		if err == io.EOF {
			return stream.SendAndClose(&types.Empty{})
		}
		if err != nil {
			return err
		}
		if err := status.Validate(); err != nil {
			return errors.Wrap(err, "Invalid Gateway Status")
		}
		if waitTime := r.statusRate.Wait(gateway.ID); waitTime != 0 {
			logger.WithField("Wait", waitTime).Warn("Gateway reached status rate limit")
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(waitTime):
			}
		}
		if err := r.router.HandleGatewayStatus(gateway.ID, status); err != nil {
			logger.WithError(err).Warn("Failed to handle gateway status")
		}
	}
}

// Subscribe handles downlink streams
func (r *routerRPC) Subscribe(req *pb.SubscribeRequest, stream pb.Router_SubscribeServer) error {
	ctx := stream.Context()
	gateway, err := r.gatewayFromContext(ctx)
	if err != nil {
		return err
	}
	subscriptionID := random.String(16)
	downlinks, err := r.router.SubscribeDownlink(gateway.ID, subscriptionID)
	if err != nil {
		return err
	}
	defer r.router.UnsubscribeDownlink(gateway.ID, subscriptionID)
	if err = stream.SendHeader(metadata.MD{}); err != nil {
		return err
	}
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case downlink, ok := <-downlinks:
			if !ok {
				return nil
			}
			if err := stream.Send(downlink); err != nil {
				return err
			}
			if gateway.MonitorStream != nil {
				clone := *downlink // There can be multiple subscribers
				clone.Trace = clone.Trace.WithEvent(trace.SendEvent)
				gateway.MonitorStream.Send(&clone)
			}
		}
	}
}

// Activate implements RouterServer interface (github.com/TheThingsNetwork/api/router)
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

// RegisterRPC registers this router as a RouterServer (github.com/TheThingsNetwork/api/router)
func (r *router) RegisterRPC(s *grpc.Server) {
	server := &routerRPC{router: r}
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

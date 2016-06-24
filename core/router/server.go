// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package router

import (
	"errors"
	"io"

	api "github.com/TheThingsNetwork/ttn/api"
	pb "github.com/TheThingsNetwork/ttn/api/router"
	"github.com/TheThingsNetwork/ttn/core/types"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type routerRPC struct {
	router Router
}

func getGatewayFromMetadata(ctx context.Context) (gatewayEUI types.GatewayEUI, err error) {
	md, ok := metadata.FromContext(ctx)
	// TODO: Check OK
	euiString, ok := md["gateway_eui"]
	if !ok || len(euiString) < 1 {
		err = errors.New("ttn/router: Gateway did not provide \"gateway_eui\" in context")
		return
	}
	gatewayEUI, err = types.ParseGatewayEUI(euiString[0])
	if err != nil {
		return
	}
	token, ok := md["token"]
	if !ok || len(token) < 1 {
		err = errors.New("ttn/router: Gateway did not provide \"token\" in context")
		return
	}
	if token[0] != "token" {
		// TODO: Validate Token
		err = errors.New("ttn/router: Gateway not authorized")
		return
	}

	return
}

// GatewayStatus implements RouterServer interface (github.com/TheThingsNetwork/ttn/api/router)
func (r *routerRPC) GatewayStatus(stream pb.Router_GatewayStatusServer) error {
	gatewayEUI, err := getGatewayFromMetadata(stream.Context())
	if err != nil {
		return err
	}
	for {
		status, err := stream.Recv()
		if err == io.EOF {
			return stream.SendAndClose(&api.Ack{})
		}
		if err != nil {
			return err
		}
		go r.router.HandleGatewayStatus(gatewayEUI, status)
	}
}

// Uplink implements RouterServer interface (github.com/TheThingsNetwork/ttn/api/router)
func (r *routerRPC) Uplink(stream pb.Router_UplinkServer) error {
	gatewayEUI, err := getGatewayFromMetadata(stream.Context())
	if err != nil {
		return err
	}
	for {
		uplink, err := stream.Recv()
		if err == io.EOF {
			return stream.SendAndClose(&api.Ack{})
		}
		if err != nil {
			return err
		}
		go r.router.HandleUplink(gatewayEUI, uplink)
	}
}

// Subscribe implements RouterServer interface (github.com/TheThingsNetwork/ttn/api/router)
func (r *routerRPC) Subscribe(req *pb.SubscribeRequest, stream pb.Router_SubscribeServer) error {
	gatewayEUI, err := getGatewayFromMetadata(stream.Context())
	if err != nil {
		return err
	}
	downlinkChannel, err := r.router.SubscribeDownlink(gatewayEUI)
	if err != nil {
		return err
	}
	defer r.router.UnsubscribeDownlink(gatewayEUI)
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
	gatewayEUI, err := getGatewayFromMetadata(ctx)
	if err != nil {
		return nil, err
	}
	return r.router.HandleActivation(gatewayEUI, req)
}

// RegisterRPC registers this router as a RouterServer (github.com/TheThingsNetwork/ttn/api/router)
func (r *router) RegisterRPC(s *grpc.Server) {
	server := &routerRPC{r}
	pb.RegisterRouterServer(s, server)
}

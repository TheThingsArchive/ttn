// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package broker

import (
	"io"

	pb "github.com/TheThingsNetwork/ttn/api/broker"
	pb_discovery "github.com/TheThingsNetwork/ttn/api/discovery"
	"github.com/TheThingsNetwork/ttn/core"
	"github.com/golang/protobuf/ptypes/empty"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

type brokerRPC struct {
	broker *broker
}

var grpcErrf = grpc.Errorf // To make go vet stop complaining

func (b *brokerRPC) Associate(stream pb.Broker_AssociateServer) error {
	router, err := b.broker.ValidateNetworkContext(stream.Context())
	if err != nil {
		return core.BuildGRPCError(err)
	}
	downlinkChannel, err := b.broker.ActivateRouter(router.Id)
	if err != nil {
		return core.BuildGRPCError(err)
	}
	defer b.broker.DeactivateRouter(router.Id)
	go func() {
		for {
			if downlinkChannel == nil {
				return
			}
			select {
			case <-stream.Context().Done():
				return
			case downlink := <-downlinkChannel:
				if downlink != nil {
					if err := stream.Send(downlink); err != nil {
						// TODO: Check if the stream should be closed here
						return
					}
				}
			}
		}
	}()
	for {
		uplink, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		if !uplink.Validate() {
			return grpcErrf(codes.InvalidArgument, "Invalid Uplink")
		}
		go b.broker.HandleUplink(uplink)
	}
}

func (b *brokerRPC) Subscribe(req *pb.SubscribeRequest, stream pb.Broker_SubscribeServer) error {
	handler, err := b.broker.ValidateNetworkContext(stream.Context())
	if err != nil {
		return core.BuildGRPCError(err)
	}
	uplinkChannel, err := b.broker.ActivateHandler(handler.Id)
	if err != nil {
		return core.BuildGRPCError(err)
	}
	defer b.broker.DeactivateHandler(handler.Id)
	for {
		if uplinkChannel == nil {
			return nil
		}
		select {
		case <-stream.Context().Done():
			return stream.Context().Err()
		case uplink := <-uplinkChannel:
			if uplink != nil {
				if err := stream.Send(uplink); err != nil {
					return err
				}
			}
		}
	}
}

func (b *brokerRPC) Publish(stream pb.Broker_PublishServer) error {
	handler, err := b.broker.ValidateNetworkContext(stream.Context())
	if err != nil {
		return core.BuildGRPCError(err)
	}
	for {
		downlink, err := stream.Recv()
		if err == io.EOF {
			return stream.SendAndClose(&empty.Empty{})
		}
		if err != nil {
			return err
		}
		if !downlink.Validate() {
			return grpcErrf(codes.InvalidArgument, "Invalid Downlink")
		}
		go func(downlink *pb.DownlinkMessage) {
			// Get latest Handler metadata
			handler, err := b.broker.Component.Discover("handler", handler.Id)
			if err != nil {
				return
			}
			// Check if this Handler can publish for this AppId
			for _, meta := range handler.Metadata {
				switch meta.Key {
				case pb_discovery.Metadata_APP_ID:
					announcedID := string(meta.Value)
					if announcedID == downlink.AppId {
						b.broker.HandleDownlink(downlink)
						return
					}
				}
			}
		}(downlink)
	}
}

func (b *brokerRPC) Activate(ctx context.Context, req *pb.DeviceActivationRequest) (res *pb.DeviceActivationResponse, err error) {
	_, err = b.broker.ValidateNetworkContext(ctx)
	if err != nil {
		return nil, core.BuildGRPCError(err)
	}
	if !req.Validate() {
		return nil, grpcErrf(codes.InvalidArgument, "Invalid Activation Request")
	}
	res, err = b.broker.HandleActivation(req)
	if err != nil {
		return nil, core.BuildGRPCError(err)
	}
	return
}

func (b *broker) RegisterRPC(s *grpc.Server) {
	server := &brokerRPC{b}
	pb.RegisterBrokerServer(s, server)
}

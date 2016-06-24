// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package broker

import (
	"io"

	pb_api "github.com/TheThingsNetwork/ttn/api"
	pb "github.com/TheThingsNetwork/ttn/api/broker"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
)

type brokerRPC struct {
	broker Broker
}

var grpcErrf = grpc.Errorf // To make go vet stop complaining

func getCallerFromMetadata(ctx context.Context) (callerID string, err error) {
	md, ok := metadata.FromContext(ctx)
	// TODO: Check OK
	id, ok := md["id"]
	if !ok || len(id) < 1 {
		err = grpcErrf(codes.Unauthenticated, "ttn/broker: Caller did not provide \"id\" in context")
		return
	}
	callerID = id[0]
	if err != nil {
		return
	}
	token, ok := md["token"]
	if !ok || len(token) < 1 {
		err = grpcErrf(codes.Unauthenticated, "ttn/broker: Caller did not provide \"token\" in context")
		return
	}
	if token[0] != "token" {
		// TODO: Validate Token
		err = grpcErrf(codes.Unauthenticated, "ttn/broker: Caller not authorized")
		return
	}

	return
}

func (b *brokerRPC) Associate(stream pb.Broker_AssociateServer) error {
	routerID, err := getCallerFromMetadata(stream.Context())
	if err != nil {
		return err
	}
	downlinkChannel, err := b.broker.ActivateRouter(routerID)
	if err != nil {
		return err
	}
	defer b.broker.DeactivateRouter(routerID)
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
		go b.broker.HandleUplink(uplink)
	}
}

func (b *brokerRPC) Subscribe(req *pb.SubscribeRequest, stream pb.Broker_SubscribeServer) error {
	handlerID, err := getCallerFromMetadata(stream.Context())
	if err != nil {
		return err
	}
	uplinkChannel, err := b.broker.ActivateHandler(handlerID)
	if err != nil {
		return err
	}
	defer b.broker.DeactivateHandler(handlerID)
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
	handlerID, err := getCallerFromMetadata(stream.Context())
	if err != nil {
		return err
	}
	for {
		downlink, err := stream.Recv()
		if err == io.EOF {
			return stream.SendAndClose(&pb_api.Ack{})
		}
		if err != nil {
			return err
		}
		// TODO: Validate that this handler can publish downlink for the application
		_ = handlerID
		go b.broker.HandleDownlink(downlink)
	}
}

func (b *brokerRPC) Activate(ctx context.Context, req *pb.DeviceActivationRequest) (res *pb.DeviceActivationResponse, err error) {
	_, err = getCallerFromMetadata(ctx)
	if err != nil {
		return nil, err
	}
	return b.broker.HandleActivation(req)
}

func (b *broker) RegisterRPC(s *grpc.Server) {
	server := &brokerRPC{b}
	pb.RegisterBrokerServer(s, server)
}

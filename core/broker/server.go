// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package broker

import (
	"time"

	pb "github.com/TheThingsNetwork/ttn/api/broker"
	"github.com/TheThingsNetwork/ttn/api/ratelimit"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	"golang.org/x/net/context" // See https://github.com/grpc/grpc-go/issues/711"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
)

type brokerRPC struct {
	broker *broker
	pb.BrokerStreamServer

	routerUpRate    *ratelimit.Registry
	handlerDownRate *ratelimit.Registry
}

func (b *brokerRPC) associateRouter(md metadata.MD) (chan *pb.UplinkMessage, <-chan *pb.DownlinkMessage, func(), error) {
	ctx := metadata.NewContext(context.Background(), md)
	router, err := b.broker.ValidateNetworkContext(ctx)
	if err != nil {
		return nil, nil, nil, err
	}
	down, err := b.broker.ActivateRouter(router.Id)
	if err != nil {
		return nil, nil, nil, err
	}

	up := make(chan *pb.UplinkMessage, 1)

	cancel := func() {
		b.broker.DeactivateRouter(router.Id)
	}

	go func() {
		for message := range up {
			if waitTime := b.routerUpRate.Wait(router.Id); waitTime != 0 {
				b.broker.Ctx.WithField("RouterID", router.Id).WithField("Wait", waitTime).Warn("Router reached uplink rate limit")
				time.Sleep(waitTime)
			}
			go b.broker.HandleUplink(message)
		}
	}()

	return up, down, cancel, nil
}

func (b *brokerRPC) getHandlerSubscribe(md metadata.MD) (<-chan *pb.DeduplicatedUplinkMessage, func(), error) {
	ctx := metadata.NewContext(context.Background(), md)
	handler, err := b.broker.ValidateNetworkContext(ctx)
	if err != nil {
		return nil, nil, err
	}

	ch, err := b.broker.ActivateHandlerUplink(handler.Id)
	if err != nil {
		return nil, nil, err
	}

	cancel := func() {
		b.broker.DeactivateHandlerUplink(handler.Id)
	}

	return ch, cancel, nil
}

func (b *brokerRPC) getHandlerPublish(md metadata.MD) (chan *pb.DownlinkMessage, error) {
	ctx := metadata.NewContext(context.Background(), md)
	handler, err := b.broker.ValidateNetworkContext(ctx)
	if err != nil {
		return nil, err
	}

	ch := make(chan *pb.DownlinkMessage, 1)
	go func() {
		for message := range ch {
			go func(downlink *pb.DownlinkMessage) {
				// Get latest Handler metadata
				handler, err := b.broker.Component.Discover("handler", handler.Id)
				if err != nil {
					return
				}
				for _, announcedID := range handler.AppIDs() {
					if announcedID == downlink.AppId {
						if waitTime := b.handlerDownRate.Wait(handler.Id); waitTime != 0 {
							b.broker.Ctx.WithField("HandlerID", handler.Id).WithField("Wait", waitTime).Warn("Handler reached downlink rate limit")
							time.Sleep(waitTime)
						}
						b.broker.HandleDownlink(downlink)
						return
					}
				}
			}(message)
		}
	}()
	return ch, nil
}

func (b *brokerRPC) Activate(ctx context.Context, req *pb.DeviceActivationRequest) (res *pb.DeviceActivationResponse, err error) {
	_, err = b.broker.ValidateNetworkContext(ctx)
	if err != nil {
		return nil, err
	}
	if err := req.Validate(); err != nil {
		return nil, errors.Wrap(err, "Invalid Activation Request")
	}
	res, err = b.broker.HandleActivation(req)
	if err == errDuplicateActivation {
		return nil, grpc.Errorf(codes.OutOfRange, err.Error())
	}
	if err != nil {
		return nil, err
	}
	return
}

func (b *broker) RegisterRPC(s *grpc.Server) {
	server := &brokerRPC{broker: b}
	server.SetLogger(b.Ctx)
	server.RouterAssociateChanFunc = server.associateRouter
	server.HandlerPublishChanFunc = server.getHandlerPublish
	server.HandlerSubscribeChanFunc = server.getHandlerSubscribe

	// TODO: Monitor actual rates and configure sensible limits
	server.routerUpRate = ratelimit.NewRegistry(1000, time.Second)
	server.handlerDownRate = ratelimit.NewRegistry(125, time.Second) // one eight of uplink

	pb.RegisterBrokerServer(s, server)
}

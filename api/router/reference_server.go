// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package router

import (
	"io"
	"sync"
	"sync/atomic"

	"github.com/TheThingsNetwork/go-utils/grpc/ttnctx"
	"github.com/TheThingsNetwork/go-utils/log"
	"github.com/TheThingsNetwork/ttn/api/gateway"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	"github.com/golang/protobuf/ptypes/empty"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

// NewReferenceRouterServer creates a new reference router server
func NewReferenceRouterServer(bufferSize int) *ReferenceRouterServer {
	s := &ReferenceRouterServer{
		ctx: log.Get(),

		bufferSize: bufferSize,

		gatewayStatuses: make(chan *gateway.Status, bufferSize),
		uplinkMessages:  make(chan *UplinkMessage, bufferSize),
		downlink:        make(map[string]*downlinkSubscription),

		metrics: new(metrics),
	}
	for i := 0; i < bufferSize; i++ {
		go func() {
			for {
				select {
				case <-s.gatewayStatuses:
					atomic.AddUint64(&s.metrics.gatewayStatuses, 1)
				case <-s.uplinkMessages:
					atomic.AddUint64(&s.metrics.uplinkMessages, 1)
				}
			}
		}()
	}
	return s
}

type metrics struct {
	gatewayStatuses uint64
	uplinkMessages  uint64
}

// ReferenceRouterServer is a new reference router server
type ReferenceRouterServer struct {
	ctx log.Interface

	bufferSize int

	gatewayStatuses chan *gateway.Status
	uplinkMessages  chan *UplinkMessage

	mu       sync.RWMutex
	downlink map[string]*downlinkSubscription

	metrics *metrics
}

type downlinkSubscription struct {
	ch          chan *DownlinkMessage
	subscribers int
}

func (s *ReferenceRouterServer) addDownlinkSubscriber(gatewayID string) chan *DownlinkMessage {
	s.mu.Lock()
	defer s.mu.Unlock()
	if sub, ok := s.downlink[gatewayID]; ok {
		sub.subscribers++
		return sub.ch
	}
	sub := &downlinkSubscription{
		subscribers: 1,
		ch:          make(chan *DownlinkMessage, s.bufferSize),
	}
	s.downlink[gatewayID] = sub
	return sub.ch
}

func (s *ReferenceRouterServer) removeDownlinkSubscriber(gatewayID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if sub, ok := s.downlink[gatewayID]; ok && sub.subscribers > 0 {
		sub.subscribers--
	}
}

func (s *ReferenceRouterServer) getAndAuthGateway(ctx context.Context) (string, error) {
	id, err := ttnctx.IDFromIncomingContext(ctx)
	if err != nil {
		return "", err
	}
	token, err := ttnctx.TokenFromIncomingContext(ctx)
	if err != nil {
		return "", err
	}
	// Actually validate token here, if failed: return nil, grpc.Errorf(codes.Unauthenticated, "Gateway Authentication Failed")
	s.ctx.WithFields(log.Fields{"ID": id, "Token": token}).Info("Gateway Authenticated")
	return id, nil
}

// GatewayStatus RPC
func (s *ReferenceRouterServer) GatewayStatus(stream Router_GatewayStatusServer) (err error) {
	gatewayID, err := s.getAndAuthGateway(stream.Context())
	if err != nil {
		return errors.NewErrPermissionDenied(err.Error())
	}
	ctx := s.ctx.WithField("GatewayID", gatewayID)
	ctx.Info("GatewayStatus stream started")
	defer func() {
		if err != nil {
			ctx.WithError(err).Info("GatewayStatus stream ended")
		} else {
			ctx.Info("GatewayStatus stream ended")
		}
	}()
	var streamErr atomic.Value
	go func() {
		<-stream.Context().Done()
		streamErr.Store(stream.Context().Err())
	}()
	for {
		streamErr := streamErr.Load()
		if streamErr != nil {
			return streamErr.(error)
		}
		msg, err := stream.Recv()
		if err == io.EOF {
			return stream.SendAndClose(&empty.Empty{})
		}
		if err != nil {
			return err
		}
		ctx.Info("Received GatewayStatus")
		select {
		case s.gatewayStatuses <- msg:
		default:
			ctx.Warn("Dropping Status")
		}
	}
}

// Uplink RPC
func (s *ReferenceRouterServer) Uplink(stream Router_UplinkServer) error {
	gatewayID, err := s.getAndAuthGateway(stream.Context())
	if err != nil {
		return errors.NewErrPermissionDenied(err.Error())
	}
	ctx := s.ctx.WithField("GatewayID", gatewayID)
	ctx.Info("GatewayUplink stream started")
	defer func() {
		if err != nil {
			ctx.WithError(err).Info("GatewayUplink stream ended")
		} else {
			ctx.Info("GatewayUplink stream ended")
		}
	}()
	var streamErr atomic.Value
	go func() {
		<-stream.Context().Done()
		streamErr.Store(stream.Context().Err())
	}()
	for {
		streamErr := streamErr.Load()
		if streamErr != nil {
			return streamErr.(error)
		}
		msg, err := stream.Recv()
		if err == io.EOF {
			return stream.SendAndClose(&empty.Empty{})
		}
		if err != nil {
			return err
		}
		ctx.Info("Received UplinkMessage")
		select {
		case s.uplinkMessages <- msg:
		default:
			ctx.Warn("Dropping UplinkMessage")
		}
	}
}

// Subscribe RPC
func (s *ReferenceRouterServer) Subscribe(req *SubscribeRequest, stream Router_SubscribeServer) error {
	gatewayID, err := s.getAndAuthGateway(stream.Context())
	if err != nil {
		return errors.NewErrPermissionDenied(err.Error())
	}
	ctx := s.ctx.WithField("GatewayID", gatewayID)
	ctx.Info("GatewayDownlink stream started")
	defer func() {
		if err != nil {
			ctx.WithError(err).Info("GatewayDownlink stream ended")
		} else {
			ctx.Info("GatewayDownlink stream ended")
		}
	}()

	sub := s.addDownlinkSubscriber(gatewayID)
	defer s.removeDownlinkSubscriber(gatewayID)

	for {
		select {
		case <-stream.Context().Done():
			return stream.Context().Err()
		case msg, ok := <-sub:
			if !ok {
				return nil
			}
			err := stream.Send(msg)
			if err != nil {
				return err
			}
			ctx.Info("Sent DownlinkMessage")
		}
	}
}

// Activate RPC
func (s *ReferenceRouterServer) Activate(ctx context.Context, req *DeviceActivationRequest) (*DeviceActivationResponse, error) {
	return nil, grpc.Errorf(codes.Unimplemented, "Not implemented")
}

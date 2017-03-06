// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package monitor

import (
	"context"
	"io"
	"sync/atomic"

	"github.com/TheThingsNetwork/go-utils/log"
	"github.com/TheThingsNetwork/ttn/api"
	"github.com/TheThingsNetwork/ttn/api/broker"
	"github.com/TheThingsNetwork/ttn/api/gateway"
	"github.com/TheThingsNetwork/ttn/api/router"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	"github.com/golang/protobuf/ptypes/empty"
)

// NewReferenceMonitorServer creates a new reference monitor server
func NewReferenceMonitorServer(bufferSize int) *ReferenceMonitorServer {
	s := &ReferenceMonitorServer{
		ctx: log.Get(),

		gatewayStatuses:  make(chan *gateway.Status, bufferSize),
		uplinkMessages:   make(chan *router.UplinkMessage, bufferSize),
		downlinkMessages: make(chan *router.DownlinkMessage, bufferSize),

		brokerUplinkMessages:   make(chan *broker.DeduplicatedUplinkMessage, bufferSize),
		brokerDownlinkMessages: make(chan *broker.DownlinkMessage, bufferSize),
	}
	for i := 0; i < bufferSize; i++ {
		go func() {
			for {
				select {
				case <-s.gatewayStatuses:
					atomic.AddUint64(&s.metrics.gatewayStatuses, 1)
				case <-s.uplinkMessages:
					atomic.AddUint64(&s.metrics.uplinkMessages, 1)
				case <-s.downlinkMessages:
					atomic.AddUint64(&s.metrics.downlinkMessages, 1)
				case <-s.brokerUplinkMessages:
					atomic.AddUint64(&s.metrics.brokerUplinkMessages, 1)
				case <-s.brokerDownlinkMessages:
					atomic.AddUint64(&s.metrics.brokerDownlinkMessages, 1)
				}
			}
		}()
	}
	return s
}

type metrics struct {
	gatewayStatuses        uint64
	uplinkMessages         uint64
	downlinkMessages       uint64
	brokerUplinkMessages   uint64
	brokerDownlinkMessages uint64
}

// ReferenceMonitorServer is a new reference monitor server
type ReferenceMonitorServer struct {
	ctx log.Interface

	gatewayStatuses  chan *gateway.Status
	uplinkMessages   chan *router.UplinkMessage
	downlinkMessages chan *router.DownlinkMessage

	brokerUplinkMessages   chan *broker.DeduplicatedUplinkMessage
	brokerDownlinkMessages chan *broker.DownlinkMessage

	metrics metrics
}

func (s *ReferenceMonitorServer) getAndAuthGateway(ctx context.Context) (string, error) {
	id, err := api.IDFromContext(ctx)
	if err != nil {
		return "", err
	}
	token, err := api.TokenFromContext(ctx)
	if err != nil {
		return "", err
	}
	// Actually validate token here, if failed: return nil, grpc.Errorf(codes.Unauthenticated, "Gateway Authentication Failed")
	s.ctx.WithFields(log.Fields{"ID": id, "Token": token}).Info("Gateway Authenticated")
	return id, nil
}

// GatewayStatus RPC
func (s *ReferenceMonitorServer) GatewayStatus(stream Monitor_GatewayStatusServer) (err error) {
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

// GatewayUplink RPC
func (s *ReferenceMonitorServer) GatewayUplink(stream Monitor_GatewayUplinkServer) error {
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

// GatewayDownlink RPC
func (s *ReferenceMonitorServer) GatewayDownlink(stream Monitor_GatewayDownlinkServer) error {
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
		ctx.Info("Received DownlinkMessage")
		select {
		case s.downlinkMessages <- msg:
		default:
			ctx.Warn("Dropping DownlinkMessage")
		}
	}
}

func (s *ReferenceMonitorServer) getAndAuthBroker(ctx context.Context) (string, error) {
	id, err := api.IDFromContext(ctx)
	if err != nil {
		return "", err
	}
	token, err := api.TokenFromContext(ctx)
	if err != nil {
		return "", err
	}
	// Actually validate token here, if failed: return nil, grpc.Errorf(codes.Unauthenticated, "Broker Authentication Failed")
	s.ctx.WithFields(log.Fields{"ID": id, "Token": token}).Info("Broker Authenticated")
	return id, nil
}

// BrokerUplink RPC
func (s *ReferenceMonitorServer) BrokerUplink(stream Monitor_BrokerUplinkServer) error {
	brokerID, err := s.getAndAuthBroker(stream.Context())
	if err != nil {
		return errors.NewErrPermissionDenied(err.Error())
	}
	ctx := s.ctx.WithField("BrokerID", brokerID)
	ctx.Info("BrokerUplink stream started")
	defer func() {
		if err != nil {
			ctx.WithError(err).Info("BrokerUplink stream ended")
		} else {
			ctx.Info("BrokerUplink stream ended")
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
		ctx.Info("Received DeduplicatedUplinkMessage")
		select {
		case s.brokerUplinkMessages <- msg:
		default:
			ctx.Warn("Dropping DeduplicatedUplinkMessage")
		}
	}
}

// BrokerDownlink RPC
func (s *ReferenceMonitorServer) BrokerDownlink(stream Monitor_BrokerDownlinkServer) error {
	brokerID, err := s.getAndAuthBroker(stream.Context())
	if err != nil {
		return errors.NewErrPermissionDenied(err.Error())
	}
	ctx := s.ctx.WithField("BrokerID", brokerID)
	ctx.Info("BrokerUplink stream started")
	defer func() {
		if err != nil {
			ctx.WithError(err).Info("BrokerUplink stream ended")
		} else {
			ctx.Info("BrokerUplink stream ended")
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
		ctx.Info("Received DownlinkMessage")
		select {
		case s.brokerDownlinkMessages <- msg:
		default:
			ctx.Warn("Dropping DownlinkMessage")
		}
	}
}

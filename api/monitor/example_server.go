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

func NewExampleMonitorServer(bufferSize int) *ExampleMonitorServer {
	s := &ExampleMonitorServer{
		ctx: log.Get(),

		gatewayStatuses:  make(chan *gateway.Status, bufferSize),
		uplinkMessages:   make(chan *router.UplinkMessage, bufferSize),
		downlinkMessages: make(chan *router.DownlinkMessage, bufferSize),

		brokerUplinkMessages:   make(chan *broker.DeduplicatedUplinkMessage, bufferSize),
		brokerDownlinkMessages: make(chan *broker.DownlinkMessage, bufferSize),
	}
	go func() {
		for {
			select {
			case <-s.gatewayStatuses:
				s.metrics.gatewayStatuses++
			case <-s.uplinkMessages:
				s.metrics.uplinkMessages++
			case <-s.downlinkMessages:
				s.metrics.downlinkMessages++
			case <-s.brokerUplinkMessages:
				s.metrics.brokerUplinkMessages++
			case <-s.brokerDownlinkMessages:
				s.metrics.brokerDownlinkMessages++
			}
		}
	}()
	return s
}

type metrics struct {
	gatewayStatuses        int
	uplinkMessages         int
	downlinkMessages       int
	brokerUplinkMessages   int
	brokerDownlinkMessages int
}

type ExampleMonitorServer struct {
	ctx log.Interface

	gatewayStatuses  chan *gateway.Status
	uplinkMessages   chan *router.UplinkMessage
	downlinkMessages chan *router.DownlinkMessage

	brokerUplinkMessages   chan *broker.DeduplicatedUplinkMessage
	brokerDownlinkMessages chan *broker.DownlinkMessage

	metrics metrics
}

func (s *ExampleMonitorServer) getAndAuthGateway(ctx context.Context) (string, error) {
	id, err := api.IDFromContext(ctx)
	if err != nil {
		return "", err
	}
	token, err := api.TokenFromContext(ctx)
	if err != nil {
		return "", err
	}
	// TODO: Validate token
	s.ctx.WithFields(log.Fields{"ID": id, "Token": token}).Info("Gateway Authenticated")
	return id, nil
}

func (s *ExampleMonitorServer) GatewayStatus(stream Monitor_GatewayStatusServer) (err error) {
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

func (s *ExampleMonitorServer) GatewayUplink(stream Monitor_GatewayUplinkServer) error {
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

func (s *ExampleMonitorServer) GatewayDownlink(stream Monitor_GatewayDownlinkServer) error {
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

func (s *ExampleMonitorServer) getAndAuthBroker(ctx context.Context) (string, error) {
	id, err := api.IDFromContext(ctx)
	if err != nil {
		return "", err
	}
	token, err := api.TokenFromContext(ctx)
	if err != nil {
		return "", err
	}
	// TODO: Validate token
	s.ctx.WithFields(log.Fields{"ID": id, "Token": token}).Info("Broker Authenticated")
	return id, nil
}

func (s *ExampleMonitorServer) BrokerUplink(stream Monitor_BrokerUplinkServer) error {
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

func (s *ExampleMonitorServer) BrokerDownlink(stream Monitor_BrokerDownlinkServer) error {
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

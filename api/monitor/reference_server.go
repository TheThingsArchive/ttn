// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package monitor

import (
	"context"
	"io"
	"sync/atomic"

	"github.com/TheThingsNetwork/go-utils/grpc/ttnctx"
	"github.com/TheThingsNetwork/go-utils/log"
	"github.com/TheThingsNetwork/ttn/api/broker"
	"github.com/TheThingsNetwork/ttn/api/fields"
	"github.com/TheThingsNetwork/ttn/api/gateway"
	"github.com/TheThingsNetwork/ttn/api/handler"
	"github.com/TheThingsNetwork/ttn/api/networkserver"
	"github.com/TheThingsNetwork/ttn/api/router"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	"github.com/golang/protobuf/ptypes/empty"
)

// NewReferenceMonitorServer creates a new reference monitor server
func NewReferenceMonitorServer(bufferSize int) *ReferenceMonitorServer {
	fields.Debug = true
	s := &ReferenceMonitorServer{
		ctx: log.Get(),

		gatewayStatuses:  make(chan *gateway.Status, bufferSize),
		uplinkMessages:   make(chan *router.UplinkMessage, bufferSize),
		downlinkMessages: make(chan *router.DownlinkMessage, bufferSize),

		brokerStatuses:         make(chan *broker.Status, bufferSize),
		brokerUplinkMessages:   make(chan *broker.DeduplicatedUplinkMessage, bufferSize),
		brokerDownlinkMessages: make(chan *broker.DownlinkMessage, bufferSize),

		handlerStatuses:         make(chan *handler.Status, bufferSize),
		handlerUplinkMessages:   make(chan *broker.DeduplicatedUplinkMessage, bufferSize),
		handlerDownlinkMessages: make(chan *broker.DownlinkMessage, bufferSize),

		routerStatuses: make(chan *router.Status, bufferSize),

		networkServerStatuses: make(chan *networkserver.Status, bufferSize),

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
				case <-s.downlinkMessages:
					atomic.AddUint64(&s.metrics.downlinkMessages, 1)
				case <-s.routerStatuses:
					atomic.AddUint64(&s.metrics.routerStatuses, 1)
				case <-s.brokerUplinkMessages:
					atomic.AddUint64(&s.metrics.brokerUplinkMessages, 1)
				case <-s.brokerDownlinkMessages:
					atomic.AddUint64(&s.metrics.brokerDownlinkMessages, 1)
				case <-s.brokerStatuses:
					atomic.AddUint64(&s.metrics.brokerStatuses, 1)
				case <-s.handlerUplinkMessages:
					atomic.AddUint64(&s.metrics.handlerUplinkMessages, 1)
				case <-s.handlerDownlinkMessages:
					atomic.AddUint64(&s.metrics.handlerDownlinkMessages, 1)
				case <-s.handlerStatuses:
					atomic.AddUint64(&s.metrics.handlerStatuses, 1)
				}
			}
		}()
	}
	return s
}

type metrics struct {
	gatewayStatuses         uint64
	uplinkMessages          uint64
	downlinkMessages        uint64
	brokerStatuses          uint64
	brokerUplinkMessages    uint64
	brokerDownlinkMessages  uint64
	handlerStatuses         uint64
	handlerUplinkMessages   uint64
	handlerDownlinkMessages uint64
	routerStatuses          uint64
	networkServerStatuses   uint64
}

// ReferenceMonitorServer is a new reference monitor server
type ReferenceMonitorServer struct {
	ctx log.Interface

	gatewayStatuses  chan *gateway.Status
	uplinkMessages   chan *router.UplinkMessage
	downlinkMessages chan *router.DownlinkMessage

	brokerStatuses         chan *broker.Status
	brokerUplinkMessages   chan *broker.DeduplicatedUplinkMessage
	brokerDownlinkMessages chan *broker.DownlinkMessage

	handlerStatuses         chan *handler.Status
	handlerUplinkMessages   chan *broker.DeduplicatedUplinkMessage
	handlerDownlinkMessages chan *broker.DownlinkMessage

	routerStatuses chan *router.Status

	networkServerStatuses chan *networkserver.Status

	metrics *metrics
}

func (s *ReferenceMonitorServer) getAndAuthGateway(ctx context.Context) (string, error) {
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
		ctx.WithFields(fields.Get(msg)).Info("Received GatewayStatus")
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
		msg.UnmarshalPayload()
		if msg.GetMessage().GetLorawan().GetJoinRequestPayload() != nil {
			ctx.WithFields(fields.Get(msg)).Info("Received ActivationRequest")
		} else {
			ctx.WithFields(fields.Get(msg)).Info("Received UplinkMessage")
		}
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
		msg.UnmarshalPayload()
		if msg.GetMessage().GetLorawan().GetJoinAcceptPayload() != nil {
			ctx.WithFields(fields.Get(msg)).Info("Received ActivationResponse")
		} else {
			ctx.WithFields(fields.Get(msg)).Info("Received DownlinkMessage")
		}
		select {
		case s.downlinkMessages <- msg:
		default:
			ctx.Warn("Dropping DownlinkMessage")
		}
	}
}

func (s *ReferenceMonitorServer) getAndAuthBroker(ctx context.Context) (string, error) {
	id, err := ttnctx.IDFromIncomingContext(ctx)
	if err != nil {
		return "", err
	}
	token, err := ttnctx.TokenFromIncomingContext(ctx)
	if err != nil {
		return "", err
	}
	// Actually validate token here, if failed: return nil, grpc.Errorf(codes.Unauthenticated, "Broker Authentication Failed")
	s.ctx.WithFields(log.Fields{"ID": id, "Token": token}).Info("Broker Authenticated")
	return id, nil
}

// BrokerStatus RPC
func (s *ReferenceMonitorServer) BrokerStatus(stream Monitor_BrokerStatusServer) (err error) {
	brokerID, err := s.getAndAuthBroker(stream.Context())
	if err != nil {
		return errors.NewErrPermissionDenied(err.Error())
	}
	ctx := s.ctx.WithField("BrokerID", brokerID)
	ctx.Info("BrokerStatus stream started")
	defer func() {
		if err != nil {
			ctx.WithError(err).Info("BrokerStatus stream ended")
		} else {
			ctx.Info("BrokerStatus stream ended")
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
		ctx.WithFields(fields.Get(msg)).Info("Received BrokerStatus")
		select {
		case s.brokerStatuses <- msg:
		default:
			ctx.Warn("Dropping Status")
		}
	}
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
		msg.UnmarshalPayload()
		if msg.GetMessage().GetLorawan().GetJoinRequestPayload() != nil {
			ctx.WithFields(fields.Get(msg)).Info("Received DeduplicatedActivationRequest")
		} else {
			ctx.WithFields(fields.Get(msg)).Info("Received DeduplicatedUplinkMessage")
		}
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
		msg.UnmarshalPayload()
		if msg.GetMessage().GetLorawan().GetJoinAcceptPayload() != nil {
			ctx.WithFields(fields.Get(msg)).Info("Received ActivationResponse")
		} else {
			ctx.WithFields(fields.Get(msg)).Info("Received DownlinkMessage")
		}
		select {
		case s.brokerDownlinkMessages <- msg:
		default:
			ctx.Warn("Dropping DownlinkMessage")
		}
	}
}

func (s *ReferenceMonitorServer) getAndAuthHandler(ctx context.Context) (string, error) {
	id, err := ttnctx.IDFromIncomingContext(ctx)
	if err != nil {
		return "", err
	}
	token, err := ttnctx.TokenFromIncomingContext(ctx)
	if err != nil {
		return "", err
	}
	// Actually validate token here, if failed: return nil, grpc.Errorf(codes.Unauthenticated, "Handler Authentication Failed")
	s.ctx.WithFields(log.Fields{"ID": id, "Token": token}).Info("Handler Authenticated")
	return id, nil
}

// HandlerStatus RPC
func (s *ReferenceMonitorServer) HandlerStatus(stream Monitor_HandlerStatusServer) (err error) {
	handlerID, err := s.getAndAuthHandler(stream.Context())
	if err != nil {
		return errors.NewErrPermissionDenied(err.Error())
	}
	ctx := s.ctx.WithField("HandlerID", handlerID)
	ctx.Info("HandlerStatus stream started")
	defer func() {
		if err != nil {
			ctx.WithError(err).Info("HandlerStatus stream ended")
		} else {
			ctx.Info("HandlerStatus stream ended")
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
		ctx.WithFields(fields.Get(msg)).Info("Received HandlerStatus")
		select {
		case s.handlerStatuses <- msg:
		default:
			ctx.Warn("Dropping Status")
		}
	}
}

// HandlerUplink RPC
func (s *ReferenceMonitorServer) HandlerUplink(stream Monitor_HandlerUplinkServer) error {
	handlerID, err := s.getAndAuthHandler(stream.Context())
	if err != nil {
		return errors.NewErrPermissionDenied(err.Error())
	}
	ctx := s.ctx.WithField("HandlerID", handlerID)
	ctx.Info("HandlerUplink stream started")
	defer func() {
		if err != nil {
			ctx.WithError(err).Info("HandlerUplink stream ended")
		} else {
			ctx.Info("HandlerUplink stream ended")
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
		msg.UnmarshalPayload()
		if msg.GetMessage().GetLorawan().GetJoinRequestPayload() != nil {
			ctx.WithFields(fields.Get(msg)).Info("Received DeduplicatedActivationRequest")
		} else {
			ctx.WithFields(fields.Get(msg)).Info("Received DeduplicatedUplinkMessage")
		}
		select {
		case s.handlerUplinkMessages <- msg:
		default:
			ctx.Warn("Dropping DeduplicatedUplinkMessage")
		}
	}
}

// HandlerDownlink RPC
func (s *ReferenceMonitorServer) HandlerDownlink(stream Monitor_HandlerDownlinkServer) error {
	handlerID, err := s.getAndAuthHandler(stream.Context())
	if err != nil {
		return errors.NewErrPermissionDenied(err.Error())
	}
	ctx := s.ctx.WithField("HandlerID", handlerID)
	ctx.Info("HandlerUplink stream started")
	defer func() {
		if err != nil {
			ctx.WithError(err).Info("HandlerUplink stream ended")
		} else {
			ctx.Info("HandlerUplink stream ended")
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
		msg.UnmarshalPayload()
		if msg.GetMessage().GetLorawan().GetJoinAcceptPayload() != nil {
			ctx.WithFields(fields.Get(msg)).Info("Received ActivationResponse")
		} else {
			ctx.WithFields(fields.Get(msg)).Info("Received DownlinkMessage")
		}
		select {
		case s.handlerDownlinkMessages <- msg:
		default:
			ctx.Warn("Dropping DownlinkMessage")
		}
	}
}

func (s *ReferenceMonitorServer) getAndAuthRouter(ctx context.Context) (string, error) {
	id, err := ttnctx.IDFromIncomingContext(ctx)
	if err != nil {
		return "", err
	}
	token, err := ttnctx.TokenFromIncomingContext(ctx)
	if err != nil {
		return "", err
	}
	// Actually validate token here, if failed: return nil, grpc.Errorf(codes.Unauthenticated, "Router Authentication Failed")
	s.ctx.WithFields(log.Fields{"ID": id, "Token": token}).Info("Router Authenticated")
	return id, nil
}

// RouterStatus RPC
func (s *ReferenceMonitorServer) RouterStatus(stream Monitor_RouterStatusServer) (err error) {
	routerID, err := s.getAndAuthRouter(stream.Context())
	if err != nil {
		return errors.NewErrPermissionDenied(err.Error())
	}
	ctx := s.ctx.WithField("RouterID", routerID)
	ctx.Info("RouterStatus stream started")
	defer func() {
		if err != nil {
			ctx.WithError(err).Info("RouterStatus stream ended")
		} else {
			ctx.Info("RouterStatus stream ended")
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
		ctx.WithFields(fields.Get(msg)).Info("Received RouterStatus")
		select {
		case s.routerStatuses <- msg:
		default:
			ctx.Warn("Dropping Status")
		}
	}
}

func (s *ReferenceMonitorServer) getAndAuthNetworkServer(ctx context.Context) (string, error) {
	id, err := ttnctx.IDFromIncomingContext(ctx)
	if err != nil {
		return "", err
	}
	token, err := ttnctx.TokenFromIncomingContext(ctx)
	if err != nil {
		return "", err
	}
	// Actually validate token here, if failed: return nil, grpc.Errorf(codes.Unauthenticated, "NetworkServer Authentication Failed")
	s.ctx.WithFields(log.Fields{"ID": id, "Token": token}).Info("NetworkServer Authenticated")
	return id, nil
}

// NetworkServerStatus RPC
func (s *ReferenceMonitorServer) NetworkServerStatus(stream Monitor_NetworkServerStatusServer) (err error) {
	networkServerID, err := s.getAndAuthNetworkServer(stream.Context())
	if err != nil {
		return errors.NewErrPermissionDenied(err.Error())
	}
	ctx := s.ctx.WithField("NetworkServerID", networkServerID)
	ctx.Info("NetworkServerStatus stream started")
	defer func() {
		if err != nil {
			ctx.WithError(err).Info("NetworkServerStatus stream ended")
		} else {
			ctx.Info("NetworkServerStatus stream ended")
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
		ctx.WithFields(fields.Get(msg)).Info("Received NetworkServerStatus")
		select {
		case s.networkServerStatuses <- msg:
		default:
			ctx.Warn("Dropping Status")
		}
	}
}

// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package router

import (
	"io"

	"github.com/TheThingsNetwork/go-utils/grpc/ttnctx"
	"github.com/TheThingsNetwork/go-utils/log"
	"github.com/TheThingsNetwork/ttn/api/gateway"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc/metadata"
)

// RouterStreamServer handles gRPC streams as channels
type RouterStreamServer struct {
	ctx                   log.Interface
	UplinkChanFunc        func(md metadata.MD) (ch chan *UplinkMessage, err error)
	GatewayStatusChanFunc func(md metadata.MD) (ch chan *gateway.Status, err error)
	DownlinkChanFunc      func(md metadata.MD) (ch <-chan *DownlinkMessage, cancel func(), err error)
}

// NewRouterStreamServer returns a new RouterStreamServer
func NewRouterStreamServer() *RouterStreamServer {
	return &RouterStreamServer{
		ctx: log.Get(),
	}
}

// SetLogger sets the logger
func (s *RouterStreamServer) SetLogger(logger log.Interface) {
	s.ctx = logger
}

// Uplink handles uplink streams
func (s *RouterStreamServer) Uplink(stream Router_UplinkServer) (err error) {
	md := ttnctx.MetadataFromIncomingContext(stream.Context())
	ch, err := s.UplinkChanFunc(md)
	if err != nil {
		return err
	}
	defer func() {
		ctx := s.ctx
		if err != nil {
			ctx = ctx.WithError(err)
		}
		close(ch)
		ctx.Debug("Closed Uplink stream")
	}()
	for {
		uplink, err := stream.Recv()
		if err == io.EOF {
			return stream.SendAndClose(&empty.Empty{})
		}
		if err != nil {
			return err
		}
		if err := uplink.Validate(); err != nil {
			s.ctx.WithError(err).Warn("Invalid Uplink")
			continue
		}
		if err := uplink.UnmarshalPayload(); err != nil {
			s.ctx.WithError(err).Warn("Could not unmarshal Uplink payload")
		}
		ch <- uplink
	}
}

// Subscribe handles downlink streams
func (s *RouterStreamServer) Subscribe(req *SubscribeRequest, stream Router_SubscribeServer) (err error) {
	md := ttnctx.MetadataFromIncomingContext(stream.Context())
	ch, cancel, err := s.DownlinkChanFunc(md)
	if err != nil {
		return err
	}
	go func() {
		<-stream.Context().Done()
		err = stream.Context().Err()
		cancel()
	}()
	for downlink := range ch {
		if err := stream.Send(downlink); err != nil {
			return err
		}
	}
	return
}

// GatewayStatus handles gateway status streams
func (s *RouterStreamServer) GatewayStatus(stream Router_GatewayStatusServer) error {
	md := ttnctx.MetadataFromIncomingContext(stream.Context())
	ch, err := s.GatewayStatusChanFunc(md)
	if err != nil {
		return err
	}
	defer func() {
		close(ch)
	}()
	for {
		status, err := stream.Recv()
		if err == io.EOF {
			return stream.SendAndClose(&empty.Empty{})
		}
		if err != nil {
			return err
		}
		if err := status.Validate(); err != nil {
			return errors.Wrap(err, "Invalid Gateway Status")
		}
		ch <- status
	}
}

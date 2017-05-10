// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package broker

import (
	"io"

	"github.com/TheThingsNetwork/go-utils/grpc/ttnctx"
	"github.com/TheThingsNetwork/go-utils/log"
	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc/metadata"
)

// BrokerStreamServer handles gRPC streams as channels
type BrokerStreamServer struct {
	ctx                      log.Interface
	RouterAssociateChanFunc  func(md metadata.MD) (up chan *UplinkMessage, down <-chan *DownlinkMessage, cancel func(), err error)
	HandlerSubscribeChanFunc func(md metadata.MD) (ch <-chan *DeduplicatedUplinkMessage, cancel func(), err error)
	HandlerPublishChanFunc   func(md metadata.MD) (ch chan *DownlinkMessage, err error)
}

// NewBrokerStreamServer returns a new BrokerStreamServer
func NewBrokerStreamServer() *BrokerStreamServer {
	return &BrokerStreamServer{
		ctx: log.Get(),
	}
}

// SetLogger sets the logger
func (s *BrokerStreamServer) SetLogger(logger log.Interface) {
	s.ctx = logger
}

// Associate handles uplink streams from and downlink streams to the router
func (s *BrokerStreamServer) Associate(stream Broker_AssociateServer) (err error) {
	md := ttnctx.MetadataFromIncomingContext(stream.Context())
	upChan, downChan, downCancel, err := s.RouterAssociateChanFunc(md)
	if err != nil {
		return err
	}
	defer func() {
		ctx := s.ctx
		if err != nil {
			ctx = ctx.WithError(err)
		}
		downCancel()
		close(upChan)
		ctx.Debug("Closed Associate stream")
	}()

	upErr := make(chan error)
	go func() (err error) {
		defer func() {
			if err != nil {
				upErr <- err
			}
			close(upErr)
		}()
		for {
			uplink, err := stream.Recv()
			if err == io.EOF {
				return nil
			}
			if err != nil {
				return err
			}
			if err := uplink.Validate(); err != nil {
				s.ctx.WithError(err).Warn("Invalid Uplink")
				continue
			}
			if err := uplink.UnmarshalPayload(); err != nil {
				s.ctx.WithError(err).Warn("Could not unmarshal uplink payload")
			}
			upChan <- uplink
		}
	}()

	for {
		select {
		case <-stream.Context().Done():
			return stream.Context().Err()
		case err, errPresent := <-upErr:
			if !errPresent {
				return nil // stream closed
			}
			return err
		case downlink, downlinkPresent := <-downChan:
			if !downlinkPresent {
				return nil // stream closed
			}
			if err := stream.Send(downlink); err != nil {
				return err
			}
		}
	}
}

// Subscribe handles uplink streams towards the handler
func (s *BrokerStreamServer) Subscribe(req *SubscribeRequest, stream Broker_SubscribeServer) (err error) {
	md := ttnctx.MetadataFromIncomingContext(stream.Context())
	ch, cancel, err := s.HandlerSubscribeChanFunc(md)
	if err != nil {
		return err
	}
	go func() {
		<-stream.Context().Done()
		err = stream.Context().Err()
		cancel()
	}()
	for uplink := range ch {
		if err := stream.Send(uplink); err != nil {
			return err
		}
	}
	return
}

// Publish handles downlink streams from the handler
func (s *BrokerStreamServer) Publish(stream Broker_PublishServer) error {
	md := ttnctx.MetadataFromIncomingContext(stream.Context())
	ch, err := s.HandlerPublishChanFunc(md)
	if err != nil {
		return err
	}
	defer func() {
		close(ch)
	}()
	for {
		downlink, err := stream.Recv()
		if err == io.EOF {
			return stream.SendAndClose(&empty.Empty{})
		}
		if err != nil {
			return err
		}
		if err := downlink.Validate(); err != nil {
			s.ctx.WithError(err).Warn("Invalid Downlink")
			continue
		}
		if err := downlink.UnmarshalPayload(); err != nil {
			s.ctx.WithError(err).Warn("Could not unmarshal downlink payload")
		}
		ch <- downlink
	}
}

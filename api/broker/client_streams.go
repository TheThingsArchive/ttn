// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package broker

import (
	"io"
	"sync"
	"time"

	"github.com/TheThingsNetwork/go-utils/log"
	"github.com/TheThingsNetwork/ttn/api/fields"
	"github.com/TheThingsNetwork/ttn/utils/backoff"
	"github.com/golang/protobuf/ptypes/empty"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

// Stream interface
type Stream interface {
	SetLogger(log.Interface)
	Close()
}

type stream struct {
	closing bool
	setup   sync.WaitGroup
	ctx     log.Interface
	client  BrokerClient
}

func (s *stream) SetLogger(logger log.Interface) {
	s.ctx = logger
}

// DefaultBufferSize indicates the default send and receive buffer sizes
var DefaultBufferSize = 10

// RouterStream for sending gateway statuses
type RouterStream interface {
	Stream
	Send(*UplinkMessage) error
	Channel() <-chan *DownlinkMessage
}

// NewMonitoredRouterStream starts and monitors a RouterStream
func NewMonitoredRouterStream(client BrokerClient, getContextFunc func() context.Context) RouterStream {
	s := &routerStream{
		up:   make(chan *UplinkMessage, DefaultBufferSize),
		down: make(chan *DownlinkMessage, DefaultBufferSize),
		err:  make(chan error),
	}
	s.setup.Add(1)
	s.client = client
	s.ctx = log.Get()

	go func() {
		var retries int

		for {
			// Session channels
			up := make(chan *UplinkMessage)
			errCh := make(chan error)

			// Session client
			var ctx context.Context
			ctx, s.cancel = context.WithCancel(getContextFunc())
			client, err := s.client.Associate(ctx)
			s.setup.Done()
			if err != nil {
				s.ctx.WithError(err).Warn("Could not start Associate stream, retrying...")
				s.setup.Add(1)
				time.Sleep(backoff.Backoff(retries))
				retries++
				continue
			}
			retries = 0

			s.ctx.Debug("Started Associate stream")

			// Receive downlink errors
			go func() {
				for {
					message, err := client.Recv()
					if message != nil {
						s.ctx.WithFields(fields.Get(message)).Debug("Receiving Downlink message")
						if err := message.Validate(); err != nil {
							s.ctx.WithError(err).Warn("Invalid Downlink")
							continue
						}
						if err := message.UnmarshalPayload(); err != nil {
							s.ctx.Warn("Could not unmarshal Downlink payload")
						}
						select {
						case s.down <- message:
						default:
							s.ctx.Warn("Dropping Downlink message, buffer full")
						}
					}
					if err != nil {
						errCh <- err
						break
					}
				}
				close(errCh)
			}()

			// Send uplink
			go func() {
				for message := range up {
					s.ctx.WithFields(fields.Get(message)).Debug("Sending Uplink message")
					if err := client.Send(message); err != nil {
						s.ctx.WithError(err).Warn("Error sending Uplink message")
						break
					}
				}
			}()

			// Monitoring
			var mErr error

		monitor:
			for {
				select {
				case mErr = <-errCh:
					break monitor
				case msg, ok := <-s.up:
					if !ok {
						break monitor // channel closed
					}
					up <- msg
				}
			}

			close(up)
			client.CloseSend()

			if mErr == nil || mErr == io.EOF || grpc.Code(mErr) == codes.Canceled {
				s.ctx.Debug("Stopped Associate stream")
			} else {
				s.ctx.WithError(mErr).Warn("Error in Associate stream")
			}

			if s.closing {
				break
			}

			s.setup.Add(1)
			time.Sleep(backoff.Backoff(retries))
			retries++
		}
	}()

	return s
}

type routerStream struct {
	stream
	cancel context.CancelFunc
	up     chan *UplinkMessage
	down   chan *DownlinkMessage
	err    chan error
}

func (s *routerStream) Send(uplink *UplinkMessage) error {
	select {
	case s.up <- uplink:
	default:
		s.ctx.Warn("Dropping Uplink message, buffer full")
	}
	return nil
}

func (s *routerStream) Channel() <-chan *DownlinkMessage {
	return s.down
}

func (s *routerStream) Close() {
	s.closing = true
	close(s.up)
	if s.cancel != nil {
		s.cancel()
	}
}

// HandlerPublishStream for sending downlink messages to the broker
type HandlerPublishStream interface {
	Stream
	Send(*DownlinkMessage) error
}

// NewMonitoredHandlerPublishStream starts and monitors a HandlerPublishStream
func NewMonitoredHandlerPublishStream(client BrokerClient, getContextFunc func() context.Context) HandlerPublishStream {
	s := &handlerPublishStream{
		ch:  make(chan *DownlinkMessage, DefaultBufferSize),
		err: make(chan error),
	}
	s.setup.Add(1)
	s.client = client
	s.ctx = log.Get()

	go func() {
		var retries int

		for {
			// Session channels
			ch := make(chan *DownlinkMessage)
			errCh := make(chan error)

			// Session client
			client, err := s.client.Publish(getContextFunc())
			s.setup.Done()
			if err != nil {
				if grpc.Code(err) == codes.Canceled {
					s.ctx.Debug("Stopped Downlink stream")
					break
				}
				s.ctx.WithError(err).Warn("Could not start Downlink stream, retrying...")
				s.setup.Add(1)
				time.Sleep(backoff.Backoff(retries))
				retries++
				continue
			}
			retries = 0

			s.ctx.Info("Started Downlink stream")

			// Receive errors
			go func() {
				empty := new(empty.Empty)
				if err := client.RecvMsg(empty); err != nil {
					errCh <- err
				}
				close(errCh)
			}()

			// Send
			go func() {
				for message := range ch {
					s.ctx.WithFields(fields.Get(message)).Debug("Sending Downlink message")
					if err := client.Send(message); err != nil {
						s.ctx.WithError(err).Warn("Error sending Downlink message")
						break
					}
				}
			}()

			// Monitoring
			var mErr error

		monitor:
			for {
				select {
				case mErr = <-errCh:
					break monitor
				case msg, ok := <-s.ch:
					if !ok {
						break monitor // channel closed
					}
					ch <- msg
				}
			}

			close(ch)
			client.CloseAndRecv()

			if mErr == nil || mErr == io.EOF || grpc.Code(mErr) == codes.Canceled {
				s.ctx.Debug("Stopped Downlink stream")
			} else {
				s.ctx.WithError(mErr).Warn("Error in Downlink stream")
			}

			if s.closing {
				break
			}

			s.setup.Add(1)
			time.Sleep(backoff.Backoff(retries))
			retries++
		}
	}()

	return s
}

type handlerPublishStream struct {
	stream
	ch  chan *DownlinkMessage
	err chan error
}

func (s *handlerPublishStream) Send(message *DownlinkMessage) error {
	select {
	case s.ch <- message:
	default:
		s.ctx.Warn("Dropping Downlink message, buffer full")
	}
	return nil
}

func (s *handlerPublishStream) Close() {
	s.setup.Wait()
	s.ctx.Debug("Closing Downlink stream")
	s.closing = true
	close(s.ch)
}

// HandlerSubscribeStream for receiving uplink messages
type HandlerSubscribeStream interface {
	Stream
	Channel() <-chan *DeduplicatedUplinkMessage
}

// NewMonitoredHandlerSubscribeStream starts and monitors a HandlerSubscribeStream
func NewMonitoredHandlerSubscribeStream(client BrokerClient, getContextFunc func() context.Context) HandlerSubscribeStream {
	s := &handlerSubscribeStream{
		ch:  make(chan *DeduplicatedUplinkMessage, DefaultBufferSize),
		err: make(chan error),
	}
	s.setup.Add(1)
	s.client = client
	s.ctx = log.Get()

	go func() {
		var client Broker_SubscribeClient
		var err error
		var retries int
		var message *DeduplicatedUplinkMessage

		for {
			// Session client
			var ctx context.Context
			ctx, s.cancel = context.WithCancel(getContextFunc())
			client, err = s.client.Subscribe(ctx, &SubscribeRequest{})
			s.setup.Done()
			if err != nil {
				if grpc.Code(err) == codes.Canceled {
					s.ctx.Debug("Stopped Uplink stream")
					break
				}
				s.ctx.WithError(err).Warn("Could not start Uplink stream, retrying...")
				s.setup.Add(1)
				time.Sleep(backoff.Backoff(retries))
				retries++
				continue
			}
			retries = 0

			s.ctx.Info("Started Uplink stream")

			for {
				message, err = client.Recv()
				if message != nil {
					s.ctx.WithFields(fields.Get(message)).Debug("Receiving Uplink message")
					if err := message.Validate(); err != nil {
						s.ctx.WithError(err).Warn("Invalid Uplink")
						continue
					}
					if err := message.UnmarshalPayload(); err != nil {
						s.ctx.Warn("Could not unmarshal Uplink payload")
					}
					select {
					case s.ch <- message:
					default:
						s.ctx.Warn("Dropping Uplink message, buffer full")
					}
				}
				if err != nil {
					break
				}
			}

			if err == nil || err == io.EOF || grpc.Code(err) == codes.Canceled {
				s.ctx.Debug("Stopped Uplink stream")
			} else {
				s.ctx.WithError(err).Warn("Error in Uplink stream")
			}

			if s.closing {
				break
			}

			s.setup.Add(1)
			time.Sleep(backoff.Backoff(retries))
			retries++
		}

		close(s.ch)
	}()
	return s
}

type handlerSubscribeStream struct {
	stream
	cancel context.CancelFunc
	ch     chan *DeduplicatedUplinkMessage
	err    chan error
}

func (s *handlerSubscribeStream) Close() {
	s.setup.Wait()
	s.ctx.Debug("Closing Uplink stream")
	s.closing = true
	if s.cancel != nil {
		s.cancel()
	}
}

func (s *handlerSubscribeStream) Channel() <-chan *DeduplicatedUplinkMessage {
	return s.ch
}

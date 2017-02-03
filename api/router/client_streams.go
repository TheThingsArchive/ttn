// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package router

import (
	"io"
	"sync"
	"time"

	"github.com/TheThingsNetwork/go-utils/log"
	"github.com/TheThingsNetwork/ttn/api/gateway"
	"github.com/TheThingsNetwork/ttn/utils/backoff"
	"github.com/golang/protobuf/ptypes/empty"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

// GatewayStream interface
type GatewayStream interface {
	Close()
}

type gatewayStream struct {
	closing bool
	setup   sync.WaitGroup
	ctx     log.Interface
	client  RouterClientForGateway
}

// DefaultBufferSize indicates the default send and receive buffer sizes
var DefaultBufferSize = 10

// GatewayStatusStream for sending gateway statuses
type GatewayStatusStream interface {
	GatewayStream
	Send(*gateway.Status) error
}

// NewMonitoredGatewayStatusStream starts and monitors a GatewayStatusStream
func NewMonitoredGatewayStatusStream(client RouterClientForGateway) GatewayStatusStream {
	s := &gatewayStatusStream{
		ch:  make(chan *gateway.Status, DefaultBufferSize),
		err: make(chan error),
	}
	s.setup.Add(1)
	s.client = client
	s.ctx = client.GetLogger()

	go func() {
		var retries int

		for {
			// Session channels
			ch := make(chan *gateway.Status)
			errCh := make(chan error)

			// Session client
			client, err := s.client.GatewayStatus()
			s.setup.Done()
			if err != nil {
				if grpc.Code(err) == codes.Canceled {
					s.ctx.Debug("Stopped GatewayStatus stream")
					break
				}
				s.ctx.WithError(err).Warn("Could not start GatewayStatus stream, retrying...")
				s.setup.Add(1)
				time.Sleep(backoff.Backoff(retries))
				retries++
				continue
			}

			s.ctx.Info("Started GatewayStatus stream")

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
				for status := range ch {
					s.ctx.Debug("Sending GatewayStatus message")
					if err := client.Send(status); err != nil {
						s.ctx.WithError(err).Warn("Error sending GatewayStatus message")
						break
					}
				}
			}()

			// Monitoring
			var mErr error

		monitor:
			for {
				select {
				case <-time.After(10 * time.Second):
					retries = 0
				case mErr = <-errCh:
					break monitor
				case msg, ok := <-s.ch:
					if !ok {
						break monitor // channel closed
					}
					ch <- msg
				case <-s.client.TokenChange():
					s.ctx.Debug("Restarting GatewayStatus stream with new token")
					break monitor
				}
			}

			close(ch)
			client.CloseAndRecv()

			if mErr == nil || mErr == io.EOF || grpc.Code(mErr) == codes.Canceled {
				s.ctx.Debug("Stopped GatewayStatus stream")
			} else {
				s.ctx.WithError(mErr).Warn("Error in GatewayStatus stream")
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

type gatewayStatusStream struct {
	gatewayStream
	ch  chan *gateway.Status
	err chan error
}

func (s *gatewayStatusStream) Send(status *gateway.Status) error {
	select {
	case s.ch <- status:
	default:
		s.ctx.Warn("Dropping GatewayStatus message, buffer full")
	}
	return nil
}

func (s *gatewayStatusStream) Close() {
	s.setup.Wait()
	s.ctx.Debug("Closing GatewayStatus stream")
	s.closing = true
	close(s.ch)
}

// UplinkStream for sending uplink messages
type UplinkStream interface {
	GatewayStream
	Send(*UplinkMessage) error
}

// NewMonitoredUplinkStream starts and monitors a UplinkStream
func NewMonitoredUplinkStream(client RouterClientForGateway) UplinkStream {
	s := &uplinkStream{
		ch:  make(chan *UplinkMessage, DefaultBufferSize),
		err: make(chan error),
	}
	s.setup.Add(1)
	s.client = client
	s.ctx = client.GetLogger()

	go func() {
		var retries int

		for {
			// Session channels
			ch := make(chan *UplinkMessage)
			errCh := make(chan error)

			// Session client
			client, err := s.client.Uplink()
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

			s.ctx.Info("Started Uplink stream")

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
					s.ctx.Debug("Sending Uplink message")
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
				case <-time.After(10 * time.Second):
					retries = 0
				case mErr = <-errCh:
					break monitor
				case msg, ok := <-s.ch:
					if !ok {
						break monitor // channel closed
					}
					ch <- msg
				case <-s.client.TokenChange():
					s.ctx.Debug("Restarting Uplink stream with new token")
					break monitor
				}
			}

			close(ch)
			client.CloseAndRecv()

			if mErr == nil || mErr == io.EOF || grpc.Code(mErr) == codes.Canceled {
				s.ctx.Debug("Stopped Uplink stream")
			} else {
				s.ctx.WithError(mErr).Warn("Error in Uplink stream")
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

type uplinkStream struct {
	gatewayStream
	ch  chan *UplinkMessage
	err chan error
}

func (s *uplinkStream) Send(message *UplinkMessage) error {
	select {
	case s.ch <- message:
	default:
		s.ctx.Warn("Dropping Uplink message, buffer full")
	}
	return nil
}

func (s *uplinkStream) Close() {
	s.setup.Wait()
	s.ctx.Debug("Closing Uplink stream")
	s.closing = true
	close(s.ch)
}

// DownlinkStream for receiving downlink messages
type DownlinkStream interface {
	GatewayStream
	Channel() <-chan *DownlinkMessage
}

// NewMonitoredDownlinkStream starts and monitors a DownlinkStream
func NewMonitoredDownlinkStream(client RouterClientForGateway) DownlinkStream {
	s := &downlinkStream{
		ch:  make(chan *DownlinkMessage, DefaultBufferSize),
		err: make(chan error),
	}
	s.setup.Add(1)
	s.client = client
	s.ctx = client.GetLogger()

	go func() {
		var client Router_SubscribeClient
		var err error
		var retries int
		var message *DownlinkMessage

		for {
			client, s.cancel, err = s.client.Subscribe()
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

			go func() {
				<-s.client.TokenChange()
				s.ctx.Debug("Restarting Downlink stream with new token")
				s.cancel()
			}()

			s.ctx.Info("Started Downlink stream")

			for {
				message, err = client.Recv()
				if message != nil {
					s.ctx.Debug("Receiving Downlink message")
					if err := message.Validate(); err != nil {
						s.ctx.WithError(err).Warn("Invalid Downlink")
						continue
					}
					if err := message.UnmarshalPayload(); err != nil {
						s.ctx.Warn("Could not unmarshal Downlink payload")
					}
					select {
					case s.ch <- message:
					default:
						s.ctx.Warn("Dropping Downlink message, buffer full")
					}
				}
				if err != nil {
					break
				}
				retries = 0
			}

			if err == nil || err == io.EOF || grpc.Code(err) == codes.Canceled {
				s.ctx.Debug("Stopped Downlink stream")
			} else {
				s.ctx.WithError(err).Warn("Error in Downlink stream")
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

type downlinkStream struct {
	gatewayStream
	cancel context.CancelFunc
	ch     chan *DownlinkMessage
	err    chan error
}

func (s *downlinkStream) Close() {
	s.setup.Wait()
	s.ctx.Debug("Closing Downlink stream")
	s.closing = true
	if s.cancel != nil {
		s.cancel()
	}
}

func (s *downlinkStream) Channel() <-chan *DownlinkMessage {
	return s.ch
}

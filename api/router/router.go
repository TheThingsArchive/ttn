// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package router

import (
	"context"
	"errors"
	"io"
	"sync"

	"github.com/TheThingsNetwork/go-utils/grpc/restartstream"
	"github.com/TheThingsNetwork/go-utils/log"
	"github.com/TheThingsNetwork/ttn/api"
	"github.com/TheThingsNetwork/ttn/api/gateway"
	"github.com/TheThingsNetwork/ttn/utils"
	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

// GenericStream is used for sending to and receiving from the router.
type GenericStream interface {
	Uplink(*UplinkMessage)
	Status(*gateway.Status)
	Downlink() (<-chan *DownlinkMessage, error)
	Close()
}

// ClientConfig for router Client
type ClientConfig struct {
	BackgroundContext context.Context
	BufferSize        int
}

// DefaultClientConfig for router Client
var DefaultClientConfig = ClientConfig{
	BackgroundContext: context.Background(),
	BufferSize:        10,
}

// NewClient creates a new Client with the given configuration
func NewClient(config ClientConfig) *Client {
	ctx, cancel := context.WithCancel(config.BackgroundContext)

	return &Client{
		log:    log.Get(),
		ctx:    ctx,
		cancel: cancel,

		config: config,
	}
}

// Client for router
type Client struct {
	log    log.Interface
	ctx    context.Context
	cancel context.CancelFunc

	config      ClientConfig
	serverConns []*serverConn
}

// AddServer adds a router server
func (c *Client) AddServer(name string, conn *grpc.ClientConn) {
	log := c.log.WithField("Router", name)
	log.Info("Adding Router server")
	s := &serverConn{
		ctx:  log,
		name: name,
		conn: conn,
	}
	c.serverConns = append(c.serverConns, s)
}

// Close the client and all its connections
func (c *Client) Close() {
	c.cancel()
	for _, server := range c.serverConns {
		server.Close()
	}
}

type serverConn struct {
	ctx  log.Interface
	name string

	ready chan struct{}
	conn  *grpc.ClientConn
}

func (c *serverConn) Close() {
	if c.ready != nil {
		<-c.ready
	}
	if c.conn != nil {
		c.conn.Close()
	}
}

type gatewayStreams struct {
	log    log.Interface
	ctx    context.Context
	cancel context.CancelFunc

	mu     sync.RWMutex
	uplink map[string]chan *UplinkMessage
	status map[string]chan *gateway.Status

	downlink chan *DownlinkMessage
}

func (s *gatewayStreams) Uplink(msg *UplinkMessage) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	s.log.WithField("Routers", len(s.uplink)).Debug("Sending UplinkMessage to router")
	for serverName, ch := range s.uplink {
		select {
		case ch <- msg:
		default:
			s.log.WithField("Router", serverName).Warn("UplinkMessage buffer full")
		}
	}
}

func (s *gatewayStreams) Status(msg *gateway.Status) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	s.log.WithField("Routers", len(s.status)).Debug("Sending Status to router")
	for serverName, ch := range s.status {
		select {
		case ch <- msg:
		default:
			s.log.WithField("Router", serverName).Warn("GatewayStatus buffer full")
		}
	}
}

// ErrDownlinkInactive is returned by the Downlink func if downlink is inactive
var ErrDownlinkInactive = errors.New("Downlink stream not active")

func (s *gatewayStreams) Downlink() (<-chan *DownlinkMessage, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.downlink == nil {
		return nil, ErrDownlinkInactive
	}
	return s.downlink, nil
}

func (s *gatewayStreams) Close() {
	s.cancel()
}

// NewGatewayStreams returns new streams using the given gateway ID and token
func (c *Client) NewGatewayStreams(id string, token string, downlinkActive bool) GenericStream {
	log := c.log.WithField("GatewayID", id)
	ctx, cancel := context.WithCancel(c.ctx)
	ctx = api.ContextWithID(ctx, id)
	ctx = api.ContextWithToken(ctx, token)
	s := &gatewayStreams{
		log:    log,
		ctx:    ctx,
		cancel: cancel,

		uplink: make(map[string]chan *UplinkMessage),
		status: make(map[string]chan *gateway.Status),
	}

	var wg utils.WaitGroup

	var wgDown sync.WaitGroup
	if downlinkActive {
		s.downlink = make(chan *DownlinkMessage, c.config.BufferSize)
		defer func() {
			go func() {
				wgDown.Wait()
				close(s.downlink)
			}()
		}()
	}

	// Hook up the router servers
	for _, server := range c.serverConns {
		wg.Add(1)
		wgDown.Add(1)

		// Stream channels
		chUplink := make(chan *UplinkMessage, c.config.BufferSize)
		chStatus := make(chan *gateway.Status, c.config.BufferSize)

		s.mu.Lock()
		s.uplink[server.name] = chUplink
		s.status[server.name] = chStatus
		s.mu.Unlock()

		go func(server *serverConn) {
			defer func() {
				s.mu.Lock()
				defer s.mu.Unlock()
				delete(s.uplink, server.name)
				delete(s.status, server.name)
				close(chUplink)
				close(chStatus)
			}()

			if server.ready != nil {
				select {
				case <-ctx.Done():
					return
				case <-server.ready:
				}
			}
			if server.conn == nil {
				return
			}
			log := log.WithField("Router", server.name)
			cli := NewRouterClient(server.conn)

			logStreamErr := func(streamName string, err error) {
				switch {
				case err == nil:
					log.Debugf("%s stream closed", streamName)
				case err == io.EOF:
					log.WithError(err).Debugf("%s stream ended", streamName)
				case err == context.Canceled || grpc.Code(err) == codes.Canceled:
					log.WithError(err).Debugf("%s stream canceled", streamName)
				case err == context.DeadlineExceeded || grpc.Code(err) == codes.DeadlineExceeded:
					log.WithError(err).Debugf("%s stream deadline exceeded", streamName)
				case grpc.ErrorDesc(err) == grpc.ErrClientConnClosing.Error():
					log.WithError(err).Debugf("%s stream connection closed", streamName)
				default:
					log.WithError(err).Warnf("%s stream closed unexpectedly", streamName)
				}
			}

			// Uplink stream
			uplink, err := cli.Uplink(ctx)
			if err != nil {
				log.WithError(err).Warn("Could not set up Uplink stream")
				s.mu.Lock()
				delete(s.uplink, server.name)
				s.mu.Unlock()
			} else {
				go func() {
					err := uplink.RecvMsg(new(empty.Empty))
					logStreamErr("Uplink", err)
					s.mu.Lock()
					defer s.mu.Unlock()
					delete(s.uplink, server.name)
				}()
			}

			// Downlink stream
			if downlinkActive {
				downlink, err := cli.Subscribe(ctx, &SubscribeRequest{})
				if err != nil {
					log.WithError(err).Warn("Could not set up Subscribe stream")
					wgDown.Done()
				} else {
					go func() {
						defer func() {
							wgDown.Done()
						}()
						for {
							msg, err := downlink.Recv()
							if err != nil {
								logStreamErr("Subscribe", err)
								return
							}
							select {
							case s.downlink <- msg:
							default:
								log.Warn("Downlink buffer full")
							}
						}
					}()
				}
			}

			// Status stream
			status, err := cli.GatewayStatus(ctx)
			if err != nil {
				log.WithError(err).Warn("Could not set up GatewayStatus stream")
				s.mu.Lock()
				delete(s.status, server.name)
				s.mu.Unlock()
			} else {
				go func() {
					err := status.RecvMsg(new(empty.Empty))
					logStreamErr("GatewayStatus", err)
					s.mu.Lock()
					defer s.mu.Unlock()
					delete(s.status, server.name)
				}()
			}

			wg.Done()
			log.Debug("Start handling Gateway streams")
			defer log.Debug("Done handling Gateway streams")
			for {
				select {
				case <-ctx.Done():
					return
				case msg := <-chStatus:
					if err := status.Send(msg); err != nil {
						log.WithError(err).Warn("Could not send GatewayStatus to router")
						if err == restartstream.ErrStreamClosed {
							return
						}
					}
				case msg := <-chUplink:
					if err := uplink.Send(msg); err != nil {
						log.WithError(err).Warn("Could not send UplinkMessage to router")
						if err == restartstream.ErrStreamClosed {
							return
						}
					}
				}
			}

		}(server)
	}

	if api.WaitForStreams > 0 {
		wg.WaitForMax(api.WaitForStreams)
	}

	return s
}

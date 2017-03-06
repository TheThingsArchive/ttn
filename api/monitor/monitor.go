// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package monitor

import (
	"io"

	"sync"

	"github.com/TheThingsNetwork/go-utils/grpc/restartstream"
	"github.com/TheThingsNetwork/go-utils/log"
	"github.com/TheThingsNetwork/ttn/api"
	"github.com/TheThingsNetwork/ttn/api/broker"
	"github.com/TheThingsNetwork/ttn/api/gateway"
	"github.com/TheThingsNetwork/ttn/api/router"
	"github.com/golang/protobuf/ptypes/empty"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

// GenericStream is used for sending anything to the monitor.
// Depending on the context, this can be
// - *router.UplinkMessage
// - *router.DownlinkMessage
// - *gateway.Status
// - *broker.DeduplicatedUplinkMessage
// - *broker.DownlinkMessage
type GenericStream interface {
	Send(interface{})
	Close()
}

// ClientConfig for monitor Client
type ClientConfig struct {
	BufferSize int
}

// DefaultClientConfig for monitor Client
var DefaultClientConfig = ClientConfig{
	BufferSize: 10,
}

// NewClient creates a new Client with the given configuration
func NewClient(config ClientConfig) *Client {
	ctx, cancel := context.WithCancel(context.Background())

	return &Client{
		log:    log.Get(),
		ctx:    ctx,
		cancel: cancel,

		config: config,
	}
}

// Client for monitor
type Client struct {
	log    log.Interface
	ctx    context.Context
	cancel context.CancelFunc

	config      ClientConfig
	serverConns []*serverConn
}

var defaultDialOptions = []grpc.DialOption{
	grpc.WithBlock(),
	grpc.FailOnNonTempDialError(false),
	grpc.WithStreamInterceptor(restartstream.Interceptor(restartstream.DefaultSettings)),
}

// AddServer adds a new monitor server
func (c *Client) AddServer(name, address string, opts ...grpc.DialOption) {
	log := c.log.WithFields(log.Fields{"Monitor": name, "Address": address})
	log.Info("Adding Monitor server")

	s := &serverConn{
		ctx:   log,
		name:  name,
		ready: make(chan struct{}),
	}
	c.serverConns = append(c.serverConns, s)

	go func() {
		conn, err := grpc.DialContext(
			c.ctx,
			address,
			append(defaultDialOptions, opts...)...,
		)
		if err != nil {
			log.WithError(err).Error("Could not connect to Monitor server")
			close(s.ready)
			return
		}
		s.conn = conn
		close(s.ready)
	}()
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

	mu       sync.RWMutex
	uplink   map[string]chan *router.UplinkMessage
	downlink map[string]chan *router.DownlinkMessage
	status   map[string]chan *gateway.Status
}

func (s *gatewayStreams) Send(msg interface{}) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	switch msg := msg.(type) {
	case *router.UplinkMessage:
		s.log.Debug("Sending UplinkMessage to monitor")
		for serverName, ch := range s.uplink {
			select {
			case ch <- msg:
			default:
				s.log.WithField("Monitor", serverName).Warn("UplinkMessage buffer full")
			}
		}
	case *router.DownlinkMessage:
		s.log.Debug("Sending DownlinkMessage to monitor")
		for serverName, ch := range s.downlink {
			select {
			case ch <- msg:
			default:
				s.log.WithField("Monitor", serverName).Warn("DownlinkMessage buffer full")
			}
		}
	case *gateway.Status:
		s.log.Debug("Sending Status to monitor")
		for serverName, ch := range s.status {
			select {
			case ch <- msg:
			default:
				s.log.WithField("Monitor", serverName).Warn("GatewayStatus buffer full")
			}
		}
	}
}

func (s *gatewayStreams) Close() {
	s.cancel()
}

// NewGatewayStreams returns new streams using the given gateway ID and token
func (c *Client) NewGatewayStreams(id string, token string) GenericStream {
	log := c.log.WithField("GatewayID", id)
	ctx, cancel := context.WithCancel(c.ctx)
	ctx = api.ContextWithID(ctx, id)
	ctx = api.ContextWithToken(ctx, token)
	s := &gatewayStreams{
		log:    log,
		ctx:    ctx,
		cancel: cancel,

		uplink:   make(map[string]chan *router.UplinkMessage),
		downlink: make(map[string]chan *router.DownlinkMessage),
		status:   make(map[string]chan *gateway.Status),
	}

	// Hook up the monitor servers
	for _, server := range c.serverConns {
		go func(server *serverConn) {
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
			log := log.WithField("Monitor", server.name)
			cli := NewMonitorClient(server.conn)

			monitor := func(streamName string, stream grpc.ClientStream) {
				err := stream.RecvMsg(new(empty.Empty))
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
			uplink, err := cli.GatewayUplink(ctx)
			if err != nil {
				log.WithError(err).Warn("Could not set up GatewayUplink stream")
			} else {
				go monitor("GatewayUplink", uplink)
			}

			// Downlink stream
			downlink, err := cli.GatewayDownlink(ctx)
			if err != nil {
				log.WithError(err).Warn("Could not set up GatewayDownlink stream")
			} else {
				go monitor("GatewayDownlink", downlink)
			}

			// Status stream
			status, err := cli.GatewayStatus(ctx)
			if err != nil {
				log.WithError(err).Warn("Could not set up GatewayStatus stream")
			} else {
				go monitor("GatewayStatus", status)
			}

			chUplink := make(chan *router.UplinkMessage, c.config.BufferSize)
			chDownlink := make(chan *router.DownlinkMessage, c.config.BufferSize)
			chStatus := make(chan *gateway.Status, c.config.BufferSize)

			s.mu.Lock()
			s.uplink[server.name] = chUplink
			s.downlink[server.name] = chDownlink
			s.status[server.name] = chStatus
			s.mu.Unlock()

			log.Debug("Start handling Gateway streams")
			defer log.Debug("Done handling Gateway streams")
			for {
				select {
				case <-ctx.Done():
					return
				case msg := <-chStatus:
					if err := status.Send(msg); err != nil {
						log.WithError(err).Warn("Could not send GatewayStatus to monitor")
						return
					}
				case msg := <-chUplink:
					if err := uplink.Send(msg); err != nil {
						log.WithError(err).Warn("Could not send UplinkMessage to monitor")
						return
					}
				case msg := <-chDownlink:
					if err := downlink.Send(msg); err != nil {
						log.WithError(err).Warn("Could not send DownlinkMessage to monitor")
						return
					}
				}
			}

		}(server)
	}

	return s
}

type brokerStreams struct {
	log    log.Interface
	ctx    context.Context
	cancel context.CancelFunc

	mu       sync.RWMutex
	uplink   map[string]chan *broker.DeduplicatedUplinkMessage
	downlink map[string]chan *broker.DownlinkMessage
}

func (s *brokerStreams) Send(msg interface{}) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	switch msg := msg.(type) {
	case *broker.DeduplicatedUplinkMessage:
		s.log.Debug("Sending DeduplicatedUplinkMessage to monitor")
		for serverName, ch := range s.uplink {
			select {
			case ch <- msg:
			default:
				s.log.WithField("Monitor", serverName).Warn("DeduplicatedUplinkMessage buffer full")
			}
		}
	case *broker.DownlinkMessage:
		s.log.Debug("Sending DownlinkMessage to monitor")
		for serverName, ch := range s.downlink {
			select {
			case ch <- msg:
			default:
				s.log.WithField("Monitor", serverName).Warn("DownlinkMessage buffer full")
			}
		}
	}
}

func (s *brokerStreams) Close() {
	s.cancel()
}

// NewBrokerStreams returns new streams using the given broker ID and token
func (c *Client) NewBrokerStreams(id string, token string) GenericStream {
	log := c.log
	ctx, cancel := context.WithCancel(c.ctx)
	ctx = api.ContextWithID(ctx, id)
	ctx = api.ContextWithToken(ctx, token)
	s := &brokerStreams{
		log:    log,
		ctx:    ctx,
		cancel: cancel,

		uplink:   make(map[string]chan *broker.DeduplicatedUplinkMessage),
		downlink: make(map[string]chan *broker.DownlinkMessage),
	}

	// Hook up the monitor servers
	for _, server := range c.serverConns {
		go func(server *serverConn) {
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

			log := log.WithField("Monitor", server.name)
			cli := NewMonitorClient(server.conn)

			monitor := func(streamName string, stream grpc.ClientStream) {
				err := stream.RecvMsg(new(empty.Empty))
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
			uplink, err := cli.BrokerUplink(ctx)
			if err != nil {
				log.WithError(err).Warn("Could not set up BrokerUplink stream")
			} else {
				go monitor("BrokerUplink", uplink)
			}

			// Downlink stream
			downlink, err := cli.BrokerDownlink(ctx)
			if err != nil {
				log.WithError(err).Warn("Could not set up BrokerDownlink stream")
			} else {
				go monitor("BrokerDownlink", downlink)
			}

			chUplink := make(chan *broker.DeduplicatedUplinkMessage, c.config.BufferSize)
			chDownlink := make(chan *broker.DownlinkMessage, c.config.BufferSize)

			s.mu.Lock()
			s.uplink[server.name] = chUplink
			s.downlink[server.name] = chDownlink
			s.mu.Unlock()

			log.Debug("Start handling Broker streams")
			defer log.Debug("Done handling Broker streams")
			for {
				select {
				case <-ctx.Done():
					return
				case msg := <-chUplink:
					if err := uplink.Send(msg); err != nil {
						log.WithError(err).Warn("Could not send UplinkMessage to monitor")
						return
					}
				case msg := <-chDownlink:
					if err := downlink.Send(msg); err != nil {
						log.WithError(err).Warn("Could not send DownlinkMessage to monitor")
						return
					}
				}
			}

		}(server)
	}

	return s
}

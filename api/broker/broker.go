// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package broker

import (
	"context"
	"io"
	"sync"

	"github.com/TheThingsNetwork/go-utils/grpc/restartstream"
	"github.com/TheThingsNetwork/go-utils/grpc/ttnctx"
	"github.com/TheThingsNetwork/go-utils/log"
	"github.com/TheThingsNetwork/ttn/api"
	"github.com/TheThingsNetwork/ttn/utils"
	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

// RouterStream is used for sending uplink and receiving downlink.
type RouterStream interface {
	Uplink(*UplinkMessage)
	Downlink() <-chan *DownlinkMessage
	Close()
}

// HandlerStream is used for sending uplink and receiving downlink.
type HandlerStream interface {
	Uplink() <-chan *DeduplicatedUplinkMessage
	Downlink(*DownlinkMessage)
	Close()
}

// ClientConfig for broker Client
type ClientConfig struct {
	BackgroundContext context.Context
	BufferSize        int
}

// DefaultClientConfig for broker Client
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

// Client for broker
type Client struct {
	log    log.Interface
	ctx    context.Context
	cancel context.CancelFunc

	config      ClientConfig
	serverConns []*serverConn
}

// AddServer adds a broker server
func (c *Client) AddServer(name string, conn *grpc.ClientConn) {
	log := c.log.WithField("Broker", name)
	log.Info("Adding Broker server")
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

func logStreamErr(log log.Interface, streamName string, err error) {
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

type routerStreams struct {
	log    log.Interface
	ctx    context.Context
	cancel context.CancelFunc

	mu       sync.RWMutex
	uplink   map[string]chan *UplinkMessage
	downlink chan *DownlinkMessage
}

func (s *routerStreams) Uplink(msg *UplinkMessage) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	s.log.WithField("Brokers", len(s.uplink)).Debug("Sending UplinkMessage to broker")
	for serverName, ch := range s.uplink {
		select {
		case ch <- msg:
		default:
			s.log.WithField("Broker", serverName).Warn("UplinkMessage buffer full")
		}
	}
}

func (s *routerStreams) Downlink() <-chan *DownlinkMessage {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.downlink
}

func (s *routerStreams) Close() {
	s.cancel()
}

// NewRouterStreams returns new streams using the given router ID and token
func (c *Client) NewRouterStreams(id string, token string) RouterStream {
	log := c.log
	ctx, cancel := context.WithCancel(c.ctx)
	ctx = ttnctx.OutgoingContextWithID(ctx, id)
	ctx = ttnctx.OutgoingContextWithToken(ctx, token)
	s := &routerStreams{
		log:    log,
		ctx:    ctx,
		cancel: cancel,

		uplink:   make(map[string]chan *UplinkMessage),
		downlink: make(chan *DownlinkMessage, c.config.BufferSize),
	}

	var wgDown sync.WaitGroup
	defer func() {
		go func() {
			wgDown.Wait()
			close(s.downlink)
		}()
	}()

	var wg utils.WaitGroup

	// Hook up the broker servers
	for _, server := range c.serverConns {
		wg.Add(1)
		wgDown.Add(1)
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
			log := log.WithField("Broker", server.name)
			cli := NewBrokerClient(server.conn)

			// Stream channels
			chUplink := make(chan *UplinkMessage, c.config.BufferSize)

			defer func() {
				s.mu.Lock()
				defer s.mu.Unlock()
				delete(s.uplink, server.name)
				close(chUplink)
			}()

			// Associate stream
			associate, err := cli.Associate(ctx)
			if err != nil {
				log.WithError(err).Warn("Could not set up Associate stream")
				wgDown.Done()
			} else {
				s.mu.Lock()
				s.uplink[server.name] = chUplink
				s.mu.Unlock()

				// Downlink
				go func() {
					defer func() {
						wgDown.Done()
					}()
					for {
						msg, err := associate.Recv()
						if err != nil {
							logStreamErr(log, "Associate", err)
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

			wg.Done()
			log.Debug("Start handling Associate stream")
			defer log.Debug("Done handling Associate stream")
			for {
				select {
				case <-ctx.Done():
					return
				case msg := <-chUplink:
					if err := associate.Send(msg); err != nil {
						log.WithError(err).Warn("Could not send UplinkMessage to broker")
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

type handlerStreams struct {
	log    log.Interface
	ctx    context.Context
	cancel context.CancelFunc

	mu       sync.RWMutex
	downlink map[string]chan *DownlinkMessage
	uplink   chan *DeduplicatedUplinkMessage
}

func (s *handlerStreams) Downlink(msg *DownlinkMessage) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	s.log.WithField("Brokers", len(s.downlink)).Debug("Sending DownlinkMessage to broker")
	for serverName, ch := range s.downlink {
		select {
		case ch <- msg:
		default:
			s.log.WithField("Broker", serverName).Warn("DownlinkMessage buffer full")
		}
	}
}

func (s *handlerStreams) Uplink() <-chan *DeduplicatedUplinkMessage {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.uplink
}

func (s *handlerStreams) Close() {
	s.cancel()
}

// NewHandlerStreams returns new streams using the given handler ID and token
func (c *Client) NewHandlerStreams(id string, token string) HandlerStream {
	log := c.log
	ctx, cancel := context.WithCancel(c.ctx)
	ctx = ttnctx.OutgoingContextWithID(ctx, id)
	ctx = ttnctx.OutgoingContextWithToken(ctx, token)
	s := &handlerStreams{
		log:    log,
		ctx:    ctx,
		cancel: cancel,

		downlink: make(map[string]chan *DownlinkMessage),
		uplink:   make(chan *DeduplicatedUplinkMessage, c.config.BufferSize),
	}

	var wgUp sync.WaitGroup
	defer func() {
		go func() {
			wgUp.Wait()
			close(s.uplink)
		}()
	}()

	var wg utils.WaitGroup

	// Hook up the broker servers
	for _, server := range c.serverConns {
		wg.Add(1)
		wgUp.Add(1)
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
			log := log.WithField("Broker", server.name)
			cli := NewBrokerClient(server.conn)

			// Stream channels
			chDownlink := make(chan *DownlinkMessage, c.config.BufferSize)

			defer func() {
				s.mu.Lock()
				defer s.mu.Unlock()
				delete(s.downlink, server.name)
				close(chDownlink)
			}()

			// Publish stream
			downlink, err := cli.Publish(ctx)
			if err != nil {
				log.WithError(err).Warn("Could not set up Publish stream")
			} else {
				s.mu.Lock()
				s.downlink[server.name] = chDownlink
				s.mu.Unlock()
				go func() {
					err := downlink.RecvMsg(new(empty.Empty))
					logStreamErr(log, "Publish", err)
					s.mu.Lock()
					defer s.mu.Unlock()
					delete(s.downlink, server.name)
				}()
			}

			// Subscribe stream
			uplink, err := cli.Subscribe(ctx, &SubscribeRequest{})
			if err != nil {
				log.WithError(err).Warn("Could not set up Subscribe stream")
				wgUp.Done()
			} else {
				go func() {
					defer func() {
						wgUp.Done()
					}()
					for {
						msg, err := uplink.Recv()
						if err != nil {
							logStreamErr(log, "Subscribe", err)
							return
						}
						select {
						case s.uplink <- msg:
						default:
							log.Warn("Uplink buffer full")
						}
					}
				}()
			}

			wg.Done()
			log.Debug("Start handling Publish/Subscribe streams")
			defer log.Debug("Done handling Publish/Subscribe streams")
			for {
				select {
				case <-ctx.Done():
					return
				case msg := <-chDownlink:
					if err := downlink.Send(msg); err != nil {
						log.WithError(err).Warn("Could not send DownlinkMessage to broker")
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

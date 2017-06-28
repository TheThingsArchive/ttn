// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package monitor

import (
	"crypto/tls"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/TheThingsNetwork/go-utils/grpc/restartstream"
	"github.com/TheThingsNetwork/go-utils/grpc/ttnctx"
	"github.com/TheThingsNetwork/go-utils/log"
	"github.com/TheThingsNetwork/ttn/api"
	"github.com/TheThingsNetwork/ttn/api/broker"
	"github.com/TheThingsNetwork/ttn/api/gateway"
	"github.com/TheThingsNetwork/ttn/api/handler"
	"github.com/TheThingsNetwork/ttn/api/networkserver"
	"github.com/TheThingsNetwork/ttn/api/pool"
	"github.com/TheThingsNetwork/ttn/api/router"
	"github.com/TheThingsNetwork/ttn/utils"
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
	BackgroundContext context.Context
	BufferSize        int
}

// DefaultClientConfig for monitor Client
var DefaultClientConfig = ClientConfig{
	BackgroundContext: context.Background(),
	BufferSize:        10,
}

// TLSConfig to use
var TLSConfig *tls.Config

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

// Client for monitor
type Client struct {
	log    log.Interface
	ctx    context.Context
	cancel context.CancelFunc

	statusTicker *time.Ticker

	config      ClientConfig
	serverConns []*serverConn
}

// AddServer adds a new monitor server. Supplying DialOptions overrides the default dial options.
// If the default DialOptions are used, TLS will be used to connect to monitors with a "-tls" suffix in their name.
// This function should not be called after streams have been started
func (c *Client) AddServer(name, address string) {
	log := c.log.WithFields(log.Fields{"Monitor": name, "Address": address})
	log.Info("Adding Monitor server")

	s := &serverConn{
		ctx:   log,
		name:  name,
		ready: make(chan struct{}),
	}
	c.serverConns = append(c.serverConns, s)

	go func() {
		var err error
		if strings.HasSuffix(name, "-tls") {
			s.conn, err = pool.Global.DialSecure(address, nil)
		} else {
			s.conn, err = pool.Global.DialInsecure(address)
		}
		if err != nil {
			log.WithError(err).Error("Could not connect to Monitor server")
		}
		close(s.ready)
	}()
}

// SetStatusInterval creates a status ticker
func (c *Client) SetStatusInterval(interval time.Duration) {
	c.statusTicker = time.NewTicker(interval)
}

// TickStatus calls f when a status needs to be sent
func (c *Client) TickStatus(f func()) {
	if c.statusTicker == nil {
		return
	}
	for range c.statusTicker.C {
		f()
	}
}

// AddConn adds a new monitor server on an existing connection
// This function should not be called after streams have been started
func (c *Client) AddConn(name string, conn *grpc.ClientConn) {
	log := c.log.WithFields(log.Fields{"Monitor": name})
	log.Info("Adding Monitor connection")
	c.serverConns = append(c.serverConns, &serverConn{
		ctx:  log,
		name: name,
		conn: conn,
	})
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
	if msg == nil {
		return
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	switch msg := msg.(type) {
	case *router.UplinkMessage:
		if len(s.uplink) == 0 {
			return
		}
		s.log.WithField("Monitors", len(s.uplink)).Debug("Sending UplinkMessage to monitor")
		for serverName, ch := range s.uplink {
			select {
			case ch <- msg:
			default:
				s.log.WithField("Monitor", serverName).Warn("UplinkMessage buffer full")
			}
		}
	case *router.DeviceActivationRequest:
		if len(s.uplink) == 0 {
			return
		}
		s.log.WithField("Monitors", len(s.uplink)).Debug("Sending DeviceActivationRequest->UplinkMessage to monitor")
		for serverName, ch := range s.uplink {
			select {
			case ch <- &router.UplinkMessage{
				Payload:          msg.Payload,
				Message:          msg.Message,
				ProtocolMetadata: msg.ProtocolMetadata,
				GatewayMetadata:  msg.GatewayMetadata,
				Trace:            msg.Trace,
			}:
			default:
				s.log.WithField("Monitor", serverName).Warn("UplinkMessage buffer full")
			}
		}
	case *router.DownlinkMessage:
		if len(s.downlink) == 0 {
			return
		}
		s.log.WithField("Monitors", len(s.downlink)).Debug("Sending DownlinkMessage to monitor")
		for serverName, ch := range s.downlink {
			select {
			case ch <- msg:
			default:
				s.log.WithField("Monitor", serverName).Warn("DownlinkMessage buffer full")
			}
		}
	case *gateway.Status:
		if len(s.status) == 0 {
			return
		}
		s.log.WithField("Monitors", len(s.status)).Debug("Sending Status to monitor")
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
	ctx = ttnctx.OutgoingContextWithID(ctx, id)
	ctx = ttnctx.OutgoingContextWithToken(ctx, token)
	s := &gatewayStreams{
		log:    log,
		ctx:    ctx,
		cancel: cancel,

		uplink:   make(map[string]chan *router.UplinkMessage),
		downlink: make(map[string]chan *router.DownlinkMessage),
		status:   make(map[string]chan *gateway.Status),
	}

	var wg utils.WaitGroup

	// Hook up the monitor servers
	for _, server := range c.serverConns {
		wg.Add(1)
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

			chUplink := make(chan *router.UplinkMessage, c.config.BufferSize)
			chDownlink := make(chan *router.DownlinkMessage, c.config.BufferSize)
			chStatus := make(chan *gateway.Status, c.config.BufferSize)

			defer func() {
				s.mu.Lock()
				defer s.mu.Unlock()
				delete(s.uplink, server.name)
				delete(s.downlink, server.name)
				delete(s.status, server.name)
				close(chUplink)
				close(chDownlink)
				close(chStatus)
			}()

			// Uplink stream
			uplink, err := cli.GatewayUplink(ctx)
			if err != nil {
				log.WithError(err).Warn("Could not set up GatewayUplink stream")
			} else {
				s.mu.Lock()
				s.uplink[server.name] = chUplink
				s.mu.Unlock()
				go func() {
					monitorStream(log, "GatewayUplink", uplink)
					s.mu.Lock()
					defer s.mu.Unlock()
					delete(s.uplink, server.name)
				}()
			}

			// Downlink stream
			downlink, err := cli.GatewayDownlink(ctx)
			if err != nil {
				log.WithError(err).Warn("Could not set up GatewayDownlink stream")
			} else {
				s.mu.Lock()
				s.downlink[server.name] = chDownlink
				s.mu.Unlock()
				go func() {
					monitorStream(log, "GatewayDownlink", downlink)
					s.mu.Lock()
					defer s.mu.Unlock()
					delete(s.downlink, server.name)
				}()
			}

			// Status stream
			status, err := cli.GatewayStatus(ctx)
			if err != nil {
				log.WithError(err).Warn("Could not set up GatewayStatus stream")
			} else {
				s.mu.Lock()
				s.status[server.name] = chStatus
				s.mu.Unlock()
				go func() {
					monitorStream(log, "GatewayStatus", status)
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
						log.WithError(err).Warn("Could not send GatewayStatus to monitor")
						if err == restartstream.ErrStreamClosed {
							return
						}
					}
				case msg := <-chUplink:
					if err := uplink.Send(msg); err != nil {
						log.WithError(err).Warn("Could not send UplinkMessage to monitor")
						if err == restartstream.ErrStreamClosed {
							return
						}
					}
				case msg := <-chDownlink:
					if err := downlink.Send(msg); err != nil {
						log.WithError(err).Warn("Could not send DownlinkMessage to monitor")
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

type routerStreams struct {
	log    log.Interface
	ctx    context.Context
	cancel context.CancelFunc

	mu     sync.RWMutex
	status map[string]chan *router.Status
}

func (s *routerStreams) Send(msg interface{}) {
	if msg == nil {
		return
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	switch msg := msg.(type) {
	case *router.Status:
		if len(s.status) == 0 {
			return
		}
		s.log.WithField("Monitors", len(s.status)).Debug("Sending Status to monitor")
		for serverName, ch := range s.status {
			select {
			case ch <- msg:
			default:
				s.log.WithField("Monitor", serverName).Warn("Status buffer full")
			}
		}
	}
}

func (s *routerStreams) Close() {
	s.cancel()
}

// NewRouterStreams returns new streams using the given router ID and token
func (c *Client) NewRouterStreams(id string, token string) GenericStream {
	log := c.log
	ctx, cancel := context.WithCancel(c.ctx)
	ctx = ttnctx.OutgoingContextWithID(ctx, id)
	ctx = ttnctx.OutgoingContextWithToken(ctx, token)
	s := &routerStreams{
		log:    log,
		ctx:    ctx,
		cancel: cancel,

		status: make(map[string]chan *router.Status),
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

			chStatus := make(chan *router.Status, c.config.BufferSize)

			defer func() {
				s.mu.Lock()
				defer s.mu.Unlock()
				delete(s.status, server.name)
				close(chStatus)
			}()

			// Status stream
			status, err := cli.RouterStatus(ctx)
			if err != nil {
				log.WithError(err).Warn("Could not set up RouterStatus stream")
			} else {
				s.mu.Lock()
				s.status[server.name] = chStatus
				s.mu.Unlock()
				go func() {
					monitorStream(log, "RouterStatus", status)
					s.mu.Lock()
					defer s.mu.Unlock()
					delete(s.status, server.name)
				}()
			}

			log.Debug("Start handling Router streams")
			defer log.Debug("Done handling Router streams")
			for {
				select {
				case <-ctx.Done():
					return
				case msg := <-chStatus:
					if err := status.Send(msg); err != nil {
						log.WithError(err).Warn("Could not send Status to monitor")
						if err == restartstream.ErrStreamClosed {
							return
						}
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
	status   map[string]chan *broker.Status
}

func (s *brokerStreams) Send(msg interface{}) {
	if msg == nil {
		return
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	switch msg := msg.(type) {
	case *broker.DeduplicatedUplinkMessage:
		if len(s.uplink) == 0 {
			return
		}
		s.log.WithField("Monitors", len(s.uplink)).Debug("Sending DeduplicatedUplinkMessage to monitor")
		for serverName, ch := range s.uplink {
			select {
			case ch <- msg:
			default:
				s.log.WithField("Monitor", serverName).Warn("DeduplicatedUplinkMessage buffer full")
			}
		}
	case *broker.DeduplicatedDeviceActivationRequest:
		if len(s.uplink) == 0 {
			return
		}
		s.log.WithField("Monitors", len(s.uplink)).Debug("Sending DeduplicatedDeviceActivationRequest->DeduplicatedUplinkMessage to monitor")
		for serverName, ch := range s.uplink {
			select {
			case ch <- &broker.DeduplicatedUplinkMessage{
				Payload:          msg.Payload,
				Message:          msg.Message,
				DevEui:           msg.DevEui,
				AppEui:           msg.AppEui,
				AppId:            msg.AppId,
				DevId:            msg.DevId,
				ProtocolMetadata: msg.ProtocolMetadata,
				GatewayMetadata:  msg.GatewayMetadata,
				ServerTime:       msg.ServerTime,
				Trace:            msg.Trace,
			}:
			default:
				s.log.WithField("Monitor", serverName).Warn("DeduplicatedUplinkMessage buffer full")
			}
		}
	case *broker.DownlinkMessage:
		if len(s.downlink) == 0 {
			return
		}
		s.log.WithField("Monitors", len(s.downlink)).Debug("Sending DownlinkMessage to monitor")
		for serverName, ch := range s.downlink {
			select {
			case ch <- msg:
			default:
				s.log.WithField("Monitor", serverName).Warn("DownlinkMessage buffer full")
			}
		}
	case *broker.Status:
		if len(s.status) == 0 {
			return
		}
		s.log.WithField("Monitors", len(s.status)).Debug("Sending Status to monitor")
		for serverName, ch := range s.status {
			select {
			case ch <- msg:
			default:
				s.log.WithField("Monitor", serverName).Warn("Status buffer full")
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
	ctx = ttnctx.OutgoingContextWithID(ctx, id)
	ctx = ttnctx.OutgoingContextWithToken(ctx, token)
	s := &brokerStreams{
		log:    log,
		ctx:    ctx,
		cancel: cancel,

		uplink:   make(map[string]chan *broker.DeduplicatedUplinkMessage),
		downlink: make(map[string]chan *broker.DownlinkMessage),
		status:   make(map[string]chan *broker.Status),
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

			chUplink := make(chan *broker.DeduplicatedUplinkMessage, c.config.BufferSize)
			chDownlink := make(chan *broker.DownlinkMessage, c.config.BufferSize)
			chStatus := make(chan *broker.Status, c.config.BufferSize)

			defer func() {
				s.mu.Lock()
				defer s.mu.Unlock()
				delete(s.uplink, server.name)
				delete(s.downlink, server.name)
				delete(s.status, server.name)
				close(chUplink)
				close(chDownlink)
				close(chStatus)
			}()

			// Uplink stream
			uplink, err := cli.BrokerUplink(ctx)
			if err != nil {
				log.WithError(err).Warn("Could not set up BrokerUplink stream")
			} else {
				s.mu.Lock()
				s.uplink[server.name] = chUplink
				s.mu.Unlock()
				go func() {
					monitorStream(log, "BrokerUplink", uplink)
					s.mu.Lock()
					defer s.mu.Unlock()
					delete(s.uplink, server.name)
				}()
			}

			// Downlink stream
			downlink, err := cli.BrokerDownlink(ctx)
			if err != nil {
				log.WithError(err).Warn("Could not set up BrokerDownlink stream")
			} else {
				s.mu.Lock()
				s.downlink[server.name] = chDownlink
				s.mu.Unlock()
				go func() {
					monitorStream(log, "BrokerDownlink", downlink)
					s.mu.Lock()
					defer s.mu.Unlock()
					delete(s.downlink, server.name)
				}()
			}

			// Status stream
			status, err := cli.BrokerStatus(ctx)
			if err != nil {
				log.WithError(err).Warn("Could not set up BrokerStatus stream")
			} else {
				s.mu.Lock()
				s.status[server.name] = chStatus
				s.mu.Unlock()
				go func() {
					monitorStream(log, "BrokerStatus", status)
					s.mu.Lock()
					defer s.mu.Unlock()
					delete(s.status, server.name)
				}()
			}

			log.Debug("Start handling Broker streams")
			defer log.Debug("Done handling Broker streams")
			for {
				select {
				case <-ctx.Done():
					return
				case msg := <-chUplink:
					if err := uplink.Send(msg); err != nil {
						log.WithError(err).Warn("Could not send UplinkMessage to monitor")
						if err == restartstream.ErrStreamClosed {
							return
						}
					}
				case msg := <-chDownlink:
					if err := downlink.Send(msg); err != nil {
						log.WithError(err).Warn("Could not send DownlinkMessage to monitor")
						if err == restartstream.ErrStreamClosed {
							return
						}
					}
				case msg := <-chStatus:
					if err := status.Send(msg); err != nil {
						log.WithError(err).Warn("Could not send Status to monitor")
						if err == restartstream.ErrStreamClosed {
							return
						}
					}
				}
			}

		}(server)
	}

	return s
}

type networkServerStreams struct {
	log    log.Interface
	ctx    context.Context
	cancel context.CancelFunc

	mu     sync.RWMutex
	status map[string]chan *networkserver.Status
}

func (s *networkServerStreams) Send(msg interface{}) {
	if msg == nil {
		return
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	switch msg := msg.(type) {
	case *networkserver.Status:
		if len(s.status) == 0 {
			return
		}
		s.log.WithField("Monitors", len(s.status)).Debug("Sending Status to monitor")
		for serverName, ch := range s.status {
			select {
			case ch <- msg:
			default:
				s.log.WithField("Monitor", serverName).Warn("Status buffer full")
			}
		}
	}
}

func (s *networkServerStreams) Close() {
	s.cancel()
}

// NewNetworkServerStreams returns new streams using the given networkServer ID and token
func (c *Client) NewNetworkServerStreams(id string, token string) GenericStream {
	log := c.log
	ctx, cancel := context.WithCancel(c.ctx)
	ctx = ttnctx.OutgoingContextWithID(ctx, id)
	ctx = ttnctx.OutgoingContextWithToken(ctx, token)
	s := &networkServerStreams{
		log:    log,
		ctx:    ctx,
		cancel: cancel,

		status: make(map[string]chan *networkserver.Status),
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

			chStatus := make(chan *networkserver.Status, c.config.BufferSize)

			defer func() {
				s.mu.Lock()
				defer s.mu.Unlock()
				delete(s.status, server.name)
				close(chStatus)
			}()

			// Status stream
			status, err := cli.NetworkServerStatus(ctx)
			if err != nil {
				log.WithError(err).Warn("Could not set up NetworkServerStatus stream")
			} else {
				s.mu.Lock()
				s.status[server.name] = chStatus
				s.mu.Unlock()
				go func() {
					monitorStream(log, "NetworkServerStatus", status)
					s.mu.Lock()
					defer s.mu.Unlock()
					delete(s.status, server.name)
				}()
			}

			log.Debug("Start handling NetworkServer streams")
			defer log.Debug("Done handling NetworkServer streams")
			for {
				select {
				case <-ctx.Done():
					return
				case msg := <-chStatus:
					if err := status.Send(msg); err != nil {
						log.WithError(err).Warn("Could not send Status to monitor")
						if err == restartstream.ErrStreamClosed {
							return
						}
					}
				}
			}

		}(server)
	}

	return s
}

type handlerStreams struct {
	log    log.Interface
	ctx    context.Context
	cancel context.CancelFunc

	mu       sync.RWMutex
	uplink   map[string]chan *broker.DeduplicatedUplinkMessage
	downlink map[string]chan *broker.DownlinkMessage
	status   map[string]chan *handler.Status
}

func (s *handlerStreams) Send(msg interface{}) {
	if msg == nil {
		return
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	switch msg := msg.(type) {
	case *broker.DeduplicatedUplinkMessage:
		if len(s.uplink) == 0 {
			return
		}
		s.log.WithField("Monitors", len(s.uplink)).Debug("Sending DeduplicatedUplinkMessage to monitor")
		for serverName, ch := range s.uplink {
			select {
			case ch <- msg:
			default:
				s.log.WithField("Monitor", serverName).Warn("DeduplicatedUplinkMessage buffer full")
			}
		}
	case *broker.DeduplicatedDeviceActivationRequest:
		if len(s.uplink) == 0 {
			return
		}
		s.log.WithField("Monitors", len(s.uplink)).Debug("Sending DeduplicatedDeviceActivationRequest->DeduplicatedUplinkMessage to monitor")
		for serverName, ch := range s.uplink {
			select {
			case ch <- &broker.DeduplicatedUplinkMessage{
				Payload:          msg.Payload,
				Message:          msg.Message,
				DevEui:           msg.DevEui,
				AppEui:           msg.AppEui,
				AppId:            msg.AppId,
				DevId:            msg.DevId,
				ProtocolMetadata: msg.ProtocolMetadata,
				GatewayMetadata:  msg.GatewayMetadata,
				ServerTime:       msg.ServerTime,
				Trace:            msg.Trace,
			}:
			default:
				s.log.WithField("Monitor", serverName).Warn("DeduplicatedUplinkMessage buffer full")
			}
		}
	case *broker.DownlinkMessage:
		if len(s.downlink) == 0 {
			return
		}
		s.log.WithField("Monitors", len(s.downlink)).Debug("Sending DownlinkMessage to monitor")
		for serverName, ch := range s.downlink {
			select {
			case ch <- msg:
			default:
				s.log.WithField("Monitor", serverName).Warn("DownlinkMessage buffer full")
			}
		}
	case *handler.Status:
		if len(s.status) == 0 {
			return
		}
		s.log.WithField("Monitors", len(s.status)).Debug("Sending Status to monitor")
		for serverName, ch := range s.status {
			select {
			case ch <- msg:
			default:
				s.log.WithField("Monitor", serverName).Warn("Status buffer full")
			}
		}
	}
}

func (s *handlerStreams) Close() {
	s.cancel()
}

// NewHandlerStreams returns new streams using the given handler ID and token
func (c *Client) NewHandlerStreams(id string, token string) GenericStream {
	log := c.log
	ctx, cancel := context.WithCancel(c.ctx)
	ctx = ttnctx.OutgoingContextWithID(ctx, id)
	ctx = ttnctx.OutgoingContextWithToken(ctx, token)
	s := &handlerStreams{
		log:    log,
		ctx:    ctx,
		cancel: cancel,

		uplink:   make(map[string]chan *broker.DeduplicatedUplinkMessage),
		downlink: make(map[string]chan *broker.DownlinkMessage),
		status:   make(map[string]chan *handler.Status),
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

			chUplink := make(chan *broker.DeduplicatedUplinkMessage, c.config.BufferSize)
			chDownlink := make(chan *broker.DownlinkMessage, c.config.BufferSize)
			chStatus := make(chan *handler.Status, c.config.BufferSize)

			defer func() {
				s.mu.Lock()
				defer s.mu.Unlock()
				delete(s.uplink, server.name)
				delete(s.downlink, server.name)
				delete(s.status, server.name)
				close(chUplink)
				close(chDownlink)
				close(chStatus)
			}()

			// Uplink stream
			uplink, err := cli.HandlerUplink(ctx)
			if err != nil {
				log.WithError(err).Warn("Could not set up HandlerUplink stream")
			} else {
				s.mu.Lock()
				s.uplink[server.name] = chUplink
				s.mu.Unlock()
				go func() {
					monitorStream(log, "HandlerUplink", uplink)
					s.mu.Lock()
					defer s.mu.Unlock()
					delete(s.uplink, server.name)
				}()
			}

			// Downlink stream
			downlink, err := cli.HandlerDownlink(ctx)
			if err != nil {
				log.WithError(err).Warn("Could not set up HandlerDownlink stream")
			} else {
				s.mu.Lock()
				s.downlink[server.name] = chDownlink
				s.mu.Unlock()
				go func() {
					monitorStream(log, "HandlerDownlink", downlink)
					s.mu.Lock()
					defer s.mu.Unlock()
					delete(s.downlink, server.name)
				}()
			}

			// Status stream
			status, err := cli.HandlerStatus(ctx)
			if err != nil {
				log.WithError(err).Warn("Could not set up HandlerStatus stream")
			} else {
				s.mu.Lock()
				s.status[server.name] = chStatus
				s.mu.Unlock()
				go func() {
					monitorStream(log, "HandlerStatus", status)
					s.mu.Lock()
					defer s.mu.Unlock()
					delete(s.status, server.name)
				}()
			}

			log.Debug("Start handling Handler streams")
			defer log.Debug("Done handling Handler streams")
			for {
				select {
				case <-ctx.Done():
					return
				case msg := <-chUplink:
					if err := uplink.Send(msg); err != nil {
						log.WithError(err).Warn("Could not send UplinkMessage to monitor")
						if err == restartstream.ErrStreamClosed {
							return
						}
					}
				case msg := <-chDownlink:
					if err := downlink.Send(msg); err != nil {
						log.WithError(err).Warn("Could not send DownlinkMessage to monitor")
						if err == restartstream.ErrStreamClosed {
							return
						}
					}
				case msg := <-chStatus:
					if err := status.Send(msg); err != nil {
						log.WithError(err).Warn("Could not send Status to monitor")
						if err == restartstream.ErrStreamClosed {
							return
						}
					}
				}
			}

		}(server)
	}

	return s
}

func monitorStream(log log.Interface, streamName string, stream grpc.ClientStream) {
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

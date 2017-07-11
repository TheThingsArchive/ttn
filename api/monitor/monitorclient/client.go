// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package monitorclient

import (
	"context"
	"net"
	"strings"
	"sync"

	"github.com/TheThingsNetwork/go-utils/grpc/sendbuffer"
	"github.com/TheThingsNetwork/go-utils/log"
	"github.com/TheThingsNetwork/ttn/api/broker"
	"github.com/TheThingsNetwork/ttn/api/gateway"
	"github.com/TheThingsNetwork/ttn/api/handler"
	"github.com/TheThingsNetwork/ttn/api/monitor"
	"github.com/TheThingsNetwork/ttn/api/networkserver"
	"github.com/TheThingsNetwork/ttn/api/pool"
	"github.com/TheThingsNetwork/ttn/api/router"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// DefaultBufferSize sets the default buffer size per stream
var DefaultBufferSize = 10

// MonitorOption changes something in the MonitorClient
type MonitorOption func(m *MonitorClient)

// WithConn adds the given conns to the MonitorClient
func WithConn(name string, conn *grpc.ClientConn) MonitorOption {
	return func(m *MonitorClient) {
		m.clients[name] = monitor.NewMonitorClient(conn)
	}
}

// WithServer [DEPRECATED] connects to the given server and adds it to the MonitorClient
// Instead of using WithServer, you should set up the gRPC connection externally and use
// WithConn to add it to the MonitorClient.
func WithServer(name, addr string, opts ...grpc.DialOption) MonitorOption {
	if len(opts) == 0 {
		if strings.HasSuffix(name, "-tls") {
			netHost, _, _ := net.SplitHostPort(addr)
			creds := credentials.NewTLS(pool.TLSConfig(netHost))
			opts = append(opts, grpc.WithTransportCredentials(creds))
		} else {
			opts = append(opts, grpc.WithInsecure())
		}
	}
	return func(m *MonitorClient) {
		conn, err := grpc.Dial(addr, opts...)
		if err != nil {
			return
		}
		m.clients[name] = monitor.NewMonitorClient(conn)
	}
}

// NewMonitorClient returns a new MonitorClient
func NewMonitorClient(opts ...MonitorOption) *MonitorClient {
	m := &MonitorClient{
		log:        log.Get(),
		bufferSize: DefaultBufferSize,
		clients:    make(map[string]monitor.MonitorClient),
	}
	for _, opt := range opts {
		opt(m)
	}
	return m
}

// MonitorClient is an overlay on top of the gRPC client
type MonitorClient struct {
	log        log.Interface
	bufferSize int
	clients    map[string]monitor.MonitorClient
}

// Stream interface allows sending anything with Send()
type Stream interface {
	Send(msg interface{})
	Open()
	Close()
	Reset()
}

type componentClient struct {
	log log.Interface
	mu  sync.RWMutex

	status   []*sendbuffer.Stream
	uplink   []*sendbuffer.Stream
	downlink []*sendbuffer.Stream

	setup  func()
	cancel context.CancelFunc
}

func (c *componentClient) Open() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.cancel != nil {
		return
	}
	c.setup()
}

func (c *componentClient) Close() {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.cancel == nil {
		return
	}
	c.cancel()
}

func (c *componentClient) Reset() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.cancel == nil {
		return
	}
	c.cancel()
	c.setup()
}

func (c *componentClient) Send(msg interface{}) {
	switch msg := msg.(type) {
	case *gateway.Status, *router.Status, *broker.Status, *networkserver.Status, *handler.Status:
		if len(c.status) == 0 {
			return
		}
		for _, cli := range c.status {
			cli.SendMsg(msg)
		}
		c.log.Debug("Forwarded status to monitor")
	case *router.UplinkMessage, *broker.DeduplicatedUplinkMessage:
		if len(c.uplink) == 0 {
			return
		}
		for _, cli := range c.uplink {
			cli.SendMsg(msg)
		}
		c.log.Debug("Forwarded uplink to monitor")
	case *router.DeviceActivationRequest:
		if len(c.uplink) == 0 {
			return
		}
		asUplink := &router.UplinkMessage{
			Payload:          msg.Payload,
			Message:          msg.Message,
			ProtocolMetadata: msg.ProtocolMetadata,
			GatewayMetadata:  msg.GatewayMetadata,
			Trace:            msg.Trace,
		}
		for _, cli := range c.uplink {
			cli.SendMsg(asUplink)
		}
		c.log.Debug("Forwarded activation as uplink to monitor")
	case *broker.DeduplicatedDeviceActivationRequest:
		if len(c.uplink) == 0 {
			return
		}
		asUplink := &broker.DeduplicatedUplinkMessage{
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
		}
		for _, cli := range c.uplink {
			cli.SendMsg(asUplink)
		}
		c.log.Debug("Forwarded activation as uplink to monitor")
	case *router.DownlinkMessage, *broker.DownlinkMessage:
		if len(c.uplink) == 0 {
			return
		}
		for _, cli := range c.downlink {
			cli.SendMsg(msg)
		}
		c.log.Debug("Forwarded downlink to monitor")
	default:
		c.log.Warnf("Unknown message type: %T", msg)
	}
}

func (c *componentClient) run(monitor, stream string, buf *sendbuffer.Stream) {
	if err := buf.Run(); err != nil && err != context.Canceled {
		c.log.WithField("Monitor", monitor).WithError(err).Warnf("%s stream failed", buf)
	}
}

// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package monitorclient

import (
	"context"
	"io"
	"net"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/TheThingsNetwork/go-utils/backoff"
	"github.com/TheThingsNetwork/go-utils/grpc/rpclog"
	"github.com/TheThingsNetwork/go-utils/grpc/streambuffer"
	"github.com/TheThingsNetwork/go-utils/log"
	"github.com/TheThingsNetwork/ttn/api/broker"
	"github.com/TheThingsNetwork/ttn/api/gateway"
	"github.com/TheThingsNetwork/ttn/api/handler"
	"github.com/TheThingsNetwork/ttn/api/monitor"
	"github.com/TheThingsNetwork/ttn/api/networkserver"
	"github.com/TheThingsNetwork/ttn/api/pool"
	"github.com/TheThingsNetwork/ttn/api/router"
	"github.com/mwitkow/go-grpc-middleware"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
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
		opts = append(opts,
			grpc.WithUnaryInterceptor(grpc_middleware.ChainUnaryClient(
				rpclog.UnaryClientInterceptor(nil),
			)),
			grpc.WithStreamInterceptor(grpc_middleware.ChainStreamClient(
				rpclog.StreamClientInterceptor(nil),
			)),
		)
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

	status   []*streambuffer.Stream
	uplink   []*streambuffer.Stream
	downlink []*streambuffer.Stream

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

func (c *componentClient) disable(s *streambuffer.Stream) {
	c.mu.Lock()
	defer c.mu.Unlock()
	curStatus := c.status
	c.status = curStatus[:0]
	for _, status := range curStatus {
		if status != s {
			c.status = append(c.status, status)
		}
	}
	curUplink := c.uplink
	c.uplink = curUplink[:0]
	for _, uplink := range curUplink {
		if uplink != s {
			c.uplink = append(c.uplink, uplink)
		}
	}
	curDownlink := c.downlink
	c.downlink = curDownlink[:0]
	for _, downlink := range curDownlink {
		if downlink != s {
			c.downlink = append(c.downlink, downlink)
		}
	}
}

func (c *componentClient) Send(msg interface{}) {
	c.mu.RLock()
	status, uplink, downlink := c.status, c.uplink, c.downlink
	c.mu.RUnlock()
	switch msg := msg.(type) {
	case *gateway.Status, *router.Status, *broker.Status, *networkserver.Status, *handler.Status:
		if len(status) == 0 {
			return
		}
		for _, cli := range status {
			cli.SendMsg(msg)
		}
		c.log.Debug("Forwarded status to monitor")
	case *router.UplinkMessage, *broker.DeduplicatedUplinkMessage:
		if len(uplink) == 0 {
			return
		}
		for _, cli := range uplink {
			cli.SendMsg(msg)
		}
		c.log.Debug("Forwarded uplink to monitor")
	case *router.DeviceActivationRequest:
		if len(uplink) == 0 {
			return
		}
		asUplink := &router.UplinkMessage{
			Payload:          msg.Payload,
			Message:          msg.Message,
			ProtocolMetadata: msg.ProtocolMetadata,
			GatewayMetadata:  msg.GatewayMetadata,
			Trace:            msg.Trace,
		}
		for _, cli := range uplink {
			cli.SendMsg(asUplink)
		}
		c.log.Debug("Forwarded activation as uplink to monitor")
	case *broker.DeduplicatedDeviceActivationRequest:
		if len(uplink) == 0 {
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
		for _, cli := range uplink {
			cli.SendMsg(asUplink)
		}
		c.log.Debug("Forwarded activation as uplink to monitor")
	case *router.DownlinkMessage, *broker.DownlinkMessage:
		if len(uplink) == 0 {
			return
		}
		for _, cli := range downlink {
			cli.SendMsg(msg)
		}
		c.log.Debug("Forwarded downlink to monitor")
	default:
		c.log.Warnf("Unknown message type: %T", msg)
	}
}

var stableTimeout = 10 * time.Second

func (c *componentClient) run(monitor, stream string, buf *streambuffer.Stream) {
	var t *time.Timer
	var streamErrors int32
	for {
		t = time.AfterFunc(stableTimeout, func() { atomic.SwapInt32(&streamErrors, 0) })
		err := buf.Run()
		t.Stop()
		if err == nil || err == context.Canceled || err == io.EOF {
			return
		}
		switch grpc.Code(err) {
		case codes.Unknown, codes.Aborted, codes.Unavailable:
			c.log.WithField("Monitor", monitor).WithError(err).Debugf("%s stream failed temporarily", stream)
			new := atomic.AddInt32(&streamErrors, 1)
			time.Sleep(backoff.Backoff(int(new - 1)))
		default:
			c.log.WithField("Monitor", monitor).WithError(err).Warnf("%s stream failed permanently", stream)
			c.disable(buf)
			return
		}
	}
}

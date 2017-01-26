// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package monitor

import (
	"sync"

	ttnlog "github.com/TheThingsNetwork/go-utils/log"
	"github.com/TheThingsNetwork/ttn/api"
	"github.com/TheThingsNetwork/ttn/api/broker"
	"github.com/TheThingsNetwork/ttn/api/gateway"
	"github.com/TheThingsNetwork/ttn/api/router"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	"golang.org/x/net/context" // See https://github.com/grpc/grpc-go/issues/711"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// BufferSize gives the size for the monitor buffers
const BufferSize = 10

// Client is a wrapper around MonitorClient
type Client struct {
	Ctx ttnlog.Interface

	client MonitorClient
	conn   *grpc.ClientConn
	addr   string

	gateways     map[string]GatewayClient
	BrokerClient BrokerClient

	mutex sync.RWMutex
}

// NewClient is a wrapper for NewMonitorClient, initializes
// connection to MonitorServer on monitorAddr with default gRPC options
func NewClient(ctx ttnlog.Interface, monitorAddr string) (cl *Client, err error) {
	cl = &Client{
		Ctx:      ctx,
		addr:     monitorAddr,
		gateways: make(map[string]GatewayClient),
	}
	return cl, cl.Open()
}

// Open opens connection to the monitor
func (cl *Client) Open() (err error) {
	cl.mutex.Lock()
	defer cl.mutex.Unlock()

	return cl.open()
}
func (cl *Client) open() (err error) {
	addr := cl.addr
	ctx := cl.Ctx.WithField("addr", addr)

	defer func() {
		if err != nil {
			ctx.Warn("Failed to open monitor connection")
		} else {
			ctx.Info("Monitor connection opened")
		}
	}()

	ctx.Debug("Opening monitor connection...")

	cl.conn, err = api.Dial(addr)
	if err != nil {
		ctx.WithError(errors.FromGRPCError(err)).Warn("Failed to establish connection to gRPC service")
		return err
	}

	cl.client = NewMonitorClient(cl.conn)

	cl.BrokerClient = &brokerClient{
		Ctx:    cl.Ctx.WithField("component", "broker"),
		client: cl,
	}
	return nil
}

// Close closes connection to the monitor
func (cl *Client) Close() (err error) {
	cl.mutex.Lock()
	defer cl.mutex.Unlock()

	return cl.close()
}
func (cl *Client) close() (err error) {
	defer func() {
		if err != nil {
			cl.Ctx.Warn("Failed to close monitor connection")
		} else {
			cl.Ctx.Info("Monitor connection closed")
		}
	}()

	for _, gtw := range cl.gateways {
		ctx := cl.Ctx.WithField("GatewayID", gtw.(*gatewayClient).id)

		ctx.Debug("Closing gateway streams...")
		err = gtw.Close()
		if err != nil {
			ctx.Warn("Failed to close gateway streams")
		}
	}

	cl.Ctx.Debug("Closing monitor connection...")
	err = cl.conn.Close()
	if err != nil {
		return err
	}

	cl.conn = nil
	return nil
}

// Reopen reopens connection to the monitor. It first attempts to close already opened connection
// and then opens a new one. If closing already opened connection fails, Reopen fails too.
func (cl *Client) Reopen() (err error) {
	cl.mutex.Lock()
	defer cl.mutex.Unlock()

	return cl.reopen()
}
func (cl *Client) reopen() (err error) {
	defer func() {
		if err != nil {
			cl.Ctx.Warn("Failed to reopen monitor connection")
		} else {
			cl.Ctx.Info("Monitor connection reopened")
		}
	}()

	cl.Ctx.Debug("Reopening monitor connection...")

	err = cl.close()
	if err != nil {
		return err
	}
	return cl.open()
}

// IsConnected returns whether connection to the monitor had been established or not
func (cl *Client) IsConnected() bool {
	return cl.client != nil && cl.conn != nil
}

// GatewayClient returns monitor GatewayClient for id and token specified
func (cl *Client) GatewayClient(id string) (gtwCl GatewayClient) {
	cl.mutex.RLock()
	gtwCl, ok := cl.gateways[id]
	cl.mutex.RUnlock()
	if !ok {
		cl.mutex.Lock()
		gtwCl = &gatewayClient{
			Ctx:    cl.Ctx.WithField("GatewayID", id),
			client: cl,
			id:     id,
		}
		cl.gateways[id] = gtwCl
		cl.mutex.Unlock()
	}
	return gtwCl
}

type gatewayClient struct {
	sync.RWMutex

	client *Client

	Ctx ttnlog.Interface

	id, token string

	status struct {
		init   sync.Once
		ch     chan *gateway.Status
		cancel func()
		sync.RWMutex
	}

	uplink struct {
		init   sync.Once
		ch     chan *router.UplinkMessage
		cancel func()
		sync.Mutex
	}

	downlink struct {
		init   sync.Once
		ch     chan *router.DownlinkMessage
		cancel func()
		sync.RWMutex
	}
}

// GatewayClient is used as the main client for Gateways to communicate with the monitor
type GatewayClient interface {
	SetToken(token string)
	IsConfigured() bool
	SendStatus(status *gateway.Status) (err error)
	SendUplink(msg *router.UplinkMessage) (err error)
	SendDownlink(msg *router.DownlinkMessage) (err error)
	Close() (err error)
}

func (cl *gatewayClient) SetToken(token string) {
	cl.Lock()
	defer cl.Unlock()
	cl.token = token
}

func (cl *gatewayClient) IsConfigured() bool {
	cl.RLock()
	defer cl.RUnlock()
	return cl.token != ""
}

// Close closes all opened monitor streams for the gateway
func (cl *gatewayClient) Close() (err error) {
	cl.closeStatus()
	cl.closeUplink()
	cl.closeDownlink()
	return err
}

// Context returns monitor connection context for gateway
func (cl *gatewayClient) Context() (monitorContext context.Context) {
	cl.RLock()
	defer cl.RUnlock()
	return metadata.NewContext(context.Background(), metadata.Pairs(
		"id", cl.id,
		"token", cl.token,
	))
}

type brokerClient struct {
	sync.RWMutex

	client *Client

	Ctx ttnlog.Interface

	uplink struct {
		init   sync.Once
		ch     chan *broker.DeduplicatedUplinkMessage
		cancel func()
		sync.Mutex
	}

	downlink struct {
		init   sync.Once
		ch     chan *broker.DownlinkMessage
		cancel func()
		sync.RWMutex
	}
}

// BrokerClient is used as the main client for Brokers to communicate with the monitor
type BrokerClient interface {
	SendUplink(msg *broker.DeduplicatedUplinkMessage) (err error)
	SendDownlink(msg *broker.DownlinkMessage) (err error)
	Close() (err error)
}

// Close closes all opened monitor streams for the broker
func (cl *brokerClient) Close() (err error) {
	cl.closeUplink()
	cl.closeDownlink()
	return err
}

// Context returns monitor connection context for broker
func (cl *brokerClient) Context() (monitorContext context.Context) {
	//TODO add auth
	return context.Background()
}

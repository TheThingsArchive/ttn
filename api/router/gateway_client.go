// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package router

import (
	"sync"

	"github.com/TheThingsNetwork/go-utils/log"
	"golang.org/x/net/context"
	"google.golang.org/grpc/metadata"
)

// RouterClientForGateway is a RouterClient for a specific gateway
type RouterClientForGateway interface {
	Close()

	GetLogger() log.Interface
	SetLogger(log.Interface)

	SetToken(token string)
	TokenChanged() <-chan struct{}

	GatewayStatus() (Router_GatewayStatusClient, error)
	Uplink() (Router_UplinkClient, error)
	Subscribe() (Router_SubscribeClient, context.CancelFunc, error)
	Activate(in *DeviceActivationRequest) (*DeviceActivationResponse, error)
}

// NewRouterClientForGateway returns a new RouterClient for the given gateway ID and access token
func NewRouterClientForGateway(client RouterClient, gatewayID, token string) RouterClientForGateway {
	ctx, cancel := context.WithCancel(context.Background())
	return &routerClientForGateway{
		ctx:          log.Get().WithField("GatewayID", gatewayID),
		client:       client,
		gatewayID:    gatewayID,
		token:        token,
		tokenChanged: make(chan struct{}),
		bgCtx:        ctx,
		cancel:       cancel,
	}
}

type routerClientForGateway struct {
	ctx          log.Interface
	client       RouterClient
	gatewayID    string
	token        string
	tokenChanged chan struct{}
	bgCtx        context.Context
	cancel       context.CancelFunc
	mu           sync.RWMutex
}

func (c *routerClientForGateway) Close() {
	c.cancel()
}

func (c *routerClientForGateway) GetLogger() log.Interface {
	return c.ctx
}

func (c *routerClientForGateway) SetLogger(logger log.Interface) {
	c.ctx = logger
}

func (c *routerClientForGateway) getContext() context.Context {
	c.mu.RLock()
	defer c.mu.RUnlock()
	md := metadata.Pairs(
		"id", c.gatewayID,
		"token", c.token,
	)
	return metadata.NewContext(c.bgCtx, md)
}

func (c *routerClientForGateway) SetToken(token string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.token = token
	close(c.tokenChanged)
	c.tokenChanged = make(chan struct{})
}

func (c *routerClientForGateway) TokenChanged() <-chan struct{} {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.tokenChanged
}

func (c *routerClientForGateway) GatewayStatus() (Router_GatewayStatusClient, error) {
	c.ctx.Debug("Starting GatewayStatus stream")
	return c.client.GatewayStatus(c.getContext())
}

func (c *routerClientForGateway) Uplink() (Router_UplinkClient, error) {
	c.ctx.Debug("Starting Uplink stream")
	return c.client.Uplink(c.getContext())
}

func (c *routerClientForGateway) Subscribe() (Router_SubscribeClient, context.CancelFunc, error) {
	c.ctx.Debug("Starting Subscribe stream")
	ctx, cancel := context.WithCancel(c.getContext())
	client, err := c.client.Subscribe(ctx, &SubscribeRequest{})
	return client, cancel, err
}

func (c *routerClientForGateway) Activate(in *DeviceActivationRequest) (*DeviceActivationResponse, error) {
	c.ctx.Debug("Calling Activate")
	return c.client.Activate(c.getContext(), in)
}

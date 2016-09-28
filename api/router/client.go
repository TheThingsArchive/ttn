// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package router

import (
	"io"
	"sync"

	"github.com/TheThingsNetwork/ttn/api"
	"github.com/TheThingsNetwork/ttn/api/discovery"
	"github.com/TheThingsNetwork/ttn/api/gateway"
	"github.com/TheThingsNetwork/ttn/utils/errors"

	"golang.org/x/net/context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
)

// GatewayClient is used as the main client for Gateways to communicate with the Router
type GatewayClient interface {
	SendGatewayStatus(*gateway.Status) error
	SendUplink(*UplinkMessage) error
	Subscribe() (<-chan *DownlinkMessage, <-chan error, error)
	Unsubscribe() error
	Activate(*DeviceActivationRequest) (*DeviceActivationResponse, error)
	Close() error
}

// NewClient returns a new Client
func NewClient(routerAnnouncement *discovery.Announcement) (*Client, error) {
	conn, err := routerAnnouncement.Dial()
	if err != nil {
		return nil, err
	}
	client := &Client{
		conn:   conn,
		client: NewRouterClient(conn),
	}
	return client, nil
}

// Client is a wrapper around RouterClient
type Client struct {
	mutex    sync.Mutex
	conn     *grpc.ClientConn
	client   RouterClient
	gateways []GatewayClient
}

// ForGateway returns a GatewayClient that is configured for the provided gateway
func (c *Client) ForGateway(gatewayID string, tokenFunc func() string) GatewayClient {
	defer c.mutex.Unlock()
	c.mutex.Lock()
	gatewayClient := NewGatewayClient(c.client, gatewayID, tokenFunc)
	c.gateways = append(c.gateways, gatewayClient)
	return gatewayClient
}

// Close purges the cache and closes the connection with the Router
func (c *Client) Close() error {
	defer c.mutex.Unlock()
	c.mutex.Lock()
	for _, gateway := range c.gateways {
		gateway.Close()
	}
	return c.conn.Close()
}

// NewGatewayClient returns a new GatewayClient
func NewGatewayClient(client RouterClient, gatewayID string, tokenFunc func() string) GatewayClient {
	gatewayClient := &gatewayClient{
		client:    client,
		id:        gatewayID,
		tokenFunc: tokenFunc,
	}
	return gatewayClient
}

type gatewayClient struct {
	mutex             sync.Mutex
	id                string
	tokenFunc         func() string
	client            RouterClient
	gatewayStatus     Router_GatewayStatusClient
	stopGatewayStatus chan bool
	uplink            Router_UplinkClient
	stopUplink        chan bool
	downlink          Router_SubscribeClient
	stopDownlink      chan bool
}

func (c *gatewayClient) getContext() context.Context {
	md := metadata.Pairs(
		"id", c.id,
		"token", c.tokenFunc(),
	)
	gatewayContext := metadata.NewContext(context.Background(), md)
	return gatewayContext
}

func (c *gatewayClient) setupGatewayStatus() error {
	api.GetLogger().Debugf("Setting up gateway status stream for %s...", c.id)
	gatewayStatusClient, err := c.client.GatewayStatus(c.getContext())
	if err != nil {
		return err
	}
	c.gatewayStatus = gatewayStatusClient
	c.stopGatewayStatus = make(chan bool)
	go func() {
		var msg interface{}
		for {
			select {
			case <-c.stopGatewayStatus:
				return
			default:
				if err := gatewayStatusClient.RecvMsg(msg); err != nil {
					api.GetLogger().Warnf("Error in gateway status stream for %s: %s", c.id, err.Error())
					c.teardownGatewayStatus()
					return
				}
				api.GetLogger().Debugf("Received: %v", msg)
			}
		}
	}()
	return nil
}

func (c *gatewayClient) teardownGatewayStatus() {
	defer c.mutex.Unlock()
	c.mutex.Lock()
	if c.gatewayStatus != nil {
		api.GetLogger().Debugf("Closing gateway status stream for %s...", c.id)
		close(c.stopGatewayStatus)
		c.gatewayStatus.CloseSend()
		c.gatewayStatus = nil
	}
}

func (c *gatewayClient) SendGatewayStatus(status *gateway.Status) error {
	defer c.mutex.Unlock()
	c.mutex.Lock()
	if c.gatewayStatus == nil {
		if err := c.setupGatewayStatus(); err != nil {
			return errors.FromGRPCError(err)
		}
	}
	if err := c.gatewayStatus.Send(status); err != nil {
		if err == io.EOF {
			api.GetLogger().Warnf("Could not send gateway status for %s on closed stream", c.id)
			go c.teardownGatewayStatus()
			return errors.FromGRPCError(err)
		}
		api.GetLogger().Warnf("Error sending gateway status for %s: %s", c.id, err.Error())
		return errors.FromGRPCError(err)
	}
	return nil
}

func (c *gatewayClient) setupUplink() error {
	api.GetLogger().Debugf("Setting up uplink stream for %s...", c.id)
	uplinkClient, err := c.client.Uplink(c.getContext())
	if err != nil {
		return err
	}
	c.uplink = uplinkClient
	c.stopUplink = make(chan bool)
	go func() {
		var msg interface{}
		for {
			select {
			case <-c.stopUplink:
				return
			default:
				if err := uplinkClient.RecvMsg(msg); err != nil {
					api.GetLogger().Warnf("Error in uplink stream for %s: %s", c.id, err.Error())
					c.teardownUplink()
					return
				}
				api.GetLogger().Debugf("Received: %v", msg)

			}
		}
	}()
	return nil
}

func (c *gatewayClient) teardownUplink() {
	defer c.mutex.Unlock()
	c.mutex.Lock()
	if c.uplink != nil {
		api.GetLogger().Debugf("Closing uplink stream for %s...", c.id)
		close(c.stopUplink)
		c.uplink.CloseSend()
		c.uplink = nil
	}
}

func (c *gatewayClient) SendUplink(uplink *UplinkMessage) error {
	defer c.mutex.Unlock()
	c.mutex.Lock()
	if c.uplink == nil {
		if err := c.setupUplink(); err != nil {
			return errors.FromGRPCError(err)
		}
	}
	if err := c.uplink.Send(uplink); err != nil {
		if err == io.EOF {
			api.GetLogger().Warnf("Could not send uplink for %s on closed stream", c.id)
			go c.teardownUplink()
			return errors.FromGRPCError(err)
		}
		api.GetLogger().Warnf("Error sending uplink for %s: %s", c.id, err.Error())
		return errors.FromGRPCError(err)
	}
	return nil
}

func (c *gatewayClient) setupDownlink() error {
	api.GetLogger().Debugf("Setting up downlink stream for %s...", c.id)
	ctx, cancel := context.WithCancel(c.getContext())
	downlinkClient, err := c.client.Subscribe(ctx, &SubscribeRequest{})
	if err != nil {
		return err
	}
	c.stopDownlink = make(chan bool)
	go func() {
		<-c.stopDownlink
		cancel()
	}()
	c.downlink = downlinkClient
	return nil
}

func (c *gatewayClient) teardownDownlink() {
	defer c.mutex.Unlock()
	c.mutex.Lock()
	if c.downlink != nil {
		api.GetLogger().Debugf("Closing downlink stream for %s...", c.id)
		close(c.stopDownlink)
		c.downlink.CloseSend()
		c.downlink = nil
	}
}

func (c *gatewayClient) Subscribe() (<-chan *DownlinkMessage, <-chan error, error) {
	defer c.mutex.Unlock()
	c.mutex.Lock()
	if c.downlink == nil {
		if err := c.setupDownlink(); err != nil {
			return nil, nil, errors.FromGRPCError(err)
		}
	}
	downChan := make(chan *DownlinkMessage)
	errChan := make(chan error)
	go func() {
		defer func() {
			close(downChan)
			close(errChan)
		}()
		for {
			select {
			case <-c.stopDownlink:
				return
			default:
				downlink, err := c.downlink.Recv()
				if err != nil {
					if grpc.Code(err) == codes.Canceled {
						api.GetLogger().Debugf("Downlink stream for %s was canceled", c.id)
						errChan <- nil
					} else {
						api.GetLogger().Warnf("Error receiving gateway downlink for %s: %s", c.id, err.Error())
						errChan <- errors.FromGRPCError(err)
					}
					c.teardownDownlink()
					return
				}
				downChan <- downlink
			}
		}
	}()
	return downChan, errChan, nil
}

func (c *gatewayClient) Unsubscribe() error {
	c.teardownDownlink()
	return nil
}

func (c *gatewayClient) Activate(req *DeviceActivationRequest) (*DeviceActivationResponse, error) {
	return c.client.Activate(c.getContext(), req)
}

func (c *gatewayClient) Close() error {
	c.teardownGatewayStatus()
	c.teardownUplink()
	c.teardownDownlink()
	return nil
}

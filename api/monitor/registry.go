// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package monitor

import (
	"sync"

	ttnlog "github.com/TheThingsNetwork/go-utils/log"
)

// Registry encapsulates dealing with monitor servers that might be down during startup.
type Registry interface {
	// InitClient initializes a new monitor client.
	InitClient(name string, addr string)

	// BrokerClients returns the list of broker monitor clients that have currently been initialized.
	BrokerClients() []BrokerClient

	// GatewayClients returns the list of gateway monitor clients that have currently been initialized for a given gateway.
	GatewayClients(id string) []GatewayClient

	// SetGatewayToken configures a token that's applied to all clients for a gateway, regardless of when it's initialized.
	SetGatewayToken(id string, token string)
}

// NewRegistry creates a monitor client registry.
func NewRegistry(ctx ttnlog.Interface) Registry {
	return &registry{
		ctx:              ctx,
		monitorClients:   make(map[string]*Client),
		brokerClients:    make([]BrokerClient, 0),
		gatewayClients:   make(map[string][]GatewayClient, 0),
		gatewayTokens:    make(map[string]string),
		newMonitorClient: NewClient, // Add as struct var so it can be stubbed in tests
	}
}

type registry struct {
	ctx              ttnlog.Interface
	monitorClients   map[string]*Client
	brokerClients    []BrokerClient
	gatewayClients   map[string][]GatewayClient
	gatewayTokens    map[string]string
	newMonitorClient func(ctx ttnlog.Interface, addr string) (*Client, error)
	sync.RWMutex
}

func (r *registry) InitClient(name string, addr string) {
	client, err := r.newMonitorClient(r.ctx, addr)
	if err != nil {
		r.ctx.WithError(err).Warn("Unable to initialize client")
		return
	}

	// Only lock from here, since NewClient blocks when monitor server is down.
	r.Lock()
	defer r.Unlock()

	r.monitorClients[name] = client
	r.brokerClients = append(r.brokerClients, client.BrokerClient)

	// Add gateway client from the new monitor client for every gateway that has been retrieved before.
	for id := range r.gatewayClients {
		gwClient := client.GatewayClient(id)
		gwClient.SetToken(r.gatewayTokens[id])
		r.gatewayClients[id] = append(r.gatewayClients[id], gwClient)
	}
}

func (r *registry) BrokerClients() []BrokerClient {
	r.RLock()
	defer r.RUnlock()

	return r.brokerClients
}

func (r *registry) GatewayClients(id string) []GatewayClient {
	r.RLock()
	clients, ok := r.gatewayClients[id]
	r.RUnlock()

	if ok {
		return clients
	}

	r.Lock()
	defer r.Unlock()

	r.gatewayClients[id] = make([]GatewayClient, 0)
	for _, client := range r.monitorClients {
		gwClient := client.GatewayClient(id)
		gwClient.SetToken(r.gatewayTokens[id])
		r.gatewayClients[id] = append(r.gatewayClients[id], gwClient)
	}

	return r.gatewayClients[id]
}

func (r *registry) SetGatewayToken(id string, token string) {
	r.Lock()
	defer r.Unlock()

	r.gatewayTokens[id] = token

	// Update token on existing clients
	if _, ok := r.gatewayClients[id]; ok {
		for _, gatewayClient := range r.gatewayClients[id] {
			gatewayClient.SetToken(token)
		}
	}
}

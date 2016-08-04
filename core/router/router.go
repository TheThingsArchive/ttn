// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package router

import (
	"sync"
	"time"

	"github.com/TheThingsNetwork/ttn/api"
	pb_broker "github.com/TheThingsNetwork/ttn/api/broker"
	pb_discovery "github.com/TheThingsNetwork/ttn/api/discovery"
	pb_gateway "github.com/TheThingsNetwork/ttn/api/gateway"
	pb "github.com/TheThingsNetwork/ttn/api/router"
	"github.com/TheThingsNetwork/ttn/core"
	"github.com/TheThingsNetwork/ttn/core/discovery"
	"github.com/TheThingsNetwork/ttn/core/router/gateway"
	"github.com/TheThingsNetwork/ttn/core/types"
)

// Router component
type Router interface {
	core.ComponentInterface
	core.ManagementInterface

	// Handle a status message from a gateway
	HandleGatewayStatus(gatewayEUI types.GatewayEUI, status *pb_gateway.Status) error
	// Handle an uplink message from a gateway
	HandleUplink(gatewayEUI types.GatewayEUI, uplink *pb.UplinkMessage) error
	// Handle a downlink message
	HandleDownlink(message *pb_broker.DownlinkMessage) error
	// Subscribe to downlink messages
	SubscribeDownlink(gatewayEUI types.GatewayEUI) (<-chan *pb.DownlinkMessage, error)
	// Unsubscribe from downlink messages
	UnsubscribeDownlink(gatewayEUI types.GatewayEUI) error
	// Handle a device activation
	HandleActivation(gatewayEUI types.GatewayEUI, activation *pb.DeviceActivationRequest) (*pb.DeviceActivationResponse, error)
}

type broker struct {
	client   pb_broker.BrokerClient
	uplink   chan *pb_broker.UplinkMessage
	downlink chan *pb_broker.DownlinkMessage
}

// NewRouter creates a new Router
func NewRouter() Router {
	return &router{
		gateways: make(map[types.GatewayEUI]*gateway.Gateway),
		brokers:  make(map[string]*broker),
	}
}

type router struct {
	*core.Component
	gateways        map[types.GatewayEUI]*gateway.Gateway
	gatewaysLock    sync.RWMutex
	brokerDiscovery discovery.BrokerDiscovery
	brokers         map[string]*broker
	brokersLock     sync.RWMutex
}

func (r *router) tickGateways() {
	r.gatewaysLock.RLock()
	defer r.gatewaysLock.RUnlock()
	for _, gtw := range r.gateways {
		gtw.Utilization.Tick()
	}
}

func (r *router) Init(c *core.Component) error {
	r.Component = c
	err := r.Component.UpdateTokenKey()
	if err != nil {
		return err
	}
	err = r.Component.Announce()
	if err != nil {
		return err
	}
	r.brokerDiscovery = discovery.NewBrokerDiscovery(r.Component)
	r.brokerDiscovery.All() // Update cache
	go func() {
		for range time.Tick(5 * time.Second) {
			r.tickGateways()
		}
	}()
	r.Component.SetStatus(core.StatusHealthy)
	return nil
}

// getGateway gets or creates a Gateway
func (r *router) getGateway(eui types.GatewayEUI) *gateway.Gateway {
	// We're going to be optimistic and guess that the gateway is already active
	r.gatewaysLock.RLock()
	gtw, ok := r.gateways[eui]
	r.gatewaysLock.RUnlock()
	if ok {
		return gtw
	}
	// If it doesn't we still have to lock
	r.gatewaysLock.Lock()
	defer r.gatewaysLock.Unlock()
	if _, ok := r.gateways[eui]; !ok {
		r.gateways[eui] = gateway.NewGateway(r.Ctx, eui)
	}
	return r.gateways[eui]
}

// getBroker gets or creates a broker association and returns the broker
// the first time it also starts a goroutine that receives downlink from the broker
func (r *router) getBroker(brokerAnnouncement *pb_discovery.Announcement) (*broker, error) {
	// We're going to be optimistic and guess that the broker is already active
	r.brokersLock.RLock()
	brk, ok := r.brokers[brokerAnnouncement.Id]
	r.brokersLock.RUnlock()
	if ok {
		return brk, nil
	}
	// If it doesn't we still have to lock
	r.brokersLock.Lock()
	defer r.brokersLock.Unlock()
	if _, ok := r.brokers[brokerAnnouncement.Id]; !ok {

		// Connect to the server
		conn, err := brokerAnnouncement.Dial()
		if err != nil {
			return nil, err
		}
		client := pb_broker.NewBrokerClient(conn)

		brk := &broker{
			client:   client,
			uplink:   make(chan *pb_broker.UplinkMessage),
			downlink: make(chan *pb_broker.DownlinkMessage),
		}

		go func() {
			errors := 0
			for {
				association, err := client.Associate(r.Component.GetContext(""))
				if err != nil {
					errors++
					<-time.After(api.Backoff)
					if errors > 10 {
						break
					}
					continue
				}

				errChan := make(chan error)

				go func() {
					for {
						downlink, err := association.Recv()
						if err != nil {
							errChan <- err
							return
						}
						brk.downlink <- downlink
					}
				}()

			associationLoop:
				for {
					select {
					case err := <-errChan:
						r.Ctx.Errorf("ttn/router: Error in Broker associate: %s", err)
						break associationLoop
					case uplink := <-brk.uplink:
						err := association.Send(uplink)
						if err != nil {
							errChan <- err
						}
					case downlink := <-brk.downlink:
						go r.HandleDownlink(downlink)
					}
				}

			}
			conn.Close()
			r.brokersLock.Lock()
			defer r.brokersLock.Unlock()
			delete(r.brokers, brokerAnnouncement.Id)
		}()

		r.brokers[brokerAnnouncement.Id] = brk
	}
	return r.brokers[brokerAnnouncement.Id], nil
}

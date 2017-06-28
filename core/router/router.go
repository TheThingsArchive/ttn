// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package router

import (
	"sync"
	"time"

	"google.golang.org/grpc"

	pb_broker "github.com/TheThingsNetwork/ttn/api/broker"
	pb_discovery "github.com/TheThingsNetwork/ttn/api/discovery"
	pb_gateway "github.com/TheThingsNetwork/ttn/api/gateway"
	pb_monitor "github.com/TheThingsNetwork/ttn/api/monitor"
	pb "github.com/TheThingsNetwork/ttn/api/router"
	"github.com/TheThingsNetwork/ttn/core/component"
	"github.com/TheThingsNetwork/ttn/core/router/gateway"
)

// Router component
type Router interface {
	component.Interface
	component.ManagementInterface

	// Handle a status message from a gateway
	HandleGatewayStatus(gatewayID string, status *pb_gateway.Status) error
	// Handle an uplink message from a gateway
	HandleUplink(gatewayID string, uplink *pb.UplinkMessage) error
	// Handle a downlink message
	HandleDownlink(message *pb_broker.DownlinkMessage) error
	// Subscribe to downlink messages
	SubscribeDownlink(gatewayID string, subscriptionID string) (<-chan *pb.DownlinkMessage, error)
	// Unsubscribe from downlink messages
	UnsubscribeDownlink(gatewayID string, subscriptionID string) error
	// Handle a device activation
	HandleActivation(gatewayID string, activation *pb.DeviceActivationRequest) (*pb.DeviceActivationResponse, error)

	getGateway(gatewayID string) *gateway.Gateway
}

type broker struct {
	conn        *grpc.ClientConn
	association pb_broker.RouterStream
	client      pb_broker.BrokerClient
	uplink      chan *pb_broker.UplinkMessage
	downlink    chan *pb_broker.DownlinkMessage
}

// NewRouter creates a new Router
func NewRouter() Router {
	return &router{
		gateways: make(map[string]*gateway.Gateway),
		brokers:  make(map[string]*broker),
	}
}

type router struct {
	*component.Component
	gateways      map[string]*gateway.Gateway
	gatewaysLock  sync.RWMutex
	brokers       map[string]*broker
	brokersLock   sync.RWMutex
	status        *status
	monitorStream pb_monitor.GenericStream
}

func (r *router) tickGateways() {
	r.gatewaysLock.RLock()
	defer r.gatewaysLock.RUnlock()
	for _, gtw := range r.gateways {
		gtw.Utilization.Tick()
	}
}

func (r *router) Init(c *component.Component) error {
	r.Component = c
	r.InitStatus()
	err := r.Component.UpdateTokenKey()
	if err != nil {
		return err
	}
	err = r.Component.Announce()
	if err != nil {
		return err
	}
	r.Discovery.GetAll("broker") // Update cache

	go func() {
		for range time.Tick(5 * time.Second) {
			r.tickGateways()
		}
	}()
	r.Component.SetStatus(component.StatusHealthy)
	if r.Component.Monitor != nil {
		r.monitorStream = r.Component.Monitor.NewRouterStreams(r.Identity.Id, r.AccessToken)
		go r.Component.Monitor.TickStatus(func() {
			r.monitorStream.Send(r.GetStatus())
		})
	}
	return nil
}

func (r *router) Shutdown() {
	r.brokersLock.Lock()
	defer r.brokersLock.Unlock()
	for _, broker := range r.brokers {
		broker.association.Close()
		broker.conn.Close()
	}
}

// getGateway gets or creates a Gateway
func (r *router) getGateway(id string) *gateway.Gateway {
	// We're going to be optimistic and guess that the gateway is already active
	r.gatewaysLock.RLock()
	gtw, ok := r.gateways[id]
	r.gatewaysLock.RUnlock()
	if ok {
		return gtw
	}
	// If it doesn't we still have to lock
	r.gatewaysLock.Lock()
	defer r.gatewaysLock.Unlock()

	gtw, ok = r.gateways[id]
	if !ok {
		gtw = gateway.NewGateway(r.Ctx, id)
		gtw.Monitor = r.Component.Monitor

		r.gateways[id] = gtw
	}

	return gtw
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
		var err error

		brk := &broker{
			uplink: make(chan *pb_broker.UplinkMessage),
		}

		// Connect to the server
		brk.conn, err = brokerAnnouncement.Dial(r.Pool)
		if err != nil {
			return nil, err
		}

		// Set up the non-streaming client
		brk.client = pb_broker.NewBrokerClient(brk.conn)

		// Set up the streaming client
		config := pb_broker.DefaultClientConfig
		config.BackgroundContext = r.Component.Context
		cli := pb_broker.NewClient(config)
		cli.AddServer(brokerAnnouncement.Id, brk.conn)
		brk.association = cli.NewRouterStreams(r.Identity.Id, "")

		go func() {
			for {
				select {
				case message := <-brk.uplink:
					brk.association.Uplink(message)
				case message, ok := <-brk.association.Downlink():
					if ok {
						go r.HandleDownlink(message)
					}
				}
			}
		}()

		r.brokers[brokerAnnouncement.Id] = brk
	}
	return r.brokers[brokerAnnouncement.Id], nil
}

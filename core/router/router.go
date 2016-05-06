package router

import (
	"sync"

	"golang.org/x/net/context"

	"google.golang.org/grpc"

	"github.com/TheThingsNetwork/ttn/api"
	pb_broker "github.com/TheThingsNetwork/ttn/api/broker"
	pb_discovery "github.com/TheThingsNetwork/ttn/api/discovery"
	pb_gateway "github.com/TheThingsNetwork/ttn/api/gateway"
	pb "github.com/TheThingsNetwork/ttn/api/router"
	"github.com/TheThingsNetwork/ttn/core/discovery"
	"github.com/TheThingsNetwork/ttn/core/router/gateway"
	"github.com/TheThingsNetwork/ttn/core/types"
)

// Router component
type Router interface {
	// Handle a status message from a gateway
	HandleGatewayStatus(gatewayEUI types.GatewayEUI, status *pb_gateway.StatusMessage) error
	// Handle an uplink message from a gateway
	HandleUplink(gatewayEUI types.GatewayEUI, uplink *pb.UplinkMessage) error
	// Handle a downlink message
	HandleDownlink(message *pb.DownlinkMessage) error
	// Subscribe to downlink messages
	SubscribeDownlink(gatewayEUI types.GatewayEUI) (chan pb.DownlinkMessage, error)
	// Unsubscribe from downlink messages
	UnsubscribeDownlink(gatewayEUI types.GatewayEUI) error
	// Handle a device activation
	HandleActivation(gatewayEUI types.GatewayEUI, activation *pb.DeviceActivationRequest) (*pb.DeviceActivationResponse, error)
}

type router struct {
	identity        *pb_discovery.Announcement
	gateways        map[types.GatewayEUI]*gateway.Gateway
	gatewaysLock    sync.RWMutex
	brokerDiscovery discovery.BrokerDiscovery
	brokers         map[string]pb_broker.Broker_AssociateClient
	brokersLock     sync.RWMutex
}

// getGateway gets or creates a Gateway
func (r *router) getGateway(eui types.GatewayEUI) *gateway.Gateway {
	r.gatewaysLock.Lock()
	defer r.gatewaysLock.Unlock()
	if _, ok := r.gateways[eui]; !ok {
		r.gateways[eui] = gateway.NewGateway(eui)
	}
	return r.gateways[eui]
}

// getBroker gets or creates a broker association
func (r *router) getBroker(broker *pb_discovery.Announcement) (pb_broker.Broker_AssociateClient, error) {
	r.gatewaysLock.Lock()
	defer r.gatewaysLock.Unlock()
	if _, ok := r.brokers[broker.NetAddress]; !ok {
		// Connect to the server
		conn, err := grpc.Dial(broker.NetAddress, api.DialOptions...)
		if err != nil {
			return nil, err
		}
		client := pb_broker.NewBrokerClient(conn)
		association, err := client.Associate(context.Background())
		if err != nil {
			return nil, err
		}
		r.brokers[broker.NetAddress] = association
	}
	return r.brokers[broker.NetAddress], nil
}

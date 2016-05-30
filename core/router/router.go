package router

import (
	"io"
	"sync"

	"golang.org/x/net/context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

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
	client      pb_broker.BrokerClient
	association pb_broker.Broker_AssociateClient
}

type router struct {
	identity        *pb_discovery.Announcement
	gateways        map[types.GatewayEUI]*gateway.Gateway
	gatewaysLock    sync.RWMutex
	brokerDiscovery discovery.BrokerDiscovery
	brokers         map[string]*broker
	brokersLock     sync.RWMutex
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
		r.gateways[eui] = gateway.NewGateway(eui)
	}
	return r.gateways[eui]
}

// getBroker gets or creates a broker association and returns the broker
// the first time it also starts a goroutine that receives downlink from the broker
func (r *router) getBroker(req *pb_discovery.Announcement) (*broker, error) {
	// We're going to be optimistic and guess that the broker is already active
	r.brokersLock.RLock()
	brk, ok := r.brokers[req.NetAddress]
	r.brokersLock.RUnlock()
	if ok {
		return brk, nil
	}
	// If it doesn't we still have to lock
	r.brokersLock.Lock()
	defer r.brokersLock.Unlock()
	if _, ok := r.brokers[req.NetAddress]; !ok {
		// Connect to the server
		conn, err := grpc.Dial(req.NetAddress, api.DialOptions...)
		if err != nil {
			return nil, err
		}
		client := pb_broker.NewBrokerClient(conn)

		association, err := client.Associate(r.getContext())
		if err != nil {
			return nil, err
		}
		// Start a goroutine that receives and processes downlink
		go func() {
			for {
				downlink, err := association.Recv()
				if err == io.EOF {
					association.CloseSend()
					break
				}
				if err != nil {
					break
				}
				go r.HandleDownlink(downlink)
			}
			// When the loop is broken: close connection and unregister broker.
			conn.Close()
			r.brokersLock.Lock()
			defer r.brokersLock.Unlock()
			delete(r.brokers, req.NetAddress)
		}()
		r.brokers[req.NetAddress] = &broker{
			client:      client,
			association: association,
		}
	}
	return r.brokers[req.NetAddress], nil
}

func (r *router) getContext() context.Context {
	var id, token string
	if r.identity != nil {
		id = r.identity.Id
		token = r.identity.Token
	}
	md := metadata.Pairs(
		"token", token,
		"id", id,
	)
	ctx := metadata.NewContext(context.Background(), md)
	return ctx
}

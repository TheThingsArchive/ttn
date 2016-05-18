package router

import (
	pb_broker "github.com/TheThingsNetwork/ttn/api/broker"
	pb "github.com/TheThingsNetwork/ttn/api/router"
	"github.com/TheThingsNetwork/ttn/core/types"
)

func (r *router) HandleUplink(gatewayEUI types.GatewayEUI, uplink *pb.UplinkMessage) error {
	gateway := r.getGateway(gatewayEUI)

	gateway.Utilization.AddRx(uplink)

	downlinkOptions := buildDownlinkOptions(uplink, false, gateway)

	// Find Broker
	devAddr := types.DevAddr{1, 2, 3, 4}
	brokers, err := r.brokerDiscovery.Discover(devAddr)
	if err != nil {
		return err
	}

	// Forward to all brokers
	for _, broker := range brokers {
		broker, err := r.getBroker(broker)
		if err != nil {
			continue
		}
		broker.association.Send(&pb_broker.UplinkMessage{
			Payload:          uplink.Payload,
			ProtocolMetadata: uplink.ProtocolMetadata,
			GatewayMetadata:  uplink.GatewayMetadata,
			DownlinkOptions:  downlinkOptions,
		})
	}

	return nil
}

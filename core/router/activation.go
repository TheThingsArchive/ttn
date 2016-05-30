package router

import (
	"errors"
	"sync"

	pb_broker "github.com/TheThingsNetwork/ttn/api/broker"
	pb "github.com/TheThingsNetwork/ttn/api/router"
	"github.com/TheThingsNetwork/ttn/core/types"
	"golang.org/x/net/context"
)

func (r *router) HandleActivation(gatewayEUI types.GatewayEUI, activation *pb.DeviceActivationRequest) (*pb.DeviceActivationResponse, error) {

	gateway := r.getGateway(gatewayEUI)

	uplink := &pb.UplinkMessage{
		Payload:          activation.Payload,
		ProtocolMetadata: activation.ProtocolMetadata,
		GatewayMetadata:  activation.GatewayMetadata,
	}

	// Only for LoRaWAN
	gateway.Utilization.AddRx(uplink)

	downlinkOptions := buildDownlinkOptions(uplink, true, gateway)

	// Find Broker
	brokers, err := r.brokerDiscovery.All()
	if err != nil {
		return nil, err
	}

	// Forward to all brokers and collect responses
	var wg sync.WaitGroup
	responses := make(chan *pb_broker.DeviceActivationResponse, len(brokers))
	for _, broker := range brokers {
		broker, err := r.getBroker(broker)
		if err != nil {
			continue
		}

		// Do async request
		wg.Add(1)
		go func() {
			res, err := broker.client.Activate(context.Background(), &pb_broker.DeviceActivationRequest{
				Payload:          activation.Payload,
				DevEui:           activation.DevEui,
				AppEui:           activation.AppEui,
				ProtocolMetadata: activation.ProtocolMetadata,
				GatewayMetadata:  activation.GatewayMetadata,
				DownlinkOptions:  downlinkOptions,
			})
			if err == nil && res != nil {
				responses <- res
			}
			wg.Done()
		}()
	}

	// Make sure to close channel when all requests are done
	go func() {
		wg.Wait()
		close(responses)
	}()

	var gotFirst bool
	for res := range responses {
		if gotFirst {
			// warn for duplicate responses
		} else {
			gotFirst = true
			downlink := &pb_broker.DownlinkMessage{
				Payload:        res.Payload,
				DownlinkOption: res.DownlinkOption,
			}
			err := r.HandleDownlink(downlink)
			if err != nil {
				gotFirst = false // try again
			}
		}
	}

	// Activation not accepted by any broker
	if !gotFirst {
		return nil, errors.New("ttn/router: Activation not accepted at this Gateway")
	}

	// Activation accepted by (at least one) broker
	return &pb.DeviceActivationResponse{}, nil
}

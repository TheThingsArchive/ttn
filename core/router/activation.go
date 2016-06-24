// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package router

import (
	"errors"
	"sync"
	"time"

	pb_broker "github.com/TheThingsNetwork/ttn/api/broker"
	pb_protocol "github.com/TheThingsNetwork/ttn/api/protocol"
	pb_lorawan "github.com/TheThingsNetwork/ttn/api/protocol/lorawan"
	pb "github.com/TheThingsNetwork/ttn/api/router"
	"github.com/TheThingsNetwork/ttn/core/types"
	"github.com/apex/log"
)

func (r *router) HandleActivation(gatewayEUI types.GatewayEUI, activation *pb.DeviceActivationRequest) (*pb.DeviceActivationResponse, error) {
	ctx := r.Ctx.WithFields(log.Fields{
		"GatewayEUI": gatewayEUI,
		"AppEUI":     *activation.AppEui,
		"DevEUI":     *activation.DevEui,
	})
	var err error
	defer func() {
		if err != nil {
			ctx.WithError(err).Warn("Could not handle activation")
		}
	}()

	gateway := r.getGateway(gatewayEUI)
	gateway.LastSeen = time.Now()

	uplink := &pb.UplinkMessage{
		Payload:          activation.Payload,
		ProtocolMetadata: activation.ProtocolMetadata,
		GatewayMetadata:  activation.GatewayMetadata,
	}

	// Only for LoRaWAN
	gateway.Schedule.Sync(uplink.GatewayMetadata.Timestamp)
	gateway.Utilization.AddRx(uplink)
	downlinkOptions := r.buildDownlinkOptions(uplink, true, gateway)

	// Find Broker
	brokers, err := r.brokerDiscovery.All()
	if err != nil {
		return nil, err
	}

	// Prepare request
	request := &pb_broker.DeviceActivationRequest{
		Payload:          activation.Payload,
		DevEui:           activation.DevEui,
		AppEui:           activation.AppEui,
		ProtocolMetadata: activation.ProtocolMetadata,
		GatewayMetadata:  activation.GatewayMetadata,
		ActivationMetadata: &pb_protocol.ActivationMetadata{
			Protocol: &pb_protocol.ActivationMetadata_Lorawan{
				Lorawan: &pb_lorawan.ActivationMetadata{
					AppEui: activation.AppEui,
					DevEui: activation.DevEui,
				},
			},
		},
		DownlinkOptions: downlinkOptions,
	}

	// Prepare LoRaWAN activation
	status, err := gateway.Status.Get()
	if err != nil {
		return nil, err
	}
	region := status.Region
	if region == "" {
		region = guessRegion(uplink.GatewayMetadata.Frequency)
	}
	band, err := getBand(region)
	if err != nil {
		return nil, err
	}
	lorawan := request.ActivationMetadata.GetLorawan()
	lorawan.Rx1DrOffset = 0
	lorawan.Rx2Dr = uint32(band.RX2DataRate)
	lorawan.RxDelay = uint32(band.ReceiveDelay1.Seconds())
	switch region {
	case "EU_863_870":
		lorawan.CfList = []uint64{867100000, 867300000, 867500000, 867700000, 867900000}
	}

	ctx = ctx.WithField("NumBrokers", len(brokers))
	ctx.Debug("Forward Activation")

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
			res, err := broker.client.Activate(r.Component.GetContext(), request)
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
			ctx.Warn("Duplicate Activation Response")
		} else {
			gotFirst = true
			downlink := &pb_broker.DownlinkMessage{
				Payload:        res.Payload,
				DownlinkOption: res.DownlinkOption,
			}
			err := r.HandleDownlink(downlink)
			if err != nil {
				ctx.Warn("Could not send downlink for Activation")
				gotFirst = false // try again
			}
		}
	}

	// Activation not accepted by any broker
	if !gotFirst {
		ctx.Debug("Activation not accepted at this gateway")
		return nil, errors.New("ttn/router: Activation not accepted at this Gateway")
	}

	// Activation accepted by (at least one) broker
	ctx.Debug("Activation accepted")
	return &pb.DeviceActivationResponse{}, nil
}

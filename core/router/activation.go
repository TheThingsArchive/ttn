// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package router

import (
	"context"
	"fmt"
	"sync"
	"time"

	pb_broker "github.com/TheThingsNetwork/ttn/api/broker"
	"github.com/TheThingsNetwork/ttn/api/fields"
	pb_protocol "github.com/TheThingsNetwork/ttn/api/protocol"
	pb_lorawan "github.com/TheThingsNetwork/ttn/api/protocol/lorawan"
	pb "github.com/TheThingsNetwork/ttn/api/router"
	"github.com/TheThingsNetwork/ttn/api/trace"
	"github.com/TheThingsNetwork/ttn/utils/errors"
)

func (r *router) HandleActivation(gatewayID string, activation *pb.DeviceActivationRequest) (res *pb.DeviceActivationResponse, err error) {
	ctx := r.Ctx.WithField("GatewayID", gatewayID).WithFields(fields.Get(activation))
	start := time.Now()
	gateway := r.getGateway(gatewayID)
	var forwarded bool
	defer func() {
		if err != nil {
			activation.Trace = activation.Trace.WithEvent(trace.DropEvent, "reason", err)
			if forwarded {
				ctx.WithError(err).Debug("Did not handle activation")
			} else {
				ctx.WithError(err).Warn("Could not handle activation")
				gateway.SendToMonitor(activation)
			}
		} else {
			ctx.WithField("Duration", time.Now().Sub(start)).Info("Handled activation")
		}
	}()
	r.status.activations.Mark(1)
	activation.Trace = activation.Trace.WithEvent(trace.ReceiveEvent, "gateway", gatewayID)

	if err := activation.UnmarshalPayload(); err != nil {
		return nil, err
	}

	uplink := &pb.UplinkMessage{
		Payload:          activation.Payload,
		ProtocolMetadata: activation.ProtocolMetadata,
		GatewayMetadata:  activation.GatewayMetadata,
		Trace:            activation.Trace,
	}
	if err = gateway.HandleUplink(uplink); err != nil {
		return nil, err
	}

	downlinkOptions, err := gateway.GetDownlinkOptions(uplink)
	if err != nil {
		uplink.Trace = uplink.Trace.WithEvent(trace.WarnEvent, trace.ErrorField, fmt.Sprintf("could not build downlink options: %s", err))
	} else {
		uplink.Trace = uplink.Trace.WithEvent(trace.BuildDownlinkEvent, "options", len(downlinkOptions))
		ctx = ctx.WithField("DownlinkOptions", len(downlinkOptions))
	}
	if r.Component != nil && r.Component.Identity != nil {
		for _, opt := range downlinkOptions {
			opt.Identifier = fmt.Sprintf("%s:%s", r.Component.Identity.Id, opt.Identifier)
		}
	}

	activation.Trace = uplink.Trace

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
					AppEui:        activation.AppEui,
					DevEui:        activation.DevEui,
					FrequencyPlan: activation.ProtocolMetadata.GetLorawan().GetFrequencyPlan(),
				},
			},
		},
		DownlinkOptions: downlinkOptions,
		Trace:           activation.Trace,
	}

	// Prepare LoRaWAN activation
	frequencyPlan := gateway.FrequencyPlan()
	if frequencyPlan == nil {
		return nil, errors.New("gateway: frequency plan unknown")
	}

	lorawan := request.ActivationMetadata.GetLorawan()
	lorawan.Rx1DrOffset = 0
	lorawan.Rx2Dr = uint32(frequencyPlan.RX2DataRate)
	if len(frequencyPlan.RX1Delays) > 0 {
		lorawan.RxDelay = uint32(frequencyPlan.RX1Delays[0].Seconds())
	} else {
		lorawan.RxDelay = uint32(frequencyPlan.ReceiveDelay1.Seconds())
	}
	if frequencyPlan.CFList != nil {
		lorawan.CfList = new(pb_lorawan.CFList)
		for _, freq := range frequencyPlan.CFList {
			lorawan.CfList.Freq = append(lorawan.CfList.Freq, freq)
		}
	}

	// Find Brokers
	brokers, err := r.Discovery.GetAll("broker")
	if err != nil {
		return nil, err
	}
	ctx = ctx.WithField("NumBrokers", len(brokers))
	request.Trace = request.Trace.WithEvent(trace.ForwardEvent, "brokers", len(brokers))

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
			ctx, cancel := context.WithTimeout(r.Component.GetContext(""), 5*time.Second)
			defer cancel()
			res, err := broker.client.Activate(ctx, request)
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
				Message:        res.Message,
				DownlinkOption: res.DownlinkOption,
				Trace:          res.Trace,
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
		return nil, errors.New("Activation not accepted at this Gateway")
	}

	// Activation accepted by (at least one) broker
	ctx.Debug("Activation accepted")
	return &pb.DeviceActivationResponse{}, nil
}

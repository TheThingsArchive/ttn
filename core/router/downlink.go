// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package router

import (
	"fmt"
	"math"
	"strings"
	"time"

	pb_broker "github.com/TheThingsNetwork/api/broker"
	pb_gateway "github.com/TheThingsNetwork/api/gateway"
	"github.com/TheThingsNetwork/api/logfields"
	pb_protocol "github.com/TheThingsNetwork/api/protocol"
	pb_lorawan "github.com/TheThingsNetwork/api/protocol/lorawan"
	pb "github.com/TheThingsNetwork/api/router"
	"github.com/TheThingsNetwork/api/trace"
	ttnlog "github.com/TheThingsNetwork/go-utils/log"
	"github.com/TheThingsNetwork/ttn/core/band"
	"github.com/TheThingsNetwork/ttn/core/router/gateway"
	"github.com/TheThingsNetwork/ttn/core/types"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	"github.com/TheThingsNetwork/ttn/utils/toa"
)

func (r *router) SubscribeDownlink(gatewayID string, subscriptionID string) (<-chan *pb.DownlinkMessage, error) {
	ctx := r.Ctx.WithFields(ttnlog.Fields{
		"GatewayID": gatewayID,
	})

	gateway := r.getGateway(gatewayID)
	if fromSchedule := gateway.Schedule.Subscribe(subscriptionID); fromSchedule != nil {
		if token := gateway.Token(); gatewayID != "" && token != "" {
			r.Discovery.AddGatewayID(gatewayID, token)
		}
		toGateway := make(chan *pb.DownlinkMessage)
		go func() {
			ctx.Debug("Activate downlink")
			for message := range fromSchedule {
				ctx.WithFields(logfields.ForMessage(message)).Debug("Send downlink")
				toGateway <- message
				if gateway.MonitorStream != nil {
					clone := *message // There can be multiple subscribers
					clone.Trace = clone.Trace.WithEvent(trace.SendEvent)
					gateway.MonitorStream.Send(&clone)
				}
			}
			ctx.Debug("Deactivate downlink")
			close(toGateway)
		}()
		return toGateway, nil
	}
	return nil, errors.NewErrInternal(fmt.Sprintf("Already subscribed to downlink for %s", gatewayID))
}

func (r *router) UnsubscribeDownlink(gatewayID string, subscriptionID string) error {
	gateway := r.getGateway(gatewayID)
	if token := gateway.Token(); gatewayID != "" && token != "" {
		r.Discovery.RemoveGatewayID(gatewayID, token)
	}
	gateway.Schedule.Stop(subscriptionID)
	return nil
}

func (r *router) HandleDownlink(downlink *pb_broker.DownlinkMessage) (err error) {
	var gateway *gateway.Gateway

	r.RegisterReceived(downlink)
	defer func() {
		if err != nil {
			downlink.Trace = downlink.Trace.WithEvent(trace.DropEvent, "reason", err)
			if gateway != nil && gateway.MonitorStream != nil {
				gateway.MonitorStream.Send(downlink)
			}
		} else {
			r.RegisterHandled(downlink)
		}
	}()
	r.status.downlink.Mark(1)

	downlink.Trace = downlink.Trace.WithEvent(trace.ReceiveEvent)

	option := downlink.DownlinkOption

	downlinkMessage := &pb.DownlinkMessage{
		Payload:               downlink.Payload,
		ProtocolConfiguration: option.ProtocolConfiguration,
		GatewayConfiguration:  option.GatewayConfiguration,
		Trace:                 downlink.Trace,
	}

	identifier := option.Identifier
	if r.Component != nil && r.Component.Identity != nil {
		identifier = strings.TrimPrefix(option.Identifier, fmt.Sprintf("%s:", r.Component.Identity.ID))
	}

	gateway = r.getGateway(downlink.DownlinkOption.GatewayID)
	return gateway.HandleDownlink(identifier, downlinkMessage)
}

// buildDownlinkOption builds a DownlinkOption with default values
func (r *router) buildDownlinkOption(gatewayID string, band band.FrequencyPlan) *pb_broker.DownlinkOption {
	dataRate, _ := types.ConvertDataRate(band.DataRates[band.RX2DataRate])
	return &pb_broker.DownlinkOption{
		GatewayID: gatewayID,
		ProtocolConfiguration: pb_protocol.TxConfiguration{Protocol: &pb_protocol.TxConfiguration_LoRaWAN{LoRaWAN: &pb_lorawan.TxConfiguration{
			Modulation: pb_lorawan.Modulation_LORA,
			DataRate:   dataRate.String(),
			CodingRate: "4/5",
		}}},
		GatewayConfiguration: pb_gateway.TxConfiguration{
			RfChain:               0,
			PolarizationInversion: true,
			Frequency:             uint64(band.RX2Frequency),
			Power:                 int32(band.DefaultTXPower),
		},
	}
}

func (r *router) buildDownlinkOptions(uplink *pb.UplinkMessage, isActivation bool, gateway *gateway.Gateway) (downlinkOptions []*pb_broker.DownlinkOption) {
	var options []*pb_broker.DownlinkOption

	gatewayStatus, _ := gateway.Status.Get() // This just returns empty if non-existing

	lorawanMetadata := uplink.ProtocolMetadata.GetLoRaWAN()
	if lorawanMetadata == nil {
		return // We can't handle any other protocols than LoRaWAN yet
	}

	frequencyPlan := gatewayStatus.FrequencyPlan
	if frequencyPlan == "" {
		frequencyPlan = band.Guess(uplink.GatewayMetadata.Frequency)
	}
	band, err := band.Get(frequencyPlan)
	if err != nil {
		return // We can't handle this frequency plan
	}
	if frequencyPlan == "EU_863_870" && isActivation {
		band.RX2DataRate = 0
	}

	dataRate, err := lorawanMetadata.GetLoRaWANDataRate()
	if err != nil {
		return
	}

	// Configuration for RX2
	buildRX2 := func() (*pb_broker.DownlinkOption, error) {
		option := r.buildDownlinkOption(gateway.ID, band)
		if frequencyPlan == "EU_863_870" {
			option.GatewayConfiguration.Power = 27 // The EU RX2 frequency allows up to 27dBm
		}
		if isActivation {
			option.GatewayConfiguration.Timestamp = uplink.GatewayMetadata.Timestamp + uint32(band.JoinAcceptDelay2/1000)
		} else {
			option.GatewayConfiguration.Timestamp = uplink.GatewayMetadata.Timestamp + uint32(band.ReceiveDelay2/1000)
		}
		option.ProtocolConfiguration.GetLoRaWAN().CodingRate = lorawanMetadata.CodingRate
		return option, nil
	}

	if option, err := buildRX2(); err == nil {
		options = append(options, option)
	}

	// Configuration for RX1
	buildRX1 := func() (*pb_broker.DownlinkOption, error) {
		option := r.buildDownlinkOption(gateway.ID, band)
		if isActivation {
			option.GatewayConfiguration.Timestamp = uplink.GatewayMetadata.Timestamp + uint32(band.JoinAcceptDelay1/1000)
		} else {
			option.GatewayConfiguration.Timestamp = uplink.GatewayMetadata.Timestamp + uint32(band.ReceiveDelay1/1000)
		}
		option.ProtocolConfiguration.GetLoRaWAN().CodingRate = lorawanMetadata.CodingRate

		freq, err := band.GetRX1Frequency(int(uplink.GatewayMetadata.Frequency))
		if err != nil {
			return nil, err
		}
		option.GatewayConfiguration.Frequency = uint64(freq)

		upDR, err := band.GetDataRate(dataRate)
		if err != nil {
			return nil, err
		}
		downDR, err := band.GetRX1DataRate(upDR, 0)
		if err != nil {
			return nil, err
		}

		if err := option.ProtocolConfiguration.GetLoRaWAN().SetDataRate(band.DataRates[downDR]); err != nil {
			return nil, err
		}
		option.GatewayConfiguration.FrequencyDeviation = uint32(option.ProtocolConfiguration.GetLoRaWAN().BitRate / 2)

		return option, nil
	}

	if option, err := buildRX1(); err == nil {
		options = append(options, option)
	}

	computeDownlinkScores(gateway, uplink, options)

	for _, option := range options {
		// Add router ID to downlink option
		if r.Component != nil && r.Component.Identity != nil {
			option.Identifier = fmt.Sprintf("%s:%s", r.Component.Identity.ID, option.Identifier)
		}

		// Filter all illegal options
		if option.Score < 1000 {
			downlinkOptions = append(downlinkOptions, option)
		}
	}

	return
}

// Calculating the score for each downlink option; lower is better, 0 is best
// If a score is over 1000, it may should not be used as feasible option.
// TODO: The weights of these parameters should be optimized. I'm sure someone
// can do some computer simulations to find the right values.
func computeDownlinkScores(gateway *gateway.Gateway, uplink *pb.UplinkMessage, options []*pb_broker.DownlinkOption) {
	gatewayStatus, _ := gateway.Status.Get() // This just returns empty if non-existing

	frequencyPlan := gatewayStatus.FrequencyPlan
	if frequencyPlan == "" {
		frequencyPlan = band.Guess(uplink.GatewayMetadata.Frequency)
	}

	gatewayRx, _ := gateway.Utilization.Get()
	for _, option := range options {

		// Invalid if no LoRaWAN
		conf := option.GetProtocolConfiguration()
		lorawan := conf.GetLoRaWAN()
		if lorawan == nil {
			option.Score = 1000
			continue
		}

		var time time.Duration

		if lorawan.Modulation == pb_lorawan.Modulation_LORA {
			// Calculate max ToA
			time, _ = toa.ComputeLoRa(
				51+13, // Max MACPayload plus LoRaWAN header, TODO: What is the length we should use?
				lorawan.DataRate,
				lorawan.CodingRate,
			)
		}

		if lorawan.Modulation == pb_lorawan.Modulation_FSK {
			// Calculate max ToA
			time, _ = toa.ComputeFSK(
				51+13, // Max MACPayload plus LoRaWAN header, TODO: What is the length we should use?
				int(lorawan.BitRate),
			)
		}

		// Invalid if time is zero
		if time == 0 {
			option.Score = 1000
			continue
		}

		timeScore := math.Min(time.Seconds()*5, 10) // 2 seconds will be 10 (max)

		signalScore := 0.0 // Between 0 and 20 (lower is better)
		{
			// Prefer high SNR
			if uplink.GatewayMetadata.SNR < 5 {
				signalScore += 10
			}
			// Prefer good RSSI
			signalScore += math.Min(float64(uplink.GatewayMetadata.RSSI*-0.1), 10)
		}

		utilizationScore := 0.0 // Between 0 and 40 (lower is better) will be over 100 if forbidden
		{
			// Avoid gateways that do more Rx
			utilizationScore += math.Min(gatewayRx*50, 20) / 2 // 40% utilization = 10 (max)

			// Avoid busy channels
			freq := option.GatewayConfiguration.Frequency
			channelRx, channelTx := gateway.Utilization.GetChannel(freq)
			utilizationScore += math.Min((channelTx+channelRx)*200, 20) / 2 // 10% utilization = 10 (max)

			// European Duty Cycle
			if frequencyPlan == "EU_863_870" {
				var duty float64
				switch {
				case freq >= 863000000 && freq < 868000000:
					duty = 0.01 // g 863.0 – 868.0 MHz 1%
				case freq >= 868000000 && freq < 868600000:
					duty = 0.01 // g1 868.0 – 868.6 MHz 1%
				case freq >= 868700000 && freq < 869200000:
					duty = 0.001 // g2 868.7 – 869.2 MHz 0.1%
				case freq >= 869400000 && freq < 869650000:
					duty = 0.1 // g3 869.4 – 869.65 MHz 10%
				case freq >= 869700000 && freq < 870000000:
					duty = 0.01 // g4 869.7 – 870.0 MHz 1%
				default:
					utilizationScore += 100 // Transmissions on this frequency are forbidden
				}
				if channelTx > duty {
					utilizationScore += 100 // Transmissions on this frequency are forbidden
				}
				if duty > 0 {
					utilizationScore += math.Min(time.Seconds()/duty/100, 20) // Impact on duty-cycle (in order to prefer RX2 for SF9BW125)
				}
			}
		}

		scheduleScore := 0.0 // Between 0 and 30 (lower is better) will be over 100 if forbidden
		{
			id, conflicts := gateway.Schedule.GetOption(option.GatewayConfiguration.Timestamp, uint32(time/1000))
			option.Identifier = id
			if conflicts >= 100 {
				scheduleScore += 100
			} else {
				scheduleScore += math.Min(float64(conflicts*10), 30) // max 30
			}
		}

		option.Score = uint32((timeScore + signalScore + utilizationScore + scheduleScore) * 10)
	}
}

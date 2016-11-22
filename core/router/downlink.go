// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package router

import (
	"fmt"
	"math"
	"strings"
	"time"

	pb_broker "github.com/TheThingsNetwork/ttn/api/broker"
	pb_gateway "github.com/TheThingsNetwork/ttn/api/gateway"
	pb_protocol "github.com/TheThingsNetwork/ttn/api/protocol"
	pb_lorawan "github.com/TheThingsNetwork/ttn/api/protocol/lorawan"
	pb "github.com/TheThingsNetwork/ttn/api/router"
	"github.com/TheThingsNetwork/ttn/core/router/gateway"
	"github.com/TheThingsNetwork/ttn/core/types"
	"github.com/TheThingsNetwork/ttn/utils/errors"
	"github.com/TheThingsNetwork/ttn/utils/toa"
	"github.com/apex/log"
	lora "github.com/brocaar/lorawan/band"
)

func (r *router) SubscribeDownlink(gatewayID string) (<-chan *pb.DownlinkMessage, error) {
	ctx := r.Ctx.WithFields(log.Fields{
		"GatewayID": gatewayID,
	})

	gateway := r.getGateway(gatewayID)
	if fromSchedule := gateway.Schedule.Subscribe(); fromSchedule != nil {
		toGateway := make(chan *pb.DownlinkMessage)
		go func() {
			ctx.Debug("Activate downlink")
			for message := range fromSchedule {
				gateway.Utilization.AddTx(message)
				ctx.Debug("Send downlink")
				toGateway <- message
			}
			ctx.Debug("Deactivate downlink")
			close(toGateway)
		}()
		return toGateway, nil
	}
	return nil, errors.NewErrInternal(fmt.Sprintf("Already subscribed to downlink for %s", gatewayID))
}

func (r *router) UnsubscribeDownlink(gatewayID string) error {
	r.getGateway(gatewayID).Schedule.Stop()
	return nil
}

func (r *router) HandleDownlink(downlink *pb_broker.DownlinkMessage) error {
	option := downlink.DownlinkOption

	downlinkMessage := &pb.DownlinkMessage{
		Payload:               downlink.Payload,
		ProtocolConfiguration: option.ProtocolConfig,
		GatewayConfiguration:  option.GatewayConfig,
	}

	identifier := option.Identifier
	if r.Component != nil && r.Component.Identity != nil {
		identifier = strings.TrimPrefix(option.Identifier, fmt.Sprintf("%s:", r.Component.Identity.Id))
	}

	return r.getGateway(downlink.DownlinkOption.GatewayId).HandleDownlink(identifier, downlinkMessage)
}

func guessRegion(frequency uint64) string {
	switch {
	case frequency >= 863000000 && frequency <= 870000000:
		return pb_lorawan.Region_EU_863_870.String()
	case frequency >= 902300000 && frequency <= 914900000:
		return pb_lorawan.Region_US_902_928.String()
	case frequency >= 779500000 && frequency <= 786500000:
		return pb_lorawan.Region_CN_779_787.String()
	case frequency >= 433175000 && frequency <= 434665000:
		return pb_lorawan.Region_EU_433.String()
	case frequency == 923200000 || frequency == 923400000:
		return pb_lorawan.Region_AS_923.String()
	case frequency >= 920900000 || frequency == 923300000:
		return pb_lorawan.Region_KR_920_923.String()
	case frequency >= 915200000 && frequency <= 927800000:
		return pb_lorawan.Region_AU_915_928.String()
	case frequency >= 470300000 && frequency <= 489300000:
		return pb_lorawan.Region_CN_470_510.String()
	}
	return ""
}

func getBand(region string) (band *lora.Band, err error) {
	var b lora.Band

	switch region {
	case pb_lorawan.Region_EU_863_870.String():
		b, err = lora.GetConfig(lora.EU_863_870)
	case pb_lorawan.Region_US_902_928.String():
		b, err = lora.GetConfig(lora.US_902_928)
	case pb_lorawan.Region_CN_779_787.String():
		err = errors.NewErrInternal("China 779-787 MHz band not supported")
	case pb_lorawan.Region_EU_433.String():
		err = errors.NewErrInternal("Europe 433 MHz band not supported")
	case pb_lorawan.Region_AU_915_928.String():
		b, err = lora.GetConfig(lora.AU_915_928)
	case pb_lorawan.Region_CN_470_510.String():
		err = errors.NewErrInternal("China 470-510 MHz band not supported")
	case pb_lorawan.Region_AS_923.String():
		err = errors.NewErrInternal("Asia 923 MHz band not supported")
	case pb_lorawan.Region_KR_920_923.String():
		err = errors.NewErrInternal("South Korea 920-923 MHz band not supported")
	default:
		err = errors.NewErrInvalidArgument("Frequency Band", "unknown")
	}
	if err != nil {
		return
	}
	band = &b

	// TTN-specific configuration
	if region == "EU_863_870" {
		// TTN uses SF9BW125 in RX2
		band.RX2DataRate = 3
		// TTN frequency plan includes extra channels next to the default channels:
		band.UplinkChannels = []lora.Channel{
			lora.Channel{Frequency: 868100000, DataRates: []int{0, 1, 2, 3, 4, 5}},
			lora.Channel{Frequency: 868300000, DataRates: []int{0, 1, 2, 3, 4, 5, 6}}, // Also SF7BW250
			lora.Channel{Frequency: 868500000, DataRates: []int{0, 1, 2, 3, 4, 5}},
			lora.Channel{Frequency: 868800000, DataRates: []int{7}}, // FSK 50kbps
			lora.Channel{Frequency: 867100000, DataRates: []int{0, 1, 2, 3, 4, 5}},
			lora.Channel{Frequency: 867300000, DataRates: []int{0, 1, 2, 3, 4, 5}},
			lora.Channel{Frequency: 867500000, DataRates: []int{0, 1, 2, 3, 4, 5}},
			lora.Channel{Frequency: 867700000, DataRates: []int{0, 1, 2, 3, 4, 5}},
			lora.Channel{Frequency: 867900000, DataRates: []int{0, 1, 2, 3, 4, 5}},
		}
		band.DownlinkChannels = band.UplinkChannels
	}

	return
}

func (r *router) buildDownlinkOptions(uplink *pb.UplinkMessage, isActivation bool, gateway *gateway.Gateway) (downlinkOptions []*pb_broker.DownlinkOption) {
	var options []*pb_broker.DownlinkOption

	gatewayStatus, _ := gateway.Status.Get() // This just returns empty if non-existing

	lorawanMetadata := uplink.ProtocolMetadata.GetLorawan()
	if lorawanMetadata == nil {
		return // We can't handle any other protocols than LoRaWAN yet
	}

	region := gatewayStatus.Region
	if region == "" {
		region = guessRegion(uplink.GatewayMetadata.Frequency)
	}
	band, err := getBand(region)
	if err != nil {
		return // We can't handle this region
	}

	var dataRate lora.DataRate

	// LORA Modulation
	if lorawanMetadata.Modulation == pb_lorawan.Modulation_LORA {
		dataRate.Modulation = lora.LoRaModulation
		dr, err := types.ParseDataRate(lorawanMetadata.DataRate)
		if err != nil {
			return // Invalid packet, probably won't happen if the gateway is just doing its job
		}
		dataRate.Bandwidth = int(dr.Bandwidth)
		dataRate.SpreadFactor = int(dr.SpreadingFactor)
	}

	if lorawanMetadata.Modulation == pb_lorawan.Modulation_FSK {
		dataRate.Modulation = lora.FSKModulation
		dataRate.BitRate = int(lorawanMetadata.BitRate)
	}

	uplinkDRIndex, err := band.GetDataRate(dataRate)
	if err != nil {
		return // Invalid packet, probably won't happen if the gateway is just doing its job
	}

	// Configuration for RX2
	{
		power := int32(band.DefaultTXPower)
		if region == "EU_863_870" {
			power = 27 // The EU Downlink frequency allows up to 27dBm
			if isActivation {
				// TTN uses SF9BW125 in RX2, we have to reset this for joins
				band.RX2DataRate = 0
			}
		}
		dataRate, _ := types.ConvertDataRate(band.DataRates[band.RX2DataRate])
		delay := band.ReceiveDelay2
		if isActivation {
			delay = band.JoinAcceptDelay2
		}
		rx2 := &pb_broker.DownlinkOption{
			GatewayId: gateway.ID,
			ProtocolConfig: &pb_protocol.TxConfiguration{Protocol: &pb_protocol.TxConfiguration_Lorawan{Lorawan: &pb_lorawan.TxConfiguration{
				Modulation: pb_lorawan.Modulation_LORA, // RX2 is always LoRa
				DataRate:   dataRate.String(),          // This is default
				CodingRate: lorawanMetadata.CodingRate, // Let's just take this from the Rx
			}}},
			GatewayConfig: &pb_gateway.TxConfiguration{
				Timestamp:             uplink.GatewayMetadata.Timestamp + uint32(delay/1000),
				RfChain:               0,
				PolarizationInversion: true,
				Frequency:             uint64(band.RX2Frequency),
				Power:                 power,
			},
		}
		options = append(options, rx2)
	}

	// Configuration for RX1
	{
		uplinkChannel, err := band.GetChannel(int(uplink.GatewayMetadata.Frequency), nil)
		if err == nil {
			downlinkChannel := band.DownlinkChannels[band.GetRX1Channel(uplinkChannel)]
			downlinkDRIndex, err := band.GetRX1DataRateForOffset(uplinkDRIndex, 0)
			if err == nil {
				var modulation pb_lorawan.Modulation
				var dataRateString string
				var bitRate int
				var frequencyDeviation int

				dr := band.DataRates[downlinkDRIndex]
				if dr.Modulation == lora.LoRaModulation {
					modulation = pb_lorawan.Modulation_LORA
					dataRate, _ := types.ConvertDataRate(dr)
					dataRateString = dataRate.String()
				}

				if dr.Modulation == lora.FSKModulation {
					modulation = pb_lorawan.Modulation_FSK
					bitRate = dr.BitRate
					frequencyDeviation = bitRate / 2
				}

				delay := band.ReceiveDelay1
				if isActivation {
					delay = band.JoinAcceptDelay1
				}
				rx1 := &pb_broker.DownlinkOption{
					GatewayId: gateway.ID,
					ProtocolConfig: &pb_protocol.TxConfiguration{Protocol: &pb_protocol.TxConfiguration_Lorawan{Lorawan: &pb_lorawan.TxConfiguration{
						Modulation: modulation,
						DataRate:   dataRateString,
						BitRate:    uint32(bitRate),
						CodingRate: lorawanMetadata.CodingRate, // Let's just take this from the Rx
					}}},
					GatewayConfig: &pb_gateway.TxConfiguration{
						Timestamp:             uplink.GatewayMetadata.Timestamp + uint32(delay/1000),
						RfChain:               0,
						PolarizationInversion: true,
						Frequency:             uint64(downlinkChannel.Frequency),
						Power:                 int32(band.DefaultTXPower),
						FrequencyDeviation:    uint32(frequencyDeviation),
					},
				}
				options = append(options, rx1)
			}
		}
	}

	computeDownlinkScores(gateway, uplink, options)

	for _, option := range options {
		// Add router ID to downlink option
		if r.Component != nil && r.Component.Identity != nil {
			option.Identifier = fmt.Sprintf("%s:%s", r.Component.Identity.Id, option.Identifier)
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

	region := gatewayStatus.Region
	if region == "" {
		region = guessRegion(uplink.GatewayMetadata.Frequency)
	}

	gatewayRx, _ := gateway.Utilization.Get()
	for _, option := range options {

		// Invalid if no LoRaWAN
		lorawan := option.GetProtocolConfig().GetLorawan()
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
			if uplink.GatewayMetadata.Snr < 5 {
				signalScore += 10
			}
			// Prefer good RSSI
			signalScore += math.Min(float64(uplink.GatewayMetadata.Rssi*-0.1), 10)
		}

		utilizationScore := 0.0 // Between 0 and 40 (lower is better) will be over 100 if forbidden
		{
			// Avoid gateways that do more Rx
			utilizationScore += math.Min(gatewayRx*50, 20) / 2 // 40% utilization = 10 (max)

			// Avoid busy channels
			freq := option.GatewayConfig.Frequency
			channelRx, channelTx := gateway.Utilization.GetChannel(freq)
			utilizationScore += math.Min((channelTx+channelRx)*200, 20) / 2 // 10% utilization = 10 (max)

			// European Duty Cycle
			if region == "EU_863_870" {
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
			id, conflicts := gateway.Schedule.GetOption(option.GatewayConfig.Timestamp, uint32(time/1000))
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

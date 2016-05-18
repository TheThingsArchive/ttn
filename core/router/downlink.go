package router

import (
	"errors"
	"math"

	pb_broker "github.com/TheThingsNetwork/ttn/api/broker"
	pb_gateway "github.com/TheThingsNetwork/ttn/api/gateway"
	pb_semtech "github.com/TheThingsNetwork/ttn/api/gateway/semtech"
	pb_protocol "github.com/TheThingsNetwork/ttn/api/protocol"
	pb_lorawan "github.com/TheThingsNetwork/ttn/api/protocol/lorawan"
	pb "github.com/TheThingsNetwork/ttn/api/router"
	"github.com/TheThingsNetwork/ttn/core/router/gateway"
	"github.com/TheThingsNetwork/ttn/core/types"
	"github.com/TheThingsNetwork/ttn/utils/toa"
	lora "github.com/brocaar/lorawan/band"
)

func (r *router) SubscribeDownlink(gatewayEUI types.GatewayEUI) (<-chan *pb.DownlinkMessage, error) {
	gateway := r.getGateway(gatewayEUI)
	if fromSchedule := gateway.Schedule.Subscribe(); fromSchedule != nil {
		toGateway := make(chan *pb.DownlinkMessage)
		go func() {
			for message := range fromSchedule {
				gateway.Utilization.AddTx(message)
				toGateway <- message
			}
			close(toGateway)
		}()
		return toGateway, nil
	}
	return nil, errors.New("ttn/router: Gateway downlink not available")
}

func (r *router) UnsubscribeDownlink(gatewayEUI types.GatewayEUI) error {
	r.getGateway(gatewayEUI).Schedule.Stop()
	return nil
}

func (r *router) HandleDownlink(downlink *pb_broker.DownlinkMessage) error {
	option := downlink.DownlinkOption

	gateway := r.getGateway(*option.GatewayEui)

	downlinkMessage := &pb.DownlinkMessage{
		Payload:               downlink.Payload,
		ProtocolConfiguration: option.ProtocolConfig,
		GatewayConfiguration:  option.GatewayConfig,
	}

	err := gateway.Schedule.Schedule(option.Identifier, downlinkMessage)
	if err != nil {
		return err
	}

	return nil
}

func getBand(region string) (band *lora.Band, err error) {
	var b lora.Band

	switch region {
	case "EU_863_870":
		b, err = lora.GetConfig(lora.EU_863_870)
	case "US_902_928":
		b, err = lora.GetConfig(lora.US_902_928)
	case "CN_779_787":
		err = errors.New("ttn/router: China 779-787 MHz band not supported")
	case "EU_433":
		err = errors.New("ttn/router: Europe 433 MHz band not supported")
	case "AU_915_928":
		b, err = lora.GetConfig(lora.AU_915_928)
	case "CN_470_510":
		err = errors.New("ttn/router: China 470-510 MHz band not supported")
	default:
		err = errors.New("ttn/router: Unknown band")
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

func buildDownlinkOptions(uplink *pb.UplinkMessage, isActivation bool, gateway *gateway.Gateway) (downlinkOptions []*pb_broker.DownlinkOption) {
	var options []*pb_broker.DownlinkOption

	gatewayStatus, _ := gateway.Status.Get() // This just returns empty if non-existing

	lorawanMetadata := uplink.ProtocolMetadata.GetLorawan()
	if lorawanMetadata == nil {
		return // We can't handle any other protocols than LoRaWAN yet
	}

	semtechMetadata := uplink.GatewayMetadata.GetSemtech()
	if semtechMetadata == nil {
		return // We can't handle any other gateways than Semtech yet
	}

	band, err := getBand(gatewayStatus.Region)
	if err != nil {
		return // We can't handle this region
	}

	dr, err := types.ParseDataRate(lorawanMetadata.DataRate)
	if err != nil {
		return // Invalid packet, probably won't happen if the gateway is just doing its job
	}
	uplinkDRIndex, err := band.GetDataRate(lora.DataRate{Modulation: lora.LoRaModulation, SpreadFactor: int(dr.SpreadingFactor), Bandwidth: int(dr.Bandwidth)})
	if err != nil {
		return // Invalid packet, probably won't happen if the gateway is just doing its job
	}

	// Configuration for RX1
	{
		uplinkChannel, err := band.GetChannel(int(uplink.GatewayMetadata.Frequency), uplinkDRIndex)
		if err == nil {
			downlinkChannel := band.DownlinkChannels[band.GetRX1Channel(uplinkChannel)]
			downlinkDRIndex, err := band.GetRX1DataRateForOffset(uplinkDRIndex, 0)
			if err == nil {
				dataRate, _ := types.ConvertDataRate(band.DataRates[downlinkDRIndex])
				delay := band.ReceiveDelay1
				if isActivation {
					delay = band.JoinAcceptDelay1
				}
				rx1 := &pb_broker.DownlinkOption{
					GatewayEui: &gateway.EUI,
					ProtocolConfig: &pb_protocol.TxConfiguration{Protocol: &pb_protocol.TxConfiguration_Lorawan{Lorawan: &pb_lorawan.TxConfiguration{
						Modulation: pb_lorawan.Modulation_LORA, // We only support LoRa
						DataRate:   dataRate.String(),          // This is default
						CodingRate: lorawanMetadata.CodingRate, // Let's just take this from the Rx
					}}},
					GatewayConfig: &pb_gateway.TxConfiguration{Gateway: &pb_gateway.TxConfiguration_Semtech{Semtech: &pb_semtech.TxConfiguration{
						Timestamp:             semtechMetadata.Timestamp + uint32(delay/1000),
						RfChain:               0,
						PolarizationInversion: true,
					}},
						Frequency: uint64(downlinkChannel.Frequency),
						Power:     int32(band.DefaultTXPower),
					},
				}
				options = append(options, rx1)
			}
		}
	}

	// Configuration for RX2
	{
		power := int32(band.DefaultTXPower)
		if gatewayStatus.Region == "EU_863_870" {
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
			GatewayEui: &gateway.EUI,
			ProtocolConfig: &pb_protocol.TxConfiguration{Protocol: &pb_protocol.TxConfiguration_Lorawan{Lorawan: &pb_lorawan.TxConfiguration{
				Modulation: pb_lorawan.Modulation_LORA, // We only support LoRa
				DataRate:   dataRate.String(),          // This is default
				CodingRate: lorawanMetadata.CodingRate, // Let's just take this from the Rx
			}}},
			GatewayConfig: &pb_gateway.TxConfiguration{Gateway: &pb_gateway.TxConfiguration_Semtech{Semtech: &pb_semtech.TxConfiguration{
				Timestamp:             semtechMetadata.Timestamp + uint32(delay/1000),
				RfChain:               0,
				PolarizationInversion: true,
			}},
				Frequency: uint64(band.RX2Frequency),
				Power:     power,
			},
		}
		options = append(options, rx2)
	}

	computeDownlinkScores(gateway, uplink, options)

	// Filter all illegal options
	for _, option := range options {
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

	gatewayRx, _ := gateway.Utilization.Get()
	for _, option := range options {

		// Calculate max ToA
		time, _ := toa.Compute(
			51+13, // Max MACPayload plus LoRaWAN header, TODO: What is the length we should use?
			option.GetProtocolConfig().GetLorawan().DataRate,
			option.GetProtocolConfig().GetLorawan().CodingRate,
		)

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
			utilizationScore += math.Min(gatewayRx*50, 20) // 40% utilization = 20 (max)

			// Avoid busy channels
			freq := option.GatewayConfig.Frequency
			channelRx, channelTx := gateway.Utilization.GetChannel(freq)
			utilizationScore += math.Min((channelTx+channelRx)*200, 20) // 10% utilization = 20 (max)

			// Enforce European Duty Cycle
			if gatewayStatus.Region == "EU_863_870" {
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
			}
		}

		scheduleScore := 0.0 // Between 0 and 30 (lower is better) will be over 100 if forbidden
		{
			id, conflicts := gateway.Schedule.GetOption(option.GatewayConfig.GetSemtech().Timestamp, uint32(time/1000))
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

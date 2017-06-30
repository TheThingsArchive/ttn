// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package handler

import (
	"strings"

	ttnlog "github.com/TheThingsNetwork/go-utils/log"
	pb_broker "github.com/TheThingsNetwork/ttn/api/broker"
	"github.com/TheThingsNetwork/ttn/core/handler/device"
	"github.com/TheThingsNetwork/ttn/core/types"
)

// ConvertMetadata converts the protobuf matadata to application metadata
func (h *handler) ConvertMetadata(ctx ttnlog.Interface, ttnUp *pb_broker.DeduplicatedUplinkMessage, appUp *types.UplinkMessage, dev *device.Device) error {
	ctx = ctx.WithField("NumGateways", len(ttnUp.GatewayMetadata))

	// Transform Metadata
	appUp.Metadata.Time = types.BuildTime(ttnUp.ServerTime)
	if lorawan := ttnUp.ProtocolMetadata.GetLorawan(); lorawan != nil {
		appUp.Metadata.Modulation = lorawan.Modulation.String()
		appUp.Metadata.DataRate = lorawan.DataRate
		appUp.Metadata.Bitrate = lorawan.BitRate
		appUp.Metadata.CodingRate = lorawan.CodingRate
	}

	// Transform Gateway Metadata
	appUp.Metadata.Gateways = make([]types.GatewayMetadata, 0, len(ttnUp.GatewayMetadata))
	for i, in := range ttnUp.GatewayMetadata {

		// Same for all gateways, take first one
		if i == 0 {
			appUp.Metadata.Frequency = float32(float64(in.Frequency) / 1000000)
		}

		gatewayMetadata := types.GatewayMetadata{
			GtwID:      in.GatewayId,
			GtwTrusted: in.GatewayTrusted,
			Timestamp:  in.Timestamp,
			Time:       types.BuildTime(in.Time),
			Channel:    in.Channel,
			RFChain:    in.RfChain,
			RSSISignal: in.RssiSignal,
			SNR:        in.Snr,
		}

		if location := in.GetLocation(); location != nil {
			gatewayMetadata.Altitude = location.Altitude
			gatewayMetadata.Longitude = location.Longitude
			gatewayMetadata.Latitude = location.Latitude
			gatewayMetadata.Accuracy = location.Accuracy
			gatewayMetadata.Source = strings.ToLower(location.Source.String())
		}

		if antennas := in.GetAntennas(); len(antennas) > 0 {
			for _, antenna := range antennas {
				gatewayMetadata.Antenna = uint8(antenna.Antenna)
				gatewayMetadata.Channel = antenna.Channel
				gatewayMetadata.RSSIChannel = antenna.RssiChannel
				gatewayMetadata.RSSIStandardDeviation = antenna.RssiStandardDeviation
				gatewayMetadata.RSSISignal = antenna.RssiSignal
				gatewayMetadata.SNR = antenna.Snr
				gatewayMetadata.FrequencyOffset = antenna.FrequencyOffset
				gatewayMetadata.EncryptedTime = antenna.EncryptedTime
				appUp.Metadata.Gateways = append(appUp.Metadata.Gateways, gatewayMetadata)
			}
		} else {
			appUp.Metadata.Gateways = append(appUp.Metadata.Gateways, gatewayMetadata)
		}
	}

	// Inject Device Metadata
	appUp.Metadata.LocationMetadata.Latitude = dev.Latitude
	appUp.Metadata.LocationMetadata.Longitude = dev.Longitude
	appUp.Metadata.LocationMetadata.Altitude = dev.Altitude
	appUp.Metadata.LocationMetadata.Source = "registry"

	return nil
}

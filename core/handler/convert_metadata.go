// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package handler

import (
	pb_broker "github.com/TheThingsNetwork/ttn/api/broker"
	"github.com/TheThingsNetwork/ttn/core/types"
	"github.com/apex/log"
)

// ConvertMetadata converts the protobuf matadata to application metadata
func (h *handler) ConvertMetadata(ctx log.Interface, ttnUp *pb_broker.DeduplicatedUplinkMessage, appUp *types.UplinkMessage) error {
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
			GtwID:     in.GatewayId,
			Timestamp: in.Timestamp,
			Time:      types.BuildTime(in.Time),
			Channel:   in.Channel,
			RFChain:   in.RfChain,
			RSSI:      in.Rssi,
			SNR:       in.Snr,
		}

		if gps := in.GetGps(); gps != nil {
			gatewayMetadata.Altitude = gps.Altitude
			gatewayMetadata.Longitude = gps.Longitude
			gatewayMetadata.Latitude = gps.Latitude
		}

		appUp.Metadata.Gateways = append(appUp.Metadata.Gateways, gatewayMetadata)
	}

	return nil
}

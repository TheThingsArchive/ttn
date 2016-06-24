// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package handler

import (
	"time"

	pb_broker "github.com/TheThingsNetwork/ttn/api/broker"
	"github.com/TheThingsNetwork/ttn/mqtt"
	"github.com/apex/log"
)

// ConvertMetadata converts the protobuf matadata to application metadata
func (h *handler) ConvertMetadata(ctx log.Interface, ttnUp *pb_broker.DeduplicatedUplinkMessage, appUp *mqtt.UplinkMessage) error {

	ctx = ctx.WithField("NumGateways", len(ttnUp.GatewayMetadata))

	// Transform Metadata
	metadata := make([]*mqtt.Metadata, 0, len(ttnUp.GatewayMetadata))
	for _, in := range ttnUp.GatewayMetadata {
		out := &mqtt.Metadata{}

		out.ServerTime = time.Unix(0, 0).Add(time.Duration(ttnUp.ServerTime)).UTC().Format(time.RFC3339Nano)

		if lorawan := ttnUp.ProtocolMetadata.GetLorawan(); lorawan != nil {
			out.DataRate = lorawan.DataRate
			out.CodingRate = lorawan.CodingRate
			out.Modulation = lorawan.Modulation.String()
		}

		if in.GatewayEui != nil {
			out.GatewayEUI = *in.GatewayEui
		}
		out.Timestamp = in.Timestamp
		out.Time = time.Unix(0, 0).Add(time.Duration(in.Time)).UTC().Format(time.RFC3339Nano)
		out.Channel = in.Channel
		out.RFChain = in.RfChain
		out.Frequency = float32(float64(in.Frequency) / 1000000)
		out.Rssi = in.Rssi
		out.Lsnr = in.Snr

		if gps := in.GetGps(); gps != nil {
			out.Altitude = gps.Altitude
			out.Longitude = gps.Longitude
			out.Latitude = gps.Latitude
		}

		metadata = append(metadata, out)
	}

	appUp.Metadata = metadata

	return nil
}

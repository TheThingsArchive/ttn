// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package util

import (
	"github.com/TheThingsNetwork/ttn/api/gateway"
	"github.com/TheThingsNetwork/ttn/api/protocol"
	"github.com/TheThingsNetwork/ttn/api/protocol/lorawan"
)

// GetProtocolMetadata returns protocol metadata for the given datarate
func GetProtocolMetadata(dataRate string) *protocol.RxMetadata {
	return &protocol.RxMetadata{Protocol: &protocol.RxMetadata_Lorawan{Lorawan: &lorawan.Metadata{
		CodingRate: "4/5",
		DataRate:   dataRate,
		Modulation: lorawan.Modulation_LORA,
	}}}
}

// GetGatewayMetadata returns gateway metadata for the given gateway ID and frequency
func GetGatewayMetadata(id string, freq uint64) *gateway.RxMetadata {
	return &gateway.RxMetadata{
		GatewayId: id,
		Timestamp: 0,
		Frequency: freq,
		Rssi:      -25.0,
		Snr:       5.0,
	}
}

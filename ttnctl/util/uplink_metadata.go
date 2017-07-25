// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package util

import (
	"github.com/TheThingsNetwork/api/gateway"
	"github.com/TheThingsNetwork/api/protocol"
	"github.com/TheThingsNetwork/api/protocol/lorawan"
)

// GetProtocolMetadata returns protocol metadata for the given datarate
func GetProtocolMetadata(dataRate string) *protocol.RxMetadata {
	return &protocol.RxMetadata{Protocol: &protocol.RxMetadata_LoRaWAN{LoRaWAN: &lorawan.Metadata{
		CodingRate: "4/5",
		DataRate:   dataRate,
		Modulation: lorawan.Modulation_LORA,
	}}}
}

// GetGatewayMetadata returns gateway metadata for the given gateway ID and frequency
func GetGatewayMetadata(id string, freq uint64) *gateway.RxMetadata {
	return &gateway.RxMetadata{
		GatewayID: id,
		Timestamp: 0,
		Frequency: freq,
		RSSI:      -25.0,
		SNR:       5.0,
	}
}

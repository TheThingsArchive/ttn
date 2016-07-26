// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package util

import (
	"github.com/TheThingsNetwork/ttn/api/gateway"
	"github.com/TheThingsNetwork/ttn/api/protocol"
	"github.com/TheThingsNetwork/ttn/api/protocol/lorawan"
	"github.com/TheThingsNetwork/ttn/core/types"
)

// GetProtocolMetadata returns protocol metadata for the given datarate
func GetProtocolMetadata(dataRate string) *protocol.RxMetadata {
	return &protocol.RxMetadata{Protocol: &protocol.RxMetadata_Lorawan{Lorawan: &lorawan.Metadata{
		CodingRate: "4/5",
		DataRate:   dataRate,
		Modulation: lorawan.Modulation_LORA,
	}}}
}

// GetGatewayMetadata returns gateway metadata for the given gateway EUI and frequency
func GetGatewayMetadata(eui types.GatewayEUI, freq uint64) *gateway.RxMetadata {
	return &gateway.RxMetadata{
		GatewayEui: &eui,
		Timestamp:  0,
		Frequency:  freq,
		Rssi:       -25.0,
		Snr:        5.0,
	}
}

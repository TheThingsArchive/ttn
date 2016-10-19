// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package amqp

import "github.com/TheThingsNetwork/ttn/core/types"

// LocationMetadata contains GPS coordinates
type LocationMetadata struct {
	Altitude  int32   `json:"altitude,omitempty"`
	Longitude float32 `json:"longitude,omitempty"`
	Latitude  float32 `json:"latitude,omitempty"`
}

// GatewayMetadata contains metadata for each gateway that received a message
type GatewayMetadata struct {
	GtwID     string         `json:"gtw_id,omitempty"`
	Timestamp uint32         `json:"timestamp,omitempty"`
	Time      types.JSONTime `json:"time,omitempty"`
	Channel   uint32         `json:"channel"`
	RSSI      float32        `json:"rssi,omitempty"`
	SNR       float32        `json:"snr,omitempty"`
	RFChain   uint32         `json:"rf_chain,omitempty"`
	LocationMetadata
}

// Metadata contains metadata of a message
type Metadata struct {
	Time       types.JSONTime    `json:"time,omitempty,omitempty"`
	Frequency  float32           `json:"frequency,omitempty"`
	Modulation string            `json:"modulation,omitempty"`
	DataRate   string            `json:"data_rate,omitempty"`
	Bitrate    uint32            `json:"bit_rate,omitempty"`
	CodingRate string            `json:"coding_rate,omitempty"`
	Gateways   []GatewayMetadata `json:"gateways,omitempty"`
	LocationMetadata
}

// UplinkMessage represents an application-layer uplink message
type UplinkMessage struct {
	AppID         string                 `json:"app_id,omitempty"`
	DevID         string                 `json:"dev_id,omitempty"`
	FPort         uint8                  `json:"port"`
	FCnt          uint32                 `json:"counter"`
	PayloadRaw    []byte                 `json:"payload_raw"`
	PayloadFields map[string]interface{} `json:"payload_fields,omitempty"`
	Metadata      Metadata               `json:"metadata,omitempty"`
}

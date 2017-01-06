// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package types

// Metadata contains metadata of a message
type Metadata struct {
	Time       JSONTime          `json:"time,omitempty,omitempty"`
	Frequency  float32           `json:"frequency,omitempty"`
	Modulation string            `json:"modulation,omitempty"`
	DataRate   string            `json:"data_rate,omitempty"`
	Bitrate    uint32            `json:"bit_rate,omitempty"`
	CodingRate string            `json:"coding_rate,omitempty"`
	Gateways   []GatewayMetadata `json:"gateways,omitempty"`
	LocationMetadata
}

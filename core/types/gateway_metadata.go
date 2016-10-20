// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package types

// GatewayMetadata contains metadata for each gateway that received a message
type GatewayMetadata struct {
	GtwID     string   `json:"gtw_id,omitempty"`
	Timestamp uint32   `json:"timestamp,omitempty"`
	Time      JSONTime `json:"time,omitempty"`
	Channel   uint32   `json:"channel"`
	RSSI      float32  `json:"rssi,omitempty"`
	SNR       float32  `json:"snr,omitempty"`
	RFChain   uint32   `json:"rf_chain,omitempty"`
	LocationMetadata
}

// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package types

// UplinkMessage represents an application-layer uplink message
type UplinkMessage struct {
	AppID          string                 `json:"app_id,omitempty"`
	DevID          string                 `json:"dev_id,omitempty"`
	HardwareSerial string                 `json:"hardware_serial,omitempty"`
	FPort          uint8                  `json:"port"`
	FCnt           uint32                 `json:"counter"`
	IsRetry        bool                   `json:"is_retry,omitempty"`
	PayloadRaw     []byte                 `json:"payload_raw"`
	PayloadFields  map[string]interface{} `json:"payload_fields,omitempty"`
	Metadata       Metadata               `json:"metadata,omitempty"`
}

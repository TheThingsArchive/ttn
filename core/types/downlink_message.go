// Copyright © 2017 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package types

// DownlinkMessage represents an application-layer downlink message
type DownlinkMessage struct {
	AppID         string                 `json:"app_id,omitempty"`
	DevID         string                 `json:"dev_id,omitempty"`
	FPort         uint8                  `json:"port"`
	Confirmed     bool                   `json:"confirmed,omitempty"`
	Schedule      string                 `json:"schedule,omitempty"` // allowed values: "replace" (default), "first", "last"
	PayloadRaw    []byte                 `json:"payload_raw,omitempty"`
	PayloadFields map[string]interface{} `json:"payload_fields,omitempty"`
}

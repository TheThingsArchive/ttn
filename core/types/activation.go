// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package types

// Activation messages are used to notify application of a device activation
type Activation struct {
	AppID    string   `json:"app_id,omitempty"`
	DevID    string   `json:"dev_id,omitempty"`
	AppEUI   AppEUI   `json:"app_eui,omitempty"`
	DevEUI   DevEUI   `json:"dev_eui,omitempty"`
	DevAddr  DevAddr  `json:"dev_addr,omitempty"`
	Metadata Metadata `json:"metadata,omitempty"`
}

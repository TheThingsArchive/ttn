// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package mqtt

import "github.com/TheThingsNetwork/ttn/core/types"

// Metadata contains the metadata that is passed up to the application
type Metadata struct {
	Frequency  float32          `json:"frequency"`
	DataRate   string           `json:"datarate"`
	CodingRate string           `json:"codingrate"`
	Timestamp  uint32           `json:"gateway_timestamp"`
	Time       string           `json:"gateway_time,omitempty"`
	ServerTime string           `json:"server_time"`
	Channel    uint32           `json:"channel"`
	Rssi       float32          `json:"rssi"`
	Lsnr       float32          `json:"lsnr"`
	RFChain    uint32           `json:"rfchain"`
	Modulation string           `json:"modulation"`
	GatewayEUI types.GatewayEUI `json:"gateway_eui"`
	Altitude   int32            `json:"altitude"`
	Longitude  float32          `json:"longitude"`
	Latitude   float32          `json:"latitude"`
}

// UplinkMessage represents an application-layer uplink message
type UplinkMessage struct {
	AppID    string                 `json:"app_id,omitempty"`
	AppEUI   types.AppEUI           `json:"app_eui,omitempty"`
	DevEUI   types.DevEUI           `json:"dev_eui,omitempty"`
	Payload  []byte                 `json:"payload,omitempty"`
	FPort    uint8                  `json:"port,omitempty"`
	Fields   map[string]interface{} `json:"fields,omitempty"`
	FCnt     uint32                 `json:"counter,omitempty"`
	Metadata []*Metadata            `json:"metadata,omitempty"`
}

// DownlinkMessage represents an application-layer downlink message
type DownlinkMessage struct {
	AppID   string                 `json:"app_id,omitempty"`
	AppEUI  types.AppEUI           `json:"app_eui,omitempty"`
	DevEUI  types.DevEUI           `json:"dev_eui,omitempty"`
	Payload []byte                 `json:"payload,omitempty"`
	FPort   uint8                  `json:"port,omitempty"`
	Fields  map[string]interface{} `json:"fields,omitempty"`
	TTL     string                 `json:"ttl,omitempty"`
}

// Activation are used to notify application of a device activation
type Activation struct {
	AppID    string       `json:"app_id,omitempty"`
	AppEUI   types.AppEUI `json:"app_eui,omitempty"`
	DevEUI   types.DevEUI `json:"dev_eui,omitempty"`
	Metadata []Metadata   `json:"metadata"`
}

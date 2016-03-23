// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

//go:generate msgp -io=false

package core

// DataUpAppReq represents the actual payloads sent to application on uplink
type DataUpAppReq struct {
	Payload  []byte        `msg:"payload" json:"payload"`
	Metadata []AppMetadata `msg:"metadata" json:"metadata"`
}

// JoinAppReq are used to notify application of an accepted OTAA
type OTAAAppReq struct {
	Metadata []AppMetadata `msg:"metadata" json:"metadata"`
}

// AppMetadata represents gathered metadata that are sent to gateways
type AppMetadata struct {
	Frequency  float32 `msg:"frequency" json:"frequency"`
	DataRate   string  `msg:"data_rate" json:"data_rate"`
	CodingRate string  `msg:"coding_rate" json:"coding_rate"`
	Timestamp  uint32  `msg:"timestamp" json:"timestamp"`
	Rssi       int32   `msg:"rssi" json:"rssi"`
	Lsnr       float32 `msg:"lsnr" json:"lsnr"`
	Altitude   int32   `msg:"altitude" json:"altitude"`
	Longitude  float32 `msg:"longitude" json:"longitude"`
	Latitude   float32 `msg:"latitude" json:"latitude"`
}

// DataDownAppReq represents downlink messages sent by applications
type DataDownAppReq struct {
	Payload []byte `msg:"payload" json:"payload"`
	TTL     string `msg:"ttl" json:"ttl"`
}

// ABPSubAppReq defines the shape of the request made by an application to the handler
type ABPSubAppReq struct {
	DevAddr string `msg:"dev_addr" json:"dev_addr"`
	NwkSKey string `msg:"nwks_key" json:"nwks_key"`
	AppSKey string `msg:"apps_key" json:"apps_key"`
}

// AuthBrokerClient gathers both BrokerClient & BrokerManagerClient interfaces
type AuthBrokerClient interface {
	BrokerClient
	BrokerManagerClient
}

// AuthHandlerClient gathers both HandlerClient & HandlerManagerClient interfaces
type AuthHandlerClient interface {
	HandlerClient
	HandlerManagerClient
}

// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

//go:generate msgp -io=false

package core

// DataUpAppReq represents the actual payloads sent to application on uplink
type DataUpAppReq struct {
	Payload  []byte        `msg:"payload" json:"payload"`
	Metadata []AppMetadata `msg:"metadata" json:"metadata"`
}

// OTAAAppReq are used to notify application of an accepted OTAA
type OTAAAppReq struct {
	Metadata []AppMetadata `msg:"metadata" json:"metadata"`
}

// AppMetadata represents gathered metadata that are sent to gateways
type AppMetadata struct {
	Frequency  float32 `msg:"frequency" json:"frequency"`
	DataRate   string  `msg:"datarate" json:"datarate"`
	CodingRate string  `msg:"codingrate" json:"codingrate"`
	Timestamp  uint32  `msg:"timestamp" json:"timestamp"`
	Time       string  `msg:"time" json:"time"`
	Rssi       int32   `msg:"rssi" json:"rssi"`
	Lsnr       float32 `msg:"lsnr" json:"lsnr"`
	RFChain    uint32  `msg:"rfchain" json:"rfchain"`
	CRCStatus  int32   `msg:"crc" json:"crc"`
	Modulation string  `msg:"modulation" json:"modulation"`
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
	DevAddr string `msg:"devaddr" json:"devaddr"`
	NwkSKey string `msg:"nwkskey" json:"nwkskey"`
	AppSKey string `msg:"appskey" json:"appskey"`
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

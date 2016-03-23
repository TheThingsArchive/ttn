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
	Frequency  float32 `msg:"freq" json:"freq"`
	DataRate   string  `msg:"datr" json:"datr"`
	CodingRate string  `msg:"codr" json:"codr"`
	Timestamp  uint32  `msg:"tmst" json:"tmst"`
	Time       string  `msg:"time" json:"time"`
	Rssi       int32   `msg:"rssi" json:"rssi"`
	Lsnr       float32 `msg:"lsnr" json:"lsnr"`
	RFChain    uint32  `msg:"rfch" json:"rfch"`
	CRCStatus  int32   `msg:"stat" json:"stat"`
	Modulation string  `msg:"modu" json:"modu"`
	Altitude   int32   `msg:"alti" json:"alti"`
	Longitude  float32 `msg:"long" json:"long"`
	Latitude   float32 `msg:"lati" json:"lati"`
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

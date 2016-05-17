// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package core

// DataUpAppReq represents the actual payloads sent to application on uplink
type DataUpAppReq struct {
	Payload  []byte                 `json:"payload"`
	Fields   map[string]interface{} `json:"fields,omitempty"`
	FPort    uint8                  `json:"port,omitempty"`
	FCnt     uint32                 `json:"counter"`
	DevEUI   string                 `json:"dev_eui"`
	Metadata []AppMetadata          `json:"metadata"`
}

// OTAAAppReq are used to notify application of an accepted OTAA
type OTAAAppReq struct {
	Metadata []AppMetadata `json:"metadata"`
}

// AppMetadata represents gathered metadata that are sent to gateways
type AppMetadata struct {
	Frequency  float32 `json:"frequency"`
	DataRate   string  `json:"datarate"`
	CodingRate string  `json:"codingrate"`
	Timestamp  uint32  `json:"gateway_timestamp"`
	Time       string  `json:"gateway_time,omitempty"`
	Channel    uint32  `json:"channel"`
	ServerTime string  `json:"server_time"`
	Rssi       int32   `json:"rssi"`
	Lsnr       float32 `json:"lsnr"`
	RFChain    uint32  `json:"rfchain"`
	CRCStatus  int32   `json:"crc"`
	Modulation string  `json:"modulation"`
	GatewayEUI string  `json:"gateway_eui"`
	Altitude   int32   `json:"altitude"`
	Longitude  float32 `json:"longitude"`
	Latitude   float32 `json:"latitude"`
}

// DataDownAppReq represents downlink messages sent by applications
type DataDownAppReq struct {
	Payload []byte `json:"payload"`
	FPort   uint8  `json:"port,omitempty"`
	TTL     string `json:"ttl"`
}

// ABPSubAppReq defines the shape of the request made by an application to the handler
type ABPSubAppReq struct {
	DevAddr string `json:"devaddr"`
	NwkSKey string `json:"nwkskey"`
	AppSKey string `json:"appskey"`
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

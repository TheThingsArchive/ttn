// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

//go:generate msgp -io=false

package core

import (
	"reflect"
)

// DataUpAppReq represents the actual payloads sent to application on uplink
type DataUpAppReq struct {
	Payload  []byte        `msg:"payload" json:"payload"`
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
}

// ABPSubAppReq defines the shape of the request made by an application to the handler
type ABPSubAppReq struct {
	DevAddr string `msg:"dev_addr" json:"dev_addr"`
	NwkSKey string `msg:"nwks_key" json:"nwks_key"`
	AppSKey string `msg:"apps_key" json:"apps_key"`
}

// ProtoMetaToAppMeta converts a set of Metadata generate with Protobuf to a set of valid
// AppMetadata ready to be marshaled to json
func ProtoMetaToAppMeta(srcs ...*Metadata) []AppMetadata {
	var dest []AppMetadata

	for _, src := range srcs {
		if src == nil {
			continue
		}
		to := new(AppMetadata)
		v := reflect.ValueOf(src).Elem()
		t := v.Type()
		d := reflect.ValueOf(to).Elem()

		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i).Name
			if d.FieldByName(field).CanSet() {
				d.FieldByName(field).Set(v.Field(i))
			}
		}

		dest = append(dest, *to)
	}

	return dest
}

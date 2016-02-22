// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package refactor

import (
	"encoding/json"
	"testing"

	. "github.com/TheThingsNetwork/ttn/utils/testing"
	"github.com/brocaar/lorawan"
)

func TestMarshalJSONRPacket(t *testing.T) {
	tests := []marshalJSONTest{
		marshalJSONTest{ // Empty Payload
			Packet:     NewRPacket(lorawan.PHYPayload{}, genFullMetadata()),
			WantFields: []string{},
		},
		marshalJSONTest{ // Empty Metadata
			Packet:     NewRPacket(genPHYPayload(true), Metadata{}),
			WantFields: []string{"payload", "metadata"},
		},
		marshalJSONTest{ // With Metadata and Payload
			Packet:     NewRPacket(genPHYPayload(true), genFullMetadata()),
			WantFields: []string{"payload", "metadata"},
		},
	}

	for _, test := range tests {
		Desc(t, "Marshal packet to json: %s", test.Packet.String())
		raw, _ := json.Marshal(test.Packet)
		checkFields(t, test.WantFields, raw)
	}
}

func TestUnmarshalJSONRPacket(t *testing.T) {
	tests := []unmarshalJSONTest{
		unmarshalJSONTest{
			JSON:       `{"payload":"gAQDAgEAAAAK4mTU97VqDnU=","metadata":{}}`,
			WantPacket: NewRPacket(genPHYPayload(true), Metadata{}),
		},
		unmarshalJSONTest{
			JSON:       `{"payload":"gAQDAgEAAAAK4mTU97VqDnU=","metadata":{"chan":2,"codr":"4/6","fdev":3,"freq":863.125,"imme":false,"ipol":false,"lsnr":5.2,"modu":"LORA","ncrc":true,"powe":3,"prea":8,"rfch":2,"rssi":-27,"size":14,"stat":0,"tmst":1452694288207288421,"datr":"LORA","time":"2016-01-13T14:11:28.207288421Z"}}`,
			WantPacket: NewRPacket(genPHYPayload(true), genFullMetadata()),
		},
		unmarshalJSONTest{
			JSON:       `invalid`,
			WantPacket: RPacket{},
		},
		unmarshalJSONTest{
			JSON:       `{"metadata":{}}`,
			WantPacket: RPacket{},
		},
	}

	for _, test := range tests {
		Desc(t, "Unmarshal json to packet: %s", test.JSON)
		packet := RPacket{}
		json.Unmarshal([]byte(test.JSON), &packet)
		checkPackets(t, test.WantPacket, packet)
	}
}

// ---- Declaration
type marshalJSONTest struct {
	Packet     Packet
	WantFields []string
}

type unmarshalJSONTest struct {
	JSON       string
	WantPacket Packet
}

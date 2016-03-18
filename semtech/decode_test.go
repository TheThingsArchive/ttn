// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package semtech

import (
	"testing"
	"time"

	"github.com/TheThingsNetwork/ttn/utils/pointer"
	. "github.com/TheThingsNetwork/ttn/utils/testing"
)

func TestUnmarshalBinary(t *testing.T) {
	tests := []struct {
		Desc       string
		JSON       string
		Header     []byte
		WantPacket Packet
		WantError  bool
	}{
		{
			Desc:      "Invalid PUSH_DATA, invalid gateway id",
			Header:    []byte{VERSION, 1, 2, PUSH_DATA, 1, 4, 5, 6},
			JSON:      `{}`,
			WantError: true,
		},
		{
			Desc:   "PUSH_DATA with no payload",
			Header: []byte{VERSION, 0x14, 0x42, PUSH_DATA, 1, 2, 3, 4, 5, 6, 7, 8},
			JSON:   `{}`,
			WantPacket: Packet{
				Version:    VERSION,
				Token:      []byte{0x14, 0x42},
				Identifier: PUSH_DATA,
				GatewayId:  []byte{1, 2, 3, 4, 5, 6, 7, 8},
				Payload:    &Payload{},
			},
			WantError: false,
		},
		{
			Desc:   "PUSH_DATA with only basic typed-attributes  uint, string, float64 and int",
			Header: []byte{1, 0x14, 0x42, PUSH_DATA, 1, 2, 3, 4, 5, 6, 7, 8},
			JSON:   `{"rxpk":[{"chan":14,"codr":"4/7","freq":873.14,"rssi":-42}]}`,
			WantPacket: Packet{
				Version:    VERSION,
				Token:      []byte{0x14, 0x42},
				Identifier: PUSH_DATA,
				GatewayId:  []byte{1, 2, 3, 4, 5, 6, 7, 8},
				Payload: &Payload{
					RXPK: []RXPK{
						RXPK{
							Chan: pointer.Uint32(14),
							Codr: pointer.String("4/7"),
							Freq: pointer.Float32(873.14),
							Rssi: pointer.Int32(-42),
						},
					},
				},
			},
			WantError: false,
		},
		{
			Desc:   "PUSH_DATA with datr field and modu -> LORA",
			Header: []byte{1, 0x14, 0x42, PUSH_DATA, 1, 2, 3, 4, 5, 6, 7, 8},
			JSON:   `{"rxpk":[{"modu":"LORA","datr":"SF7BW125"}]}`,
			WantPacket: Packet{
				Version:    VERSION,
				Token:      []byte{0x14, 0x42},
				Identifier: PUSH_DATA,
				GatewayId:  []byte{1, 2, 3, 4, 5, 6, 7, 8},
				Payload: &Payload{
					RXPK: []RXPK{
						RXPK{
							Datr: pointer.String("SF7BW125"),
							Modu: pointer.String("LORA"),
						},
					},
				},
			},
			WantError: false,
		},
		{
			Desc:   "PUSH_DATA with datr field and modu -> FSK",
			Header: []byte{1, 0x14, 0x42, PUSH_DATA, 1, 2, 3, 4, 5, 6, 7, 8},
			JSON:   `{"rxpk":[{"modu":"FSK","datr":50000}]}`,
			WantPacket: Packet{
				Version:    VERSION,
				Token:      []byte{0x14, 0x42},
				Identifier: PUSH_DATA,
				GatewayId:  []byte{1, 2, 3, 4, 5, 6, 7, 8},
				Payload: &Payload{
					RXPK: []RXPK{
						RXPK{
							Datr: pointer.String("50000"),
							Modu: pointer.String("FSK"),
						},
					},
				},
			},
			WantError: false,
		},
		{
			Desc:   "PUSH_DATA with time field",
			Header: []byte{1, 0x14, 0x42, PUSH_DATA, 1, 2, 3, 4, 5, 6, 7, 8},
			JSON:   `{"rxpk":[{"time":"2016-01-13T17:40:57.000000376Z"}]}`,
			WantPacket: Packet{
				Version:    VERSION,
				Token:      []byte{0x14, 0x42},
				Identifier: PUSH_DATA,
				GatewayId:  []byte{1, 2, 3, 4, 5, 6, 7, 8},
				Payload: &Payload{
					RXPK: []RXPK{
						RXPK{
							Time: pointer.Time(time.Date(2016, 1, 13, 17, 40, 57, 376, time.UTC)),
						},
					},
				},
			},
			WantError: false,
		},
		{
			Desc:   "PUSH_DATA with several RXPK",
			Header: []byte{1, 0x14, 0x42, PUSH_DATA, 1, 2, 3, 4, 5, 6, 7, 8},
			JSON:   `{"rxpk":[{"size":14},{"chan":14}]}`,
			WantPacket: Packet{
				Version:    VERSION,
				Token:      []byte{0x14, 0x42},
				Identifier: PUSH_DATA,
				GatewayId:  []byte{1, 2, 3, 4, 5, 6, 7, 8},
				Payload: &Payload{
					RXPK: []RXPK{
						RXPK{
							Size: pointer.Uint32(14),
						},
						RXPK{
							Chan: pointer.Uint32(14),
						},
					},
				},
			},
			WantError: false,
		},
		{
			Desc:   "PUSH_DATA with several RXPK and Stat(basic fields)",
			Header: []byte{1, 0x14, 0x42, PUSH_DATA, 1, 2, 3, 4, 5, 6, 7, 8},
			JSON:   `{"rxpk":[{"size":14}],"stat":{"ackr":0.78,"alti":72,"rxok":42}}`,
			WantPacket: Packet{
				Version:    VERSION,
				Token:      []byte{0x14, 0x42},
				Identifier: PUSH_DATA,
				GatewayId:  []byte{1, 2, 3, 4, 5, 6, 7, 8},
				Payload: &Payload{
					RXPK: []RXPK{
						RXPK{
							Size: pointer.Uint32(14),
						},
					},
					Stat: &Stat{
						Ackr: pointer.Float32(0.78),
						Alti: pointer.Int32(72),
						Rxok: pointer.Uint32(42),
					},
				},
			},
			WantError: false,
		},
		{
			Desc:   "PUSH_DATA with Stat(time field)",
			Header: []byte{1, 0x14, 0x42, PUSH_DATA, 1, 2, 3, 4, 5, 6, 7, 8},
			JSON:   `{"stat":{"time":"2016-01-13 17:40:57 GMT"}}`,
			WantPacket: Packet{
				Version:    VERSION,
				Token:      []byte{0x14, 0x42},
				Identifier: PUSH_DATA,
				GatewayId:  []byte{1, 2, 3, 4, 5, 6, 7, 8},
				Payload: &Payload{
					Stat: &Stat{
						Time: pointer.Time(time.Date(2016, 1, 13, 17, 40, 57, 0, time.UTC)),
					},
				},
			},
			WantError: false,
		},
		{
			Desc:   "PUSH_DATA with rxpk and txpk (?)",
			Header: []byte{1, 0x14, 0x42, PUSH_DATA, 1, 2, 3, 4, 5, 6, 7, 8},
			JSON:   `{"rxpk":[{"codr":"4/7","rssi":-42}],"txpk":{"ipol":true,"powe":12}}`,
			WantPacket: Packet{
				Version:    VERSION,
				Token:      []byte{0x14, 0x42},
				Identifier: PUSH_DATA,
				GatewayId:  []byte{1, 2, 3, 4, 5, 6, 7, 8},
				Payload: &Payload{
					RXPK: []RXPK{
						RXPK{
							Codr: pointer.String("4/7"),
							Rssi: pointer.Int32(-42),
						},
					},
					TXPK: &TXPK{
						Ipol: pointer.Bool(true),
						Powe: pointer.Uint32(12),
					},
				},
			},
			WantError: false,
		},
		{
			Desc:   "PUSH_ACK valid",
			Header: []byte{1, 0x14, 0x42, PUSH_ACK},
			WantPacket: Packet{
				Version:    VERSION,
				Token:      []byte{0x14, 0x42},
				Identifier: PUSH_ACK,
			},
			WantError: false,
		},
		{
			Desc:      "PUSH_ACK missing token",
			Header:    []byte{1, PUSH_ACK},
			WantError: true,
		},
		{
			Desc:   "PULL_DATA valid",
			Header: []byte{1, 0x14, 0x42, PULL_DATA, 1, 2, 3, 4, 5, 6, 7, 8},
			WantPacket: Packet{
				Version:    VERSION,
				Token:      []byte{0x14, 0x42},
				Identifier: PULL_DATA,
				GatewayId:  []byte{1, 2, 3, 4, 5, 6, 7, 8},
			},
			WantError: false,
		},
		{
			Desc:   "PULL_RESP with data",
			Header: []byte{1, 0, 0, PULL_RESP},
			JSON:   `{"txpk":{"ipol":true,"powe":12}}`,
			WantPacket: Packet{
				Version:    VERSION,
				Identifier: PULL_RESP,
				Payload: &Payload{
					TXPK: &TXPK{
						Ipol: pointer.Bool(true),
						Powe: pointer.Uint32(12),
					},
				},
			},
			WantError: false,
		},
		{
			Desc:   "PULL_RESP with data and time",
			Header: []byte{1, 0, 0, PULL_RESP},
			JSON:   `{"txpk":{"ipol":true,"powe":12,"time":"2016-01-13T17:40:57.000000376Z"}}`,
			WantPacket: Packet{
				Version:    VERSION,
				Identifier: PULL_RESP,
				Payload: &Payload{
					TXPK: &TXPK{
						Ipol: pointer.Bool(true),
						Powe: pointer.Uint32(12),
						Time: pointer.Time(time.Date(2016, 1, 13, 17, 40, 57, 376, time.UTC)),
					},
				},
			},
			WantError: false,
		},
		{
			Desc:   "PULL_RESP with datr only",
			Header: []byte{1, 0, 0, PULL_RESP},
			JSON:   `{"txpk":{"datr":"SF7BW125"}}`,
			WantPacket: Packet{
				Version:    VERSION,
				Identifier: PULL_RESP,
				Payload: &Payload{
					TXPK: &TXPK{
						Datr: pointer.String("SF7BW125"),
					},
				},
			},
			WantError: false,
		},
		{
			Desc:   "PULL_RESP with time only",
			Header: []byte{1, 0, 0, PULL_RESP},
			JSON:   `{"txpk":{"time":"2016-01-13T17:40:57.000000376Z"}}`,
			WantPacket: Packet{
				Version:    VERSION,
				Identifier: PULL_RESP,
				Payload: &Payload{
					TXPK: &TXPK{
						Time: pointer.Time(time.Date(2016, 1, 13, 17, 40, 57, 376, time.UTC)),
					},
				},
			},
			WantError: false,
		},
		{
			Desc:   "PULL_RESP empty payload",
			Header: []byte{1, 0, 0, PULL_RESP},
			JSON:   `{}`,
			WantPacket: Packet{
				Version:    VERSION,
				Identifier: PULL_RESP,
				Payload:    &Payload{},
			},
			WantError: false,
		},
		{
			Desc:   "PULL_RESP no payload",
			Header: []byte{1, 0, 0, PULL_RESP},
			JSON:   ``,
			WantPacket: Packet{
				Version:    VERSION,
				Identifier: PULL_RESP,
				Payload:    &Payload{},
			},
			WantError: false,
		},
		{
			Desc:   "PULL_ACK valid",
			Header: []byte{1, 0x14, 0x42, PULL_ACK},
			WantPacket: Packet{
				Version:    VERSION,
				Token:      []byte{0x14, 0x42},
				Identifier: PULL_ACK,
			},
			WantError: false,
		},
		{
			Desc:      "Unreckognized version",
			Header:    []byte{VERSION + 14, 1, 2, PUSH_DATA, 1, 4, 5, 6},
			JSON:      `{}`,
			WantError: true,
		},
		{
			Desc:      "Unreckognized Identifier",
			Header:    []byte{VERSION, 1, 2, 178, 1, 4, 5, 6},
			JSON:      `{}`,
			WantError: true,
		},
	}

	for _, test := range tests {
		Desc(t, test.Desc)
		var packet Packet
		err := packet.UnmarshalBinary(append(test.Header, []byte(test.JSON)...))
		checkErrors(t, test.WantError, err)
		if test.WantError {
			continue
		}
		checkPackets(t, test.WantPacket, packet)
	}
}

// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package semtech

import (
	"testing"
	"time"

	"github.com/TheThingsNetwork/ttn/utils/pointer"
	. "github.com/TheThingsNetwork/ttn/utils/testing"
)

func TestMarshalBinary(t *testing.T) {
	tests := []struct {
		Desc       string
		Packet     Packet
		WantError  bool
		WantHeader []byte
		WantJSON   string
	}{
		{
			Desc: "Invalid PUSH_DATA, invalid token",
			Packet: Packet{
				Version: VERSION,
				//No Token
				Identifier: PUSH_DATA,
				GatewayId:  []byte{1, 2, 3, 4, 5, 6, 7, 8},
				Payload: &Payload{
					RXPK: []RXPK{
						RXPK{
							Chan: pointer.Uint(14),
							Codr: pointer.String("4/7"),
							Freq: pointer.Float64(873.14),
							Rssi: pointer.Int(-42),
						},
					},
				},
			},
			WantError: true,
		},
		{
			Desc: "Invalid PUSH_DATA, invalid gateway id",
			Packet: Packet{
				Version:    VERSION,
				Token:      []byte{0x14, 0x42},
				Identifier: PUSH_DATA,
				// No Gateway id
				Payload: &Payload{
					RXPK: []RXPK{
						RXPK{
							Chan: pointer.Uint(14),
							Codr: pointer.String("4/7"),
							Freq: pointer.Float64(873.14),
							Rssi: pointer.Int(-42),
						},
					},
				},
			},
			WantError: true,
		},
		{
			Desc: "PUSH_DATA with no payload",
			Packet: Packet{
				Version:    VERSION,
				Token:      []byte{0x14, 0x42},
				Identifier: PUSH_DATA,
				GatewayId:  []byte{1, 2, 3, 4, 5, 6, 7, 8},
			},
			WantError:  false,
			WantHeader: []byte{1, 0x14, 0x42, PUSH_DATA, 1, 2, 3, 4, 5, 6, 7, 8},
			WantJSON:   `{}`,
		},
		{
			Desc: "PUSH_DATA with only basic typed-attributes  uint, string, float64 and int",
			Packet: Packet{
				Version:    VERSION,
				Token:      []byte{0x14, 0x42},
				Identifier: PUSH_DATA,
				GatewayId:  []byte{1, 2, 3, 4, 5, 6, 7, 8},
				Payload: &Payload{
					RXPK: []RXPK{
						RXPK{
							Chan: pointer.Uint(14),
							Codr: pointer.String("4/7"),
							Freq: pointer.Float64(873.14),
							Rssi: pointer.Int(-42),
						},
					},
				},
			},
			WantError:  false,
			WantHeader: []byte{1, 0x14, 0x42, PUSH_DATA, 1, 2, 3, 4, 5, 6, 7, 8},
			WantJSON:   `{"rxpk":[{"chan":14,"codr":"4/7","freq":873.14,"rssi":-42}]}`,
		},
		{
			Desc: "PUSH_DATA with datr field and modu -> LORA",
			Packet: Packet{
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
			WantError:  false,
			WantHeader: []byte{1, 0x14, 0x42, PUSH_DATA, 1, 2, 3, 4, 5, 6, 7, 8},
			WantJSON:   `{"rxpk":[{"modu":"LORA","datr":"SF7BW125"}]}`,
		},
		{
			Desc: "PUSH_DATA with datr field and modu -> FSK",
			Packet: Packet{
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
			WantError:  false,
			WantHeader: []byte{1, 0x14, 0x42, PUSH_DATA, 1, 2, 3, 4, 5, 6, 7, 8},
			WantJSON:   `{"rxpk":[{"modu":"FSK","datr":50000}]}`,
		},
		{
			Desc: "PUSH_DATA with time field",
			Packet: Packet{
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
			WantError:  false,
			WantHeader: []byte{1, 0x14, 0x42, PUSH_DATA, 1, 2, 3, 4, 5, 6, 7, 8},
			WantJSON:   `{"rxpk":[{"time":"2016-01-13T17:40:57.000000376Z"}]}`,
		},
		{
			Desc: "PUSH_DATA with several RXPK",
			Packet: Packet{
				Version:    VERSION,
				Token:      []byte{0x14, 0x42},
				Identifier: PUSH_DATA,
				GatewayId:  []byte{1, 2, 3, 4, 5, 6, 7, 8},
				Payload: &Payload{
					RXPK: []RXPK{
						RXPK{
							Size: pointer.Uint(14),
						},
						RXPK{
							Chan: pointer.Uint(14),
						},
					},
				},
			},
			WantError:  false,
			WantHeader: []byte{1, 0x14, 0x42, PUSH_DATA, 1, 2, 3, 4, 5, 6, 7, 8},
			WantJSON:   `{"rxpk":[{"size":14},{"chan":14}]}`,
		},
		{
			Desc: "PUSH_DATA with several RXPK and Stat(basic fields)",
			Packet: Packet{
				Version:    VERSION,
				Token:      []byte{0x14, 0x42},
				Identifier: PUSH_DATA,
				GatewayId:  []byte{1, 2, 3, 4, 5, 6, 7, 8},
				Payload: &Payload{
					RXPK: []RXPK{
						RXPK{
							Size: pointer.Uint(14),
						},
					},
					Stat: &Stat{
						Ackr: pointer.Float64(0.78),
						Alti: pointer.Int(72),
						Rxok: pointer.Uint(42),
					},
				},
			},
			WantError:  false,
			WantHeader: []byte{1, 0x14, 0x42, PUSH_DATA, 1, 2, 3, 4, 5, 6, 7, 8},
			WantJSON:   `{"rxpk":[{"size":14}],"stat":{"ackr":0.78,"alti":72,"rxok":42}}`,
		},
		{
			Desc: "PUSH_DATA with Stat(time field)",
			Packet: Packet{
				Version:    VERSION,
				Token:      []byte{0x14, 0x42},
				Identifier: PUSH_DATA,
				GatewayId:  []byte{1, 2, 3, 4, 5, 6, 7, 8},
				Payload: &Payload{
					Stat: &Stat{
						Time: pointer.Time(time.Date(2016, 1, 13, 17, 40, 57, 376, time.UTC)),
					},
				},
			},
			WantError:  false,
			WantHeader: []byte{1, 0x14, 0x42, PUSH_DATA, 1, 2, 3, 4, 5, 6, 7, 8},
			WantJSON:   `{"stat":{"time":"2016-01-13 17:40:57 GMT"}}`,
		},
		{
			Desc: "PUSH_DATA with rxpk and txpk (?)",
			Packet: Packet{
				Version:    VERSION,
				Token:      []byte{0x14, 0x42},
				Identifier: PUSH_DATA,
				GatewayId:  []byte{1, 2, 3, 4, 5, 6, 7, 8},
				Payload: &Payload{
					RXPK: []RXPK{
						RXPK{
							Codr: pointer.String("4/7"),
							Rssi: pointer.Int(-42),
						},
					},
					TXPK: &TXPK{
						Ipol: pointer.Bool(true),
						Powe: pointer.Uint(12),
					},
				},
			},
			WantError:  false,
			WantHeader: []byte{1, 0x14, 0x42, PUSH_DATA, 1, 2, 3, 4, 5, 6, 7, 8},
			WantJSON:   `{"rxpk":[{"codr":"4/7","rssi":-42}],"txpk":{"ipol":true,"powe":12}}`,
		},
		{
			Desc: "PUSH_ACK valid",
			Packet: Packet{
				Version:    VERSION,
				Token:      []byte{0x14, 0x42},
				Identifier: PUSH_ACK,
			},
			WantError:  false,
			WantHeader: []byte{1, 0x14, 0x42, PUSH_ACK},
		},
		{
			Desc: "PUSH_ACK missing token",
			Packet: Packet{
				Version:    VERSION,
				Identifier: PUSH_ACK,
			},
			WantError: true,
		},
		{
			Desc: "PULL_DATA valid",
			Packet: Packet{
				Version:    VERSION,
				Token:      []byte{0x14, 0x42},
				Identifier: PULL_DATA,
				GatewayId:  []byte{1, 2, 3, 4, 5, 6, 7, 8},
			},
			WantError:  false,
			WantHeader: []byte{1, 0x14, 0x42, PULL_DATA, 1, 2, 3, 4, 5, 6, 7, 8},
		},
		{
			Desc: "PULL_DATA missing token",
			Packet: Packet{
				Version:    VERSION,
				Identifier: PULL_DATA,
				GatewayId:  []byte{1, 2, 3, 4, 5, 6, 7, 8},
			},
			WantError: true,
		},
		{
			Desc: "PULL_DATA missing gatewayid",
			Packet: Packet{
				Version:    VERSION,
				Token:      []byte{0x14, 0x42},
				Identifier: PULL_DATA,
			},
			WantError: true,
		},
		{
			Desc: "PULL_RESP with data",
			Packet: Packet{
				Version:    VERSION,
				Identifier: PULL_RESP,
				Payload: &Payload{
					TXPK: &TXPK{
						Ipol: pointer.Bool(true),
						Powe: pointer.Uint(12),
					},
				},
			},
			WantError:  false,
			WantHeader: []byte{1, 0, 0, PULL_RESP},
			WantJSON:   `{"txpk":{"ipol":true,"powe":12}}`,
		},
		{
			Desc: "PULL_RESP with data and time",
			Packet: Packet{
				Version:    VERSION,
				Identifier: PULL_RESP,
				Payload: &Payload{
					TXPK: &TXPK{
						Ipol: pointer.Bool(true),
						Powe: pointer.Uint(12),
						Time: pointer.Time(time.Date(2016, 1, 13, 17, 40, 57, 376, time.UTC)),
					},
				},
			},
			WantError:  false,
			WantHeader: []byte{1, 0, 0, PULL_RESP},
			WantJSON:   `{"txpk":{"ipol":true,"powe":12,"time":"2016-01-13T17:40:57.000000376Z"}}`,
		},
		{
			Desc: "PULL_RESP with time only",
			Packet: Packet{
				Version:    VERSION,
				Identifier: PULL_RESP,
				Payload: &Payload{
					TXPK: &TXPK{
						Time: pointer.Time(time.Date(2016, 1, 13, 17, 40, 57, 376, time.UTC)),
					},
				},
			},
			WantError:  false,
			WantHeader: []byte{1, 0, 0, PULL_RESP},
			WantJSON:   `{"txpk":{"time":"2016-01-13T17:40:57.000000376Z"}}`,
		},
		{
			Desc: "PULL_RESP with datr field and modu -> LORA",
			Packet: Packet{
				Version:    VERSION,
				Identifier: PULL_RESP,
				Payload: &Payload{
					TXPK: &TXPK{
						Datr: pointer.String("SF7BW125"),
						Modu: pointer.String("LORA"),
					},
				},
			},
			WantError:  false,
			WantHeader: []byte{1, 0, 0, PULL_RESP},
			WantJSON:   `{"txpk":{"modu":"LORA","datr":"SF7BW125"}}`,
		},
		{
			Desc: "PULL_RESP with datr field and modu -> FSK",
			Packet: Packet{
				Version:    VERSION,
				Identifier: PULL_RESP,
				Payload: &Payload{
					TXPK: &TXPK{
						Datr: pointer.String("50000"),
						Modu: pointer.String("FSK"),
					},
				},
			},
			WantError:  false,
			WantHeader: []byte{1, 0, 0, PULL_RESP},
			WantJSON:   `{"txpk":{"modu":"FSK","datr":50000}}`,
		},
		{
			Desc: "PULL_RESP empty payload",
			Packet: Packet{
				Version:    VERSION,
				Identifier: PULL_RESP,
				Payload:    &Payload{},
			},
			WantError:  false,
			WantHeader: []byte{1, 0, 0, PULL_RESP},
			WantJSON:   `{}`,
		},
		{
			Desc: "PULL_RESP no payload",
			Packet: Packet{
				Version:    VERSION,
				Identifier: PULL_RESP,
			},
			WantError:  false,
			WantHeader: []byte{1, 0, 0, PULL_RESP},
			WantJSON:   `{}`,
		},
		{
			Desc: "PULL_ACK valid",
			Packet: Packet{
				Version:    VERSION,
				Token:      []byte{0x14, 0x42},
				Identifier: PULL_ACK,
			},
			WantError:  false,
			WantHeader: []byte{1, 0x14, 0x42, PULL_ACK},
		},
		{
			Desc: "PULL_ACK missing token",
			Packet: Packet{
				Version:    VERSION,
				Identifier: PULL_ACK,
			},
			WantError: true,
		},
	}

	for _, test := range tests {
		Desc(t, test.Desc)
		raw, err := test.Packet.MarshalBinary()
		checkErrors(t, test.WantError, err)
		if test.WantError {
			continue
		}
		checkHeaders(t, test.WantHeader, raw)
		checkJSON(t, test.WantJSON, raw)
	}
}

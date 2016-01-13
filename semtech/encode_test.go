// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package semtech

import (
	"github.com/thethingsnetwork/ttn/utils/pointer"
	. "github.com/thethingsnetwork/ttn/utils/testing"
	"reflect"
	"testing"
	"time"
)

func TestMarshalBinary(t *testing.T) {
	tests := []struct {
		Desc       string
		Packet     Packet
		WantError  bool
		WantHeader [12]byte
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
			WantHeader: [12]byte{1, 0x14, 0x42, 0, 1, 2, 3, 4, 5, 6, 7, 8},
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
			WantHeader: [12]byte{1, 0x14, 0x42, 0, 1, 2, 3, 4, 5, 6, 7, 8},
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
			WantHeader: [12]byte{1, 0x14, 0x42, 0, 1, 2, 3, 4, 5, 6, 7, 8},
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
			WantHeader: [12]byte{1, 0x14, 0x42, 0, 1, 2, 3, 4, 5, 6, 7, 8},
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
			WantHeader: [12]byte{1, 0x14, 0x42, 0, 1, 2, 3, 4, 5, 6, 7, 8},
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
			WantHeader: [12]byte{1, 0x14, 0x42, 0, 1, 2, 3, 4, 5, 6, 7, 8},
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
			WantHeader: [12]byte{1, 0x14, 0x42, 0, 1, 2, 3, 4, 5, 6, 7, 8},
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
			WantHeader: [12]byte{1, 0x14, 0x42, 0, 1, 2, 3, 4, 5, 6, 7, 8},
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
			WantHeader: [12]byte{1, 0x14, 0x42, 0, 1, 2, 3, 4, 5, 6, 7, 8},
			WantJSON:   `{"rxpk":[{"codr":"4/7","rssi":-42}],"txpk":{"ipol":true,"powe":12}}`,
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

// ----- Check utilities

func checkErrors(t *testing.T, want bool, got error) {
	if (!want && got == nil) || (want && got != nil) {
		Ok(t, "Check errors")
		return
	}
	Ko(t, "Expected no error but got: %v", got)
}

func checkHeaders(t *testing.T, want [12]byte, got []byte) {
	if len(got) < 12 {
		Ko(t, "Received header does not match expectations.\nWant: %+x\nGot:  %+x", want, got)
		return
	}
	if !reflect.DeepEqual(want[:], got[:12]) {
		Ko(t, "Received header does not match expectations.\nWant: %+x\nGot:  %+x", want, got[:12])
		return
	}
	Ok(t, "Check Headers")
}

func checkJSON(t *testing.T, want string, got []byte) {
	l := len([]byte(want))
	if len(got) < l {
		Ko(t, "Received JSON does not match expectations.\nWant: %s\nGot:  %v", want, got)
		return
	}
	str := string(got[len(got)-l:])
	if want != str {
		Ko(t, "Received JSON does not match expectations.\nWant: %s\nGot:  %s", want, str)
		return
	}
	Ok(t, "Check JSON")
}

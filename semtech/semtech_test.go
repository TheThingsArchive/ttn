// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package semtech

import (
	"fmt"
	"strings"
	"testing"

	"github.com/TheThingsNetwork/ttn/utils/pointer"
	testutil "github.com/TheThingsNetwork/ttn/utils/testing"
)

func TestString(t *testing.T) {
	{
		testutil.Desc(t, "No Payload")

		packet := Packet{
			Version:    VERSION,
			Token:      []byte{1, 2},
			Identifier: PUSH_DATA,
			GatewayId:  []byte{1, 2, 3, 4, 5, 6, 7, 8},
		}

		str := packet.String()
		CheckStrings(t, "Packet", str)
		CheckStrings(t, fmt.Sprintf("Version:%v", VERSION), str)
		CheckStrings(t, fmt.Sprintf("Token:%v", packet.Token), str)
		CheckStrings(t, fmt.Sprintf("Identifier:%v", PUSH_DATA), str)
		CheckStrings(t, fmt.Sprintf("GatewayId:%v", packet.GatewayId), str)
	}

	// --------------------

	{
		testutil.Desc(t, "With Stat Payload")

		packet := Packet{
			Version:    VERSION,
			Token:      []byte{1, 2},
			Identifier: PUSH_DATA,
			GatewayId:  []byte{1, 2, 3, 4, 5, 6, 7, 8},
			Payload: &Payload{
				Stat: &Stat{
					Ackr: pointer.Float32(3.4),
					Alti: pointer.Int32(14),
				},
			},
		}

		str := packet.String()
		CheckStrings(t, "Packet", str)
		CheckStrings(t, fmt.Sprintf("Version:%v", VERSION), str)
		CheckStrings(t, fmt.Sprintf("Token:%v", packet.Token), str)
		CheckStrings(t, fmt.Sprintf("Identifier:%v", PUSH_DATA), str)
		CheckStrings(t, fmt.Sprintf("GatewayId:%v", packet.GatewayId), str)
		CheckStrings(t, "Payload", str)
		CheckStrings(t, "Stat", str)
		CheckStrings(t, "Ackr", str)
		CheckStrings(t, "Alti", str)
	}

	// --------------------

	{
		testutil.Desc(t, "With TXPK Payload")

		packet := Packet{
			Version:    VERSION,
			Token:      []byte{1, 2},
			Identifier: PUSH_DATA,
			GatewayId:  []byte{1, 2, 3, 4, 5, 6, 7, 8},
			Payload: &Payload{
				TXPK: &TXPK{
					Freq: pointer.Float32(883.445),
					Codr: pointer.String("4/5"),
				},
			},
		}

		str := packet.String()
		CheckStrings(t, "Packet", str)
		CheckStrings(t, fmt.Sprintf("Version:%v", VERSION), str)
		CheckStrings(t, fmt.Sprintf("Token:%v", packet.Token), str)
		CheckStrings(t, fmt.Sprintf("Identifier:%v", PUSH_DATA), str)
		CheckStrings(t, fmt.Sprintf("GatewayId:%v", packet.GatewayId), str)
		CheckStrings(t, "Payload", str)
		CheckStrings(t, "TXPK", str)
		CheckStrings(t, "Codr", str)
		CheckStrings(t, "Freq", str)
	}

	// --------------------

	{
		testutil.Desc(t, "With RXPK Payloads")

		packet := Packet{
			Version:    VERSION,
			Token:      []byte{1, 2},
			Identifier: PUSH_DATA,
			GatewayId:  []byte{1, 2, 3, 4, 5, 6, 7, 8},
			Payload: &Payload{
				RXPK: []RXPK{
					{
						Freq: pointer.Float32(883.445),
						Codr: pointer.String("4/5"),
					},
					{
						Freq: pointer.Float32(883.445),
						Codr: pointer.String("4/5"),
					},
				},
			},
		}

		str := packet.String()
		CheckStrings(t, "Packet", str)
		CheckStrings(t, fmt.Sprintf("Version:%v", VERSION), str)
		CheckStrings(t, fmt.Sprintf("Token:%v", packet.Token), str)
		CheckStrings(t, fmt.Sprintf("Identifier:%v", PUSH_DATA), str)
		CheckStrings(t, fmt.Sprintf("GatewayId:%v", packet.GatewayId), str)
		CheckStrings(t, "Payload", str)
		CheckStrings(t, "RXPK", str)
		CheckStrings(t, "Codr", str)
		CheckStrings(t, "Freq", str)
	}

	// --------------------

	{
		testutil.Desc(t, "Nil payload")

		var packet *Packet
		str := packet.String()
		CheckStrings(t, "nil", str)
	}
}

func CheckStrings(t *testing.T, want string, got string) {
	if !strings.Contains(got, want) {
		testutil.Ko(t, "Expected %s to contain \"%s\"", got, want)
	}
	testutil.Ok(t, "Check Strings")
}

// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package components

import (
	"encoding/json"
	"github.com/thethingsnetwork/core/utils/pointer"
	. "github.com/thethingsnetwork/core/utils/testing"
	"testing"
	"time"
)

// The broker can handle an uplink packet
func TestMarshaljson(t *testing.T) {
	tests := []struct {
		Metadata  Metadata
		WantError error
		WantJSON  string
	}{
		{ // Basic attributes, uint, string and float64
			Metadata{Chan: pointer.Uint(2), Codr: pointer.String("4/6"), Freq: pointer.Float64(864.125)},
			nil,
			`{"chan":2,"codr":"4/6","freq":864.125}`,
		},

		{ // Basic attributes #2, int and bool
			Metadata{Imme: pointer.Bool(true), Rssi: pointer.Int(-54)},
			nil,
			`{"imme":true,"rssi":-54}`,
		},

		{ // Datr attr, FSK type
			Metadata{Datr: pointer.String("50000"), Modu: pointer.String("FSK")},
			nil,
			`{"modu":"FSK","datr":50000}`,
		},

		{ // Datr attr, lora modulation
			Metadata{Datr: pointer.String("SF7BW125"), Modu: pointer.String("LORA")},
			nil,
			`{"modu":"LORA","datr":"SF7BW125"}`,
		},

		{ // Time attr
			Metadata{Time: pointer.Time(time.Date(2016, 1, 6, 15, 11, 12, 142, time.UTC))},
			nil,
			`{"time":"2016-01-06T15:11:12.000000142Z"}`,
		},

		{ // Mixed
			Metadata{
				Time: pointer.Time(time.Date(2016, 1, 6, 15, 11, 12, 142, time.UTC)),
				Modu: pointer.String("FSK"),
				Datr: pointer.String("50000"),
				Size: pointer.Uint(14),
				Lsnr: pointer.Float64(5.7),
			},
			nil,
			`{"lsnr":5.7,"modu":"FSK","size":14,"datr":50000,"time":"2016-01-06T15:11:12.000000142Z"}`,
		},
	}

	for _, test := range tests {
		Desc(t, "Marshal medatadata: %v", pointer.DumpPStruct(test.Metadata, false))
		raw, err := json.Marshal(test.Metadata)

		if err != test.WantError {
			Ko(t, "Expected error to be %v but got %v", test.WantError, err)
			continue
		}

		str := string(raw)
		if str != test.WantJSON {
			Ko(t, "Marshaled data don't match expectation.\nWant: %s\nGot:  %s", test.WantJSON, str)
			continue
		}
		Ok(t)
	}
}

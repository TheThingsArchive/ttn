// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package core

import (
	"encoding/json"
	"testing"
	"time"

	. "github.com/TheThingsNetwork/ttn/utils/errors/checks"
	"github.com/TheThingsNetwork/ttn/utils/pointer"
	. "github.com/TheThingsNetwork/ttn/utils/testing"
)

var commonTests = []struct {
	Metadata  Metadata
	WantError *string
	JSON      string
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

var unmarshalTests = []struct {
	Metadata  Metadata
	WantError *string
	JSON      string
}{
	{ // Local time
		Metadata{Time: pointer.Time(time.Date(2016, 1, 6, 15, 11, 12, 0, time.UTC))},
		nil,
		`{"time":"2016-01-06 15:11:12 GMT"}`,
	},

	{ // RFC3339 time
		Metadata{Time: pointer.Time(time.Date(2016, 1, 6, 15, 11, 12, 142000000, time.UTC))},
		nil,
		`{"time":"2016-01-06T15:11:12.142000Z"}`,
	},
}

func TestMarshaljson(t *testing.T) {
	for _, test := range commonTests {
		Desc(t, "Marshal medatadata: %s", test.Metadata.String())
		raw, err := json.Marshal(test.Metadata)
		CheckErrors(t, test.WantError, err)
		checkJSON(t, test.JSON, raw)
	}
}

func TestUnmarshalJSON(t *testing.T) {
	for _, test := range append(commonTests, unmarshalTests...) {
		Desc(t, "Unmarshal json: %s", test.JSON)
		metadata := Metadata{}
		err := json.Unmarshal([]byte(test.JSON), &metadata)
		CheckErrors(t, test.WantError, err)
		checkMetadata(t, test.Metadata, metadata)
	}
}

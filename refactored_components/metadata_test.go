// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package components

import (
	"encoding/json"
	"github.com/thethingsnetwork/core/utils/pointer"
	. "github.com/thethingsnetwork/core/utils/testing"
	"reflect"
	"testing"
	"time"
)

var commonTests = []struct {
	Metadata  Metadata
	WantError error
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
	WantError error
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

// The broker can handle an uplink packet
func TestMarshaljson(t *testing.T) {
	for _, test := range commonTests {
		Desc(t, "Marshal medatadata: %v", pointer.DumpPStruct(test.Metadata, false))
		raw, err := json.Marshal(test.Metadata)
		checkErrors(t, test.WantError, err)
		checkJSON(t, test.JSON, raw)
	}
}

func TestUnmarshalJSON(t *testing.T) {
	for _, test := range append(commonTests, unmarshalTests...) {
		Desc(t, "Unmarshal json: %s", test.JSON)
		metadata := Metadata{}
		err := json.Unmarshal([]byte(test.JSON), &metadata)
		checkErrors(t, test.WantError, err)
		checkMetadata(t, test.Metadata, metadata)
	}
}

// ----- Check utilities

// Check that errors match
func checkErrors(t *testing.T, want error, got error) {
	if got == want {
		Ok(t, "check Errors")
		return
	}
	Ko(t, "Expected error to be %v but got %v", want, got)
}

// Check that obtained json matches expected one
func checkJSON(t *testing.T, want string, got []byte) {
	str := string(got)
	if str == want {
		Ok(t, "check JSON")
		return
	}
	Ko(t, "Marshaled data don't match expectations.\nWant: %s\nGot:  %s", want, str)
	return
}

// Check that obtained metadata matches expected one
func checkMetadata(t *testing.T, want Metadata, got Metadata) {
	if reflect.DeepEqual(want, got) {
		Ok(t, "check Metadata")
		return
	}
	Ko(t, "Unmarshaled json don't match expectations. \nWant: %s\nGot:  %s", pointer.DumpPStruct(want, false), pointer.DumpPStruct(got, false))
}

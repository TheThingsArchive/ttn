// Copyright © 2015 Matthias Benkort <matthias.benkort@gmail.com>
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package protocol

import (
	"encoding/json"
	"errors"
	"strings"
	"time"
)

// Unmarshal parse a raw response from a server and turn in into a packet.
// Will return an error if the response fields are incorrect.
func Unmarshal(raw []byte) (*Packet, error) {
	size := len(raw)

	if size < 3 {
		return nil, errors.New("Invalid raw data format")
	}

	packet := &Packet{
		Version:    raw[0],
		Token:      raw[1:3],
		Identifier: raw[3],
		GatewayId:  nil,
		Payload:    nil,
	}

	if packet.Version != VERSION {
		return nil, errors.New("Unreckognized protocol version")
	}

	if packet.Identifier > PULL_ACK {
		return nil, errors.New("Unreckognized protocol identifier")
	}

	cursor := 4
	if packet.Identifier == PULL_DATA || packet.Identifier == PUSH_DATA {
		if size < 12 {
			return nil, errors.New("Invalid gateway identifier")
		}
		packet.GatewayId = raw[cursor:12]
		cursor = 12
	}

	var err error
	if size > cursor && (packet.Identifier == PUSH_DATA || packet.Identifier == PULL_RESP) {
		packet.Payload, err = unmarshalPayload(raw[cursor:])
	}

	return packet, err
}

// timeParser is used as a proxy to Unmarshal JSON objects with different date types as the time
// module parse RFC3339 by default
type timeParser struct {
	Value  time.Time    // The parsed time value
	Parsed bool         // Set to true if the value has been parsed
}

// implement the Unmarshaller interface from encoding/json
func (t *timeParser) UnmarshalJSON(raw []byte) error {
	var err error
	value := strings.Trim(string(raw), `"`)
	t.Value, err = time.Parse("2006-01-02 15:04:05 GMT", value)
	if err != nil {
		t.Value, err = time.Parse(time.RFC3339, value)
	}
	if err != nil {
		t.Value, err = time.Parse(time.RFC3339Nano, value)
	}
	if err != nil {
		return errors.New("Unkown date format. Unable to parse time")
	}

	t.Parsed = true
	return nil
}

// datrParser is used as a proxy to Unmarshal datr field in json payloads.
// Depending on the modulation type, the datr type could be either a string or a number.
// We're gonna parse it as a string in any case.
type datrParser struct {
	Value  string   // The parsed value
	Parsed bool     // Set to true if the value has been parsed
}

// implement the Unmarshaller interface from encoding/json
func (d *datrParser) UnmarshalJSON(raw []byte) error {
	d.Value = strings.Trim(string(raw), `"`)

	if d.Value == "" {
		return errors.New("Invalid datr format")
	}

	d.Parsed = true
	return nil
}

// unmarshalPayload is an until used by Unmarshal to parse a Payload from a sequence of bytes.
func unmarshalPayload(raw []byte) (*Payload, error) {
	payload := &Payload{raw, nil, nil, nil}
	customStruct := &struct {
		Stat *struct {
			Time timeParser `json:"time"`
		} `json:"stat"`
		RXPK *[]struct {
			Time timeParser `json:"time"`
			Datr datrParser `json:"datr"`
		} `json:"rxpk"`
		TXPK *struct {
			Time timeParser `json:"time"`
			Datr datrParser `json:"datr"`
		} `json:"txpk"`
	}{}

	err := json.Unmarshal(raw, payload)
	err = json.Unmarshal(raw, customStruct)

	if err != nil {
		return nil, err
	}

	if customStruct.Stat != nil && customStruct.Stat.Time.Parsed {
		payload.Stat.Time = customStruct.Stat.Time.Value
	}

	if customStruct.RXPK != nil {
		for i, x := range *customStruct.RXPK {
			if x.Time.Parsed {
				(*payload.RXPK)[i].Time = x.Time.Value
			}

			if x.Datr.Parsed {
				(*payload.RXPK)[i].Datr = x.Datr.Value
			}
		}
	}

	if customStruct.TXPK != nil {
		if customStruct.TXPK.Time.Parsed {
			payload.TXPK.Time = customStruct.TXPK.Time.Value
		}

		if customStruct.TXPK.Datr.Parsed {
			payload.TXPK.Datr = customStruct.TXPK.Datr.Value
		}
	}

	return payload, nil
}

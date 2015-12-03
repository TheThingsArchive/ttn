// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package gateway

import (
	"encoding/json"
	"errors"
	"strings"
	"time"
)

// Unmarshal parses a raw response from a server and turn in into a packet.
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
	Value *time.Time // The parsed time value
}

// UnmarshalJSON implements the Unmarshaler interface from encoding/json
func (t *timeParser) UnmarshalJSON(raw []byte) error {
	str := strings.Trim(string(raw), `"`)
	v, err := time.Parse("2006-01-02 15:04:05 GMT", str)
	if err != nil {
		v, err = time.Parse(time.RFC3339, str)
	}
	if err != nil {
		v, err = time.Parse(time.RFC3339Nano, str)
	}
	if err != nil {
		return errors.New("Unkown date format. Unable to parse time")
	}
	t.Value = &v
	return nil
}

// datrParser is used as a proxy to Unmarshal datr field in json payloads.
// Depending on the modulation type, the datr type could be either a string or a number.
// We're gonna parse it as a string in any case.
type datrParser struct {
	Value *string // The parsed value
}

// UnmarshalJSON implements the Unmarshaler interface from encoding/json
func (d *datrParser) UnmarshalJSON(raw []byte) error {
	v := strings.Trim(string(raw), `"`)

	if v == "" {
		return errors.New("Invalid datr format")
	}

	d.Value = &v
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

	if customStruct.Stat != nil {
		payload.Stat.Time = customStruct.Stat.Time.Value
	}

	if customStruct.RXPK != nil {
		for i, x := range *customStruct.RXPK {
			(*payload.RXPK)[i].Time = x.Time.Value
			(*payload.RXPK)[i].Datr = x.Datr.Value
		}
	}

	if customStruct.TXPK != nil {
		payload.TXPK.Time = customStruct.TXPK.Time.Value
		payload.TXPK.Datr = customStruct.TXPK.Datr.Value
	}

	return payload, nil
}

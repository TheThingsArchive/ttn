// Copyright © 2015 Matthias Benkort <matthias.benkort@gmail.com>
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package protocol

import (
	"encoding/json"
	"errors"
	"strings"
	"time"
)

type timeParser struct {
	Value  time.Time
	Parsed bool
}

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

type datrParser struct {
	Value  string
	Parsed bool
}

func (d *datrParser) UnmarshalJSON(raw []byte) error {
	d.Value = strings.Trim(string(raw), `"`)

	if d.Value == "" {
		return errors.New("Invalid datr format")
	}

	d.Parsed = true
	return nil
}

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

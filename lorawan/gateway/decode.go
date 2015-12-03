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
		packet.Payload = new(Payload)
		packet.Payload.Raw = raw[cursor:]
		err = json.Unmarshal(raw[cursor:], packet.Payload)
	}

	return packet, err
}

// timeParser is used as a proxy to Unmarshal JSON objects with different date types as the time
// module parse RFC3339 by default
type timeparser struct {
	Value *time.Time // The parsed time value
}

// UnmarshalJSON implements the Unmarshaler interface from encoding/json
func (t *timeparser) UnmarshalJSON(raw []byte) error {
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
type datrparser struct {
	Value *string // The parsed value
}

// UnmarshalJSON implements the Unmarshaler interface from encoding/json
func (d *datrparser) UnmarshalJSON(raw []byte) error {
	v := strings.Trim(string(raw), `"`)

	if v == "" {
		return errors.New("Invalid datr format")
	}

	d.Value = &v
	return nil
}

// UnmarshalJSON implements the Unmarshaler interface from encoding/json
func (p *Payload) UnmarshalJSON(raw []byte) error {
    proxy := new(struct {
        Stat struct{
            *Stat
            Time *timeparser `json:"time"`
        }
        RXPK []struct{
            *RXPK
            Time *timeparser `json:"time"`
            Datr *datrparser `json:"datr"`
        }
        TXPK struct{
            *TXPK
            Time *timeparser `json:"time"`
            Datr *datrparser `json:"datr"`
        }
    })

    stat := new(Stat)
    txpk := new(TXPK)
    proxy.Stat.Stat = stat
    proxy.TXPK.TXPK = txpk

    if err := json.Unmarshal(raw, proxy); err != nil {
        return err
    }

    if proxy.Stat.Stat != nil {
        if proxy.Stat.Time != nil {
            stat.Time = proxy.Stat.Time.Value
        }
        p.Stat = stat
    }

    if proxy.TXPK.TXPK != nil {
        if proxy.TXPK.Time != nil {
            txpk.Time = proxy.TXPK.Time.Value
        }
        if proxy.TXPK.Datr != nil {
            txpk.Datr = proxy.TXPK.Datr.Value
        }
        p.TXPK = txpk
    }

    for _, rxpk := range(proxy.RXPK) {
        if rxpk.Time != nil {
            rxpk.RXPK.Time = rxpk.Time.Value
        }
        if rxpk.Datr != nil {
            rxpk.RXPK.Datr = rxpk.Datr.Value
        }
        p.RXPK = append(p.RXPK, *rxpk.RXPK)
    }

    return nil
}

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
func (r *RXPK) UnmarshalJSON(raw []byte) error {
	proxy := new(struct {
		Chan *uint       `json:"chan"`
		Codr *string     `json:"codr"`
		Data *string     `json:"data"`
		Datr *datrparser `json:"datr"`
		Freq *float64    `json:"freq"`
		Lsnr *float64    `json:"lsnr"`
		Modu *string     `json:"modu"`
		Rfch *uint       `json:"rfch"`
		Rssi *int        `json:"rssi"`
		Size *uint       `json:"size"`
		Stat *int        `json:"stat"`
		Time *timeparser `json:"time"`
		Tmst *uint       `json:"tmst"`
	})
	if err := json.Unmarshal(raw, proxy); err != nil {
		return err
	}

	r.Chan = proxy.Chan
	r.Codr = proxy.Codr
	r.Data = proxy.Data
	r.Freq = proxy.Freq
	r.Lsnr = proxy.Lsnr
	r.Modu = proxy.Modu
	r.Rfch = proxy.Rfch
	r.Rssi = proxy.Rssi
	r.Size = proxy.Size
	r.Stat = proxy.Stat
	r.Tmst = proxy.Tmst

	if proxy.Datr != nil {
		r.Datr = proxy.Datr.Value
	}

	if proxy.Time != nil {
		r.Time = proxy.Time.Value
	}
	return nil
}

// UnmarshalJSON implements the Unmarshaler interface from encoding/json
func (s *Stat) UnmarshalJSON(raw []byte) error {
	proxy := new(struct {
		Ackr *float64    `json:"ackr"`
		Alti *int        `json:"alti"`
		Dwnb *uint       `json:"dwnb"`
		Lati *float64    `json:"lati"`
		Long *float64    `json:"long"`
		Rxfw *uint       `json:"rxfw"`
		Rxnb *uint       `json:"rxnb"`
		Rxok *uint       `json:"rxok"`
		Time *timeparser `json:"time"`
		Txnb *uint       `json:"txnb"`
	})

	if err := json.Unmarshal(raw, proxy); err != nil {
		return err
	}

	s.Ackr = proxy.Ackr
	s.Alti = proxy.Alti
	s.Dwnb = proxy.Dwnb
	s.Lati = proxy.Lati
	s.Long = proxy.Long
	s.Rxfw = proxy.Rxfw
	s.Rxnb = proxy.Rxnb
	s.Rxok = proxy.Rxok
	s.Txnb = proxy.Txnb

	if proxy.Time != nil {
		s.Time = proxy.Time.Value
	}

	return nil
}

func (t *TXPK) UnmarshalJSON(raw []byte) error {
	proxy := new(struct {
		Codr *string     `json:"codr"`
		Data *string     `json:"data"`
		Datr *datrparser `json:"datr"`
		Fdev *uint       `json:"fdev"`
		Freq *float64    `json:"freq"`
		Imme *bool       `json:"imme"`
		Ipol *bool       `json:"ipol"`
		Modu *string     `json:"modu"`
		Ncrc *bool       `json:"ncrc"`
		Powe *uint       `json:"powe"`
		Prea *uint       `json:"prea"`
		Rfch *uint       `json:"rfch"`
		Size *uint       `json:"size"`
		Time *timeparser `json:"time"`
		Tmst *uint       `json:"tmst"`
	})

	if err := json.Unmarshal(raw, proxy); err != nil {
		return err
	}

	t.Codr = proxy.Codr
	t.Data = proxy.Data
	t.Fdev = proxy.Fdev
	t.Freq = proxy.Freq
	t.Imme = proxy.Imme
	t.Ipol = proxy.Ipol
	t.Modu = proxy.Modu
	t.Ncrc = proxy.Ncrc
	t.Powe = proxy.Powe
	t.Prea = proxy.Prea
	t.Rfch = proxy.Rfch
	t.Size = proxy.Size
	t.Tmst = proxy.Tmst

	if proxy.Datr != nil {
		t.Datr = proxy.Datr.Value
	}

	if proxy.Time != nil {
		t.Time = proxy.Time.Value
	}
	return nil
}

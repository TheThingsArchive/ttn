// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package gateway

import (
	"encoding/json"
	"errors"
	"time"
)

// Marshal transforms a packet to a sequence of bytes.
func Marshal(packet *Packet) ([]byte, error) {
	raw := append(make([]byte, 0), packet.Version)

	if len(packet.Token) != 2 {
		return nil, errors.New("Invalid packet token")
	}

	raw = append(raw, packet.Token...)
	raw = append(raw, packet.Identifier)

	if packet.Identifier == PUSH_DATA || packet.Identifier == PULL_DATA {
		if len(packet.GatewayId) != 8 {
			return nil, errors.New("Invalid packet gatewayId")
		}
		raw = append(raw, packet.GatewayId...)
	}

	if packet.Payload != nil && (packet.Identifier == PUSH_DATA || packet.Identifier == PULL_RESP) {
		payload, err := json.Marshal(packet.Payload)
		if err != nil {
			return nil, err
		}
		raw = append(raw, payload...)
	}

	return raw, nil
}

type timemarshaler struct {
	layout string
	value  time.Time
}

func (t *timemarshaler) MarshalJSON() ([]byte, error) {
	return append(append([]byte(`"`), []byte(t.value.Format(t.layout))...), []byte(`"`)...), nil
}

type datrmarshaler struct {
	kind  string
	value string
}

func (d *datrmarshaler) MarshalJSON() ([]byte, error) {
	if d.kind == "uint" {
		return []byte(d.value), nil
	}
	return append(append([]byte(`"`), []byte(d.value)...), []byte(`"`)...), nil
}

func (r *RXPK) MarshalJSON() ([]byte, error) {
	var rfctime *timemarshaler = nil
	var datr *datrmarshaler = nil

	if r.Time != nil {
		rfctime = &timemarshaler{time.RFC3339Nano, *r.Time}
	}

	if r.Modu != nil && r.Datr != nil {
		switch *r.Modu {
		case "FSK":
			datr = &datrmarshaler{"uint", *r.Datr}
		case "LORA":
			fallthrough
		default:
			datr = &datrmarshaler{"string", *r.Datr}
		}
	}

	return json.Marshal(struct {
		Chan *uint          `json:"chan,omitempty"`
		Codr *string        `json:"codr,omitempty"`
		Data *string        `json:"data,omitempty"`
		Datr *datrmarshaler `json:"datr,omitempty"`
		Freq *float64       `json:"freq,omitempty"`
		Lsnr *float64       `json:"lsnr,omitempty"`
		Modu *string        `json:"modu,omitempty"`
		Rfch *uint          `json:"rfch,omitempty"`
		Rssi *int           `json:"rssi,omitempty"`
		Size *uint          `json:"size,omitempty"`
		Stat *int           `json:"stat,omitempty"`
		Time *timemarshaler `json:"time,omitempty"`
		Tmst *uint          `json:"tmst,omitempty"`
	}{
		r.Chan,
		r.Codr,
		r.Data,
		datr,
		r.Freq,
		r.Lsnr,
		r.Modu,
		r.Rfch,
		r.Rssi,
		r.Size,
		r.Stat,
		rfctime,
		r.Tmst,
	})
}

func (s *Stat) MarshalJSON() ([]byte, error) {
	var rfctime *timemarshaler = nil
	if s.Time != nil {
		rfctime = &timemarshaler{"2006-01-02 15:04:05 GMT", *s.Time}
	}

	return json.Marshal(struct {
		Ackr *float64       `json:"ackr,omitempty"`
		Alti *int           `json:"alti,omitempty"`
		Dwnb *uint          `json:"dwnb,omitempty"`
		Lati *float64       `json:"lati,omitempty"`
		Long *float64       `json:"long,omitempty"`
		Rxfw *uint          `json:"rxfw,omitempty"`
		Rxnb *uint          `json:"rxnb,omitempty"`
		Rxok *uint          `json:"rxok,omitempty"`
		Time *timemarshaler `json:"time,omitempty"`
		Txnb *uint          `json:"txnb,omitempty"`
	}{
		s.Ackr,
		s.Alti,
		s.Dwnb,
		s.Lati,
		s.Long,
		s.Rxfw,
		s.Rxnb,
		s.Rxok,
		rfctime,
		s.Txnb,
	})
}

func (t *TXPK) MarshalJSON() ([]byte, error) {
	var rfctime *timemarshaler = nil
	var datr *datrmarshaler = nil

	if t.Time != nil {
		rfctime = &timemarshaler{time.RFC3339Nano, *t.Time}
	}

	if t.Modu != nil && t.Datr != nil {
		switch *t.Modu {
		case "FSK":
			datr = &datrmarshaler{"uint", *t.Datr}
		case "LORA":
			fallthrough
		default:
			datr = &datrmarshaler{"string", *t.Datr}
		}
	}

	return json.Marshal(struct {
		Codr *string        `json:"codr,omitempty"`
		Data *string        `json:"data,omitempty"`
		Datr *datrmarshaler `json:"datr,omitempty"`
		Fdev *uint          `json:"fdev,omitempty"`
		Freq *float64       `json:"freq,omitempty"`
		Imme *bool          `json:"imme,omitempty"`
		Ipol *bool          `json:"ipol,omitempty"`
		Modu *string        `json:"modu,omitempty"`
		Ncrc *bool          `json:"ncrc,omitempty"`
		Powe *uint          `json:"powe,omitempty"`
		Prea *uint          `json:"prea,omitempty"`
		Rfch *uint          `json:"rfch,omitempty"`
		Size *uint          `json:"size,omitempty"`
		Time *timemarshaler `json:"time,omitempty"`
		Tmst *uint          `json:"tmst,omitempty"`
	}{
		t.Codr,
		t.Data,
		datr,
		t.Fdev,
		t.Freq,
		t.Imme,
		t.Ipol,
		t.Modu,
		t.Ncrc,
		t.Powe,
		t.Prea,
		t.Rfch,
		t.Size,
		rfctime,
		t.Tmst,
	})
}

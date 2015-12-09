// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package semtech

import (
	"encoding/json"
	"errors"
	"time"
)

// Marshal transforms a packet to a sequence of bytes.
func Marshal(packet Packet) ([]byte, error) {
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

// MarshalJSON implements the Marshaler interface from encoding/json
func (t *timeparser) MarshalJSON() ([]byte, error) {
	if t.value == nil {
		return nil, errors.New("Cannot marshal a null time")
	}
	return append(append([]byte(`"`), []byte(t.value.Format(t.layout))...), []byte(`"`)...), nil
}

// MarshalJSON implements the Marshaler interface from encoding/json
func (d *datrparser) MarshalJSON() ([]byte, error) {
	if d.value == nil {
		return nil, errors.New("Cannot marshal a null datr")
	}

	if d.kind == "uint" {
		return []byte(*d.value), nil
	}
	return append(append([]byte(`"`), []byte(*d.value)...), []byte(`"`)...), nil
}

// MarshalJSON implements the Marshaler interface from encoding/json
func (p *Payload) MarshalJSON() ([]byte, error) {
	// Define Stat Proxy
	var proxStat *statProxy
	if p.Stat != nil {
		proxStat = new(statProxy)
		proxStat.Stat = p.Stat
		if p.Stat.Time != nil {
			proxStat.Time = &timeparser{layout: "2006-01-02 15:04:05 GMT", value: p.Stat.Time}
		}
	}

	// Define RXPK Proxy
	proxRXPK := make([]rxpkProxy, 0)
	for _, rxpk := range p.RXPK {
		proxr := new(rxpkProxy)
		proxr.RXPK = new(RXPK)
		*proxr.RXPK = rxpk
		if rxpk.Time != nil {
			proxr.Time = &timeparser{time.RFC3339Nano, rxpk.Time}
		}

		if rxpk.Modu != nil && rxpk.Datr != nil {
			switch *rxpk.Modu {
			case "FSK":
				proxr.Datr = &datrparser{kind: "uint", value: rxpk.Datr}
			case "LORA":
				fallthrough
			default:
				proxr.Datr = &datrparser{kind: "string", value: rxpk.Datr}
			}
		}
		proxRXPK = append(proxRXPK, *proxr)
	}

	// Define TXPK Proxy
	var proxTXPK *txpkProxy
	if p.TXPK != nil {
		proxTXPK = new(txpkProxy)
		proxTXPK.TXPK = p.TXPK
		if p.TXPK.Time != nil {
			proxTXPK.Time = &timeparser{time.RFC3339Nano, p.TXPK.Time}
		}
		if p.TXPK.Modu != nil && p.TXPK.Datr != nil {
			switch *p.TXPK.Modu {
			case "FSK":
				proxTXPK.Datr = &datrparser{kind: "uint", value: p.TXPK.Datr}
			case "LORA":
				fallthrough
			default:
				proxTXPK.Datr = &datrparser{kind: "string", value: p.TXPK.Datr}
			}
		}
	}

	// Define the whole Proxy
	proxy := payloadProxy{
		ProxStat: proxStat,
		ProxRXPK: proxRXPK,
		ProxTXPK: proxTXPK,
	}

	raw, err := json.Marshal(proxy)
	return raw, err
}

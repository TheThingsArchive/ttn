// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package semtech

import (
	"encoding/json"
	"errors"
	"time"
)

// Marshal transforms a packet to a sequence of bytes.
func (packet Packet) MarshalBinary() ([]byte, error) {
	raw := append(make([]byte, 0), packet.Version)

	if packet.Identifier == PULL_RESP {
		packet.Token = []byte{0x0, 0x0}
	}

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

	if packet.Identifier == PUSH_DATA || packet.Identifier == PULL_RESP {
		if packet.Payload != nil {
			payload, err := json.Marshal(packet.Payload)
			if err != nil {
				return nil, err
			}
			raw = append(raw, payload...)
		} else {
			raw = append(raw, []byte("{}")...)
		}
	}

	return raw, nil
}

// MarshalJSON implements the Marshaler interface from encoding/json
func (t *Timeparser) MarshalJSON() ([]byte, error) {
	if t.Value == nil {
		return nil, errors.New("Cannot marshal a null time")
	}
	return append(append([]byte(`"`), []byte(t.Value.Format(t.Layout))...), []byte(`"`)...), nil
}

// MarshalJSON implements the Marshaler interface from encoding/json
func (d *Datrparser) MarshalJSON() ([]byte, error) {
	if d.Kind == "uint" {
		return []byte(d.Value), nil
	}
	return append(append([]byte(`"`), []byte(d.Value)...), []byte(`"`)...), nil
}

// MarshalJSON implements the Marshaler interface from encoding/json
func (p *Payload) MarshalJSON() ([]byte, error) {
	// Define Stat Proxy
	var proxStat *statProxy
	if p.Stat != nil {
		proxStat = new(statProxy)
		proxStat.Stat = p.Stat
		if p.Stat.Time != nil {
			proxStat.Time = &Timeparser{Layout: "2006-01-02 15:04:05 GMT", Value: p.Stat.Time}
		}
	}

	// Define RXPK Proxy
	proxRXPK := make([]rxpkProxy, 0)
	for _, rxpk := range p.RXPK {
		proxr := new(rxpkProxy)
		proxr.RXPK = new(RXPK)
		*proxr.RXPK = rxpk
		if rxpk.Time != nil {
			proxr.Time = &Timeparser{time.RFC3339Nano, rxpk.Time}
		}

		if rxpk.Modu != nil && rxpk.Datr != nil {
			switch *rxpk.Modu {
			case "FSK":
				proxr.Datr = &Datrparser{Kind: "uint", Value: *rxpk.Datr}
			case "LORA":
				fallthrough
			default:
				proxr.Datr = &Datrparser{Kind: "string", Value: *rxpk.Datr}
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
			proxTXPK.Time = &Timeparser{time.RFC3339Nano, p.TXPK.Time}
		}
		if p.TXPK.Modu != nil && p.TXPK.Datr != nil {
			switch *p.TXPK.Modu {
			case "FSK":
				proxTXPK.Datr = &Datrparser{Kind: "uint", Value: *p.TXPK.Datr}
			case "LORA":
				fallthrough
			default:
				proxTXPK.Datr = &Datrparser{Kind: "string", Value: *p.TXPK.Datr}
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

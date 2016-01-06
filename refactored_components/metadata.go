// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package components

import (
	"encoding/json"
	"github.com/thethingsnetwork/core/semtech"
	"time"
)

type Metadata struct {
	Chan *uint      `json:"chan,omitempty"` // Concentrator "IF" channel used for RX (unsigned integer)
	Codr *string    `json:"codr,omitempty"` // LoRa ECC coding rate identifier
	Data *string    `json:"data,omitempty"` // Base64 encoded RF packet payload, padded
	Datr *string    `json:"-"`              // FSK datarate (unsigned in bit per second) || LoRa datarate identifier
	Fdev *uint      `json:"fdev,omitempty"` // FSK frequency deviation (unsigned integer, in Hz)
	Freq *float64   `json:"freq,omitempty"` // RX Central frequency in MHx (unsigned float, Hz precision)
	Imme *bool      `json:"imme,omitempty"` // Send packet immediately (will ignore tmst & time)
	Ipol *bool      `json:"ipol,omitempty"` // Lora modulation polarization inversion
	Lsnr *float64   `json:"lsnr,omitempty"` // LoRa SNR ratio in dB (signed float, 0.1 dB precision)
	Modu *string    `json:"modu,omitempty"` // Modulation identifier "LORA" or "FSK"
	Ncrc *bool      `json:"ncrc,omitempty"` // If true, disable the CRC of the physical layer (optional)
	Powe *uint      `json:"powe,omitempty"` // TX output power in dBm (unsigned integer, dBm precision)
	Prea *uint      `json:"prea,omitempty"` // RF preamble size (unsigned integer)
	Rfch *uint      `json:"rfch,omitempty"` // Concentrator "RF chain" used for RX (unsigned integer)
	Rssi *int       `json:"rssi,omitempty"` // RSSI in dBm (signed integer, 1 dB precision)
	Size *uint      `json:"size,omitempty"` // RF packet payload size in bytes (unsigned integer)
	Stat *int       `json:"stat,omitempty"` // CRC status: 1 - OK, -1 = fail, 0 = no CRC
	Time *time.Time `json:"-"`              // UTC time of pkt RX, us precision, ISO 8601 'compact' format
	Tmst *uint      `json:"tmst,omitempty"` // Internal timestamp of "RX finished" event (32b unsigned)
}

type metadata Metadata

// MarshalJSON implements the json.Marshal interface
func (m Metadata) MarshalJSON() ([]byte, error) {
	var d *semtech.Datrparser
	var t *semtech.Timeparser

	if m.Datr != nil {
		d = new(semtech.Datrparser)
		if m.Modu != nil && *m.Modu == "FSK" {
			*d = semtech.Datrparser{Kind: "uint", Value: *m.Datr}
		} else {
			*d = semtech.Datrparser{Kind: "string", Value: *m.Datr}
		}
	}

	if m.Time != nil {
		t = new(semtech.Timeparser)
		*t = semtech.Timeparser{Layout: time.RFC3339Nano, Value: m.Time}
	}

	return json.Marshal(metadataProxy{
		metadata: metadata(m),
		Datr:     d,
		Time:     t,
	})
}

// UnmarshalJSON implements the json.Unmarshaler interface
func (m *Metadata) UnmarshalJSON(raw []byte) error {
	return nil
}

// type metadataProxy is used to conveniently marshal and unmarshal Metadata structure.
//
// Datr field could be either string or uint depending on the Modu field.
// Time field could be parsed in a lot of different way depending of the time format.
// This proxy make sure that everything is marshaled and unmarshaled to the right thing and allow
// the Metadata struct to be user-friendly.
type metadataProxy struct {
	metadata
	Datr *semtech.Datrparser `json:"datr,omitempty"`
	Time *semtech.Timeparser `json:"time,omitempty"`
}

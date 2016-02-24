// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package refactor

import (
	"encoding"
	"fmt"
	"time"

	"github.com/brocaar/lorawan"
)

type Component interface {
	Register(reg Registration, an AckNacker) error
	HandleUp(p []byte, an AckNacker, upAdapter Adapter) error
	HandleDown(p []byte, an AckNacker, downAdapter Adapter) error
}

type AckNacker interface {
	Ack(p Packet) error
	Nack() error
}

type Adapter interface {
	Send(p Packet, r ...Recipient) ([]byte, error)
	GetRecipient(raw []byte) (Recipient, error)
	Next() ([]byte, AckNacker, error)
	NextRegistration() (Registration, AckNacker, error)
	//Join(r JoinRequest, r ...Recipient) (JoinResponse, error)
}

type Packet interface {
	encoding.BinaryMarshaler
	fmt.Stringer
}

type Addressable interface {
	DevEUI() (lorawan.EUI64, error)
}

type Registration interface {
	Recipient() Recipient
	AppEUI() (lorawan.EUI64, error)
	AppSKey() (lorawan.AES128Key, error)
	DevEUI() (lorawan.EUI64, error)
	NwkSKey() (lorawan.AES128Key, error)
}

type Recipient interface {
	encoding.BinaryMarshaler
}

type Metadata struct {
	Chan *uint      `json:"chan,omitempty"` // Concentrator "IF" channel used for RX (unsigned integer)
	Codr *string    `json:"codr,omitempty"` // LoRa ECC coding rate identifier
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

package core

import (
	"github.com/thethingsnetwork/ttn/lorawan"
	"time"
)

type Packet struct {
	Metadata Metadata
	Payload  lorawan.PHYPayload
}

type Recipient struct {
	Address interface{}
	Id      interface{}
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

type AckNacker interface {
	Ack(p Packet) error
	Nack(p Packet) error
}

type Component interface {
	Register(reg Registration) error
	HandleUp(p Packet, an AckNacker, upAdapter Adapter) error
	HandleDown(p Packet, an AckNacker, downAdapter Adapter) error
}

type Adapter interface {
	Send(p Packet, r ...Recipient) (Packet, error)
	Next() (Packet, AckNacker, error)
	NextRegistration() (Registration, AckNacker, error)
}

type Registration struct {
	DevAddr   lorawan.DevAddr
	Recipient Recipient
	Options   interface{}
}

type Router Component
type Broker Component
type Handler Component
type NetworkController Component

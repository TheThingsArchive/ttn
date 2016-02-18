// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package core

import (
	"time"

	"github.com/brocaar/lorawan"
)

type Packet struct {
	Metadata Metadata           // Metadata associated to packet. That object may change over requests
	Payload  lorawan.PHYPayload // The actual lorawan physical payload
}

type Recipient struct {
	Address interface{} // The address of the recipient. The type depends on the context
	Id      interface{} // An optional ID for the recipient. The type depends on the context
}

type Registration struct {
	DevAddr   lorawan.DevAddr // The device address which takes part in the registration
	Recipient Recipient       // The registration emitter to which it is bound
	Options   interface{}     // Options that vary from different type of registration
}

type Metadata struct {
	Chan  *uint      `json:"chan,omitempty"`     // Concentrator "IF" channel used for RX (unsigned integer)
	Codr  *string    `json:"codr,omitempty"`     // LoRa ECC coding rate identifier
	Datr  *string    `json:"-"`                  // FSK datarate (unsigned in bit per second) || LoRa datarate identifier
	Fdev  *uint      `json:"fdev,omitempty"`     // FSK frequency deviation (unsigned integer, in Hz)
	Freq  *float64   `json:"freq,omitempty"`     // RX Central frequency in MHx (unsigned float, Hz precision)
	Imme  *bool      `json:"imme,omitempty"`     // Send packet immediately (will ignore tmst & time)
	Ipol  *bool      `json:"ipol,omitempty"`     // Lora modulation polarization inversion
	Lsnr  *float64   `json:"lsnr,omitempty"`     // LoRa SNR ratio in dB (signed float, 0.1 dB precision)
	Modu  *string    `json:"modu,omitempty"`     // Modulation identifier "LORA" or "FSK"
	Ncrc  *bool      `json:"ncrc,omitempty"`     // If true, disable the CRC of the physical layer (optional)
	Powe  *uint      `json:"powe,omitempty"`     // TX output power in dBm (unsigned integer, dBm precision)
	Prea  *uint      `json:"prea,omitempty"`     // RF preamble size (unsigned integer)
	Rfch  *uint      `json:"rfch,omitempty"`     // Concentrator "RF chain" used for RX (unsigned integer)
	Rssi  *int       `json:"rssi,omitempty"`     // RSSI in dBm (signed integer, 1 dB precision)
	Size  *uint      `json:"size,omitempty"`     // RF packet payload size in bytes (unsigned integer)
	Stat  *int       `json:"stat,omitempty"`     // CRC status: 1 - OK, -1 = fail, 0 = no CRC
	Time  *time.Time `json:"-"`                  // UTC time of pkt RX, us precision, ISO 8601 'compact' format
	Tmst  *uint      `json:"tmst,omitempty"`     // Internal timestamp of "RX finished" event (32b unsigned)
	Group []Metadata `json:"metadata,omitempty"` // Gather metadata of several packets into one metadata structure
}

// AckNacker are mainly created by adapters as explicit callbacks for a given incoming request.
//
// The AckNacker encapsulates the logic which allows later component to answer to a recipient
// without even knowing about the protocol being used. This is only possible because all messages
// transmitted between component are relatively isomorph (to what one calls a Packet).
type AckNacker interface {
	// Ack acknowledges and terminates a connection by sending 0 or 1 packet as an answer
	// (depending on the component).
	//
	// Incidentally, that acknowledgement would serve as a downlink response for class A devices.
	Ack(p *Packet) error

	// Nack rejects and terminates a connection. So far, there is no way to give more information
	// about the reason that led to a rejection.
	Nack() error
}

type Component interface {
	// Register explicitely requires a component to create a new entry in its registry which links a
	// recipient to a device.
	Register(reg Registration, an AckNacker) error

	// HandleUp informs the component of an incoming uplink request.
	//
	// The component is thereby in charge of calling the given AckNacker accordingly to the
	// processing.
	HandleUp(p Packet, an AckNacker, upAdapter Adapter) error

	// HandleDown informs the component of a spontaneous downlink request.
	//
	// This should be mistaken with a typical downlink message as defined for Class A devices.
	// HandleDown should handle downlink request made without any uplink context (which basically
	// means we won't probably use it before a while ~> Class B and Class C, or probably only for
	// handlers).
	HandleDown(p Packet, an AckNacker, downAdapter Adapter) error
}

type Adapter interface {
	// Send forwards a given packet to one or more recipients. It returns only when the request has
	// been processed by the recipient and normally respond with a packet coming from the recipient.
	Send(p Packet, r ...Recipient) (Packet, error)

	// Next pulls a new packet received by the adapter. It blocks until a new packet is received.
	//
	// The adapter is in charge of creating a new AckNacker to reply to the recipient such that the
	// communication and the request are still transparent for the component handling it.
	Next() (Packet, AckNacker, error)

	// NextRegistration pulls a new registration request received by the adapter. It blocks until a
	// new registration demand is received. It follows a process similar to Next()
	NextRegistration() (Registration, AckNacker, error)
}

type Router Component
type Broker Component
type Handler Component
type NetworkController Component

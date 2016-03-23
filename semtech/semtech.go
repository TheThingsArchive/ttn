// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

// Package semtech provides useful methods and types to handle communications with a gateway.
//
// This package relies on the SemTech Protocol 1.2 accessible on github: https://github.com/TheThingsNetwork/packet_forwarder/blob/master/PROTOCOL.TXT
package semtech

import (
	"fmt"
	"time"

	"github.com/TheThingsNetwork/ttn/utils/pointer"
)

type DeviceAddress [4]byte

// RXPK represents an uplink json message format sent by the gateway
type RXPK struct {
	Chan *uint32    `full:"Channel" json:"chan,omitempty"`     // Concentrator "IF" channel used for RX (unsigned integer)
	Codr *string    `full:"CodingRate" json:"codr,omitempty"`  // LoRa ECC coding rate identifier
	Data *string    `full:"Data" json:"data,omitempty"`        // Base64 encoded RF packet payload, padded
	Datr *string    `full:"DataRate" json:"-"`                 // FSK datarate (unsigned in bit per second) || LoRa datarate identifier
	Freq *float32   `full:"Frequency" json:"freq,omitempty"`   // RX Central frequency in MHx (unsigned float, Hz precision)
	Lsnr *float32   `full:"Lsnr" json:"lsnr,omitempty"`        // LoRa SNR ratio in dB (signed float, 0.1 dB precision)
	Modu *string    `full:"Modulation" json:"modu,omitempty"`  // Modulation identifier "LORA" or "FSK"
	Rfch *uint32    `full:"RFChain" json:"rfch,omitempty"`     // Concentrator "RF chain" used for RX (unsigned integer)
	Rssi *int32     `full:"Rssi" json:"rssi,omitempty"`        // RSSI in dBm (signed integer, 1 dB precision)
	Size *uint32    `full:"PayloadSize" json:"size,omitempty"` // RF packet payload size in bytes (unsigned integer)
	Stat *int32     `full:"CRCStatus" json:"stat,omitempty"`   // CRC status: 1 - OK, -1 = fail, 0 = no CRC
	Time *time.Time `full:"Time" json:"-"`                     // UTC time of pkt RX, us precision, ISO 8601 'compact' format
	Tmst *uint32    `full:"Timestamp" json:"tmst,omitempty"`   // Internal timestamp of "RX finished" event (32b unsigned)
}

// TXPK represents a downlink j,omitemptyson message format received by the gateway.
// Most field are optional.
type TXPK struct {
	Codr *string    `full:"CodingRate" json:"codr,omitempty"`   // LoRa ECC coding rate identifier
	Data *string    `full:"Data" json:"data,omitempty"`         // Base64 encoded RF packet payload, padding optional
	Datr *string    `full:"DataRate" json:"-"`                  // LoRa datarate identifier (eg. SF12BW500) || FSK Datarate (unsigned, in bits per second)
	Fdev *uint32    `full:"FrequencyDev" json:"fdev,omitempty"` // FSK frequency deviation (unsigned integer, in Hz)
	Freq *float32   `full:"Frequency" json:"freq,omitempty"`    // TX central frequency in MHz (unsigned float, Hz precision)
	Imme *bool      `full:"Immediate" json:"imme,omitempty"`    // Send packet immediately (will ignore tmst & time)
	Ipol *bool      `full:"InvPolarity" json:"ipol,omitempty"`  // Lora modulation polarization inversion
	Modu *string    `full:"Modulation" json:"modu,omitempty"`   // Modulation identifier "LORA" or "FSK"
	Ncrc *bool      `full:"NoCRC" json:"ncrc,omitempty"`        // If true, disable the CRC of the physical layer (optional)
	Powe *uint32    `full:"Power" json:"powe,omitempty"`        // TX output power in dBm (unsigned integer, dBm precision)
	Prea *uint32    `full:"PreambleSize" json:"prea,omitempty"` // RF preamble size (unsigned integer)
	Rfch *uint32    `full:"RFChain" json:"rfch,omitempty"`      // Concentrator "RF chain" used for TX (unsigned integer)
	Size *uint32    `full:"PayloadSize" json:"size,omitempty"`  // RF packet payload size in bytes (unsigned integer)
	Time *time.Time `full:"Time" json:"-"`                      // Send packet at a certain time (GPS synchronization required)
	Tmst *uint32    `full:"Timestamp" json:"tmst,omitempty"`    // Send packet on a certain timestamp value (will ignore time)
}

// Stat represents a status json message format sent by the gateway
type Stat struct {
	Ackr *float32   `full:"Acknowledgements" json:"ackr,omitempty"` // Percentage of upstream datagrams that were acknowledged
	Alti *int32     `full:"Altitude" json:"alti,omitempty"`         // GPS altitude of the gateway in meter RX (integer)
	Dwnb *uint32    `full:"NbDownlink" json:"dwnb,omitempty"`       // Number of downlink datagrams received (unsigned integer)
	Lati *float32   `full:"Latitude" json:"lati,omitempty"`         // GPS latitude of the gateway in degree (float, N is +)
	Long *float32   `full:"Longitude" json:"long,omitempty"`        // GPS latitude of the gateway in dgree (float, E is +)
	Rxfw *uint32    `full:"RXForwarded" json:"rxfw,omitempty"`      // Number of radio packets forwarded (unsigned integer)
	Rxnb *uint32    `full:"RXReceived" json:"rxnb,omitempty"`       // Number of radio packets received (unsigned integer)
	Rxok *uint32    `full:"RXValid" json:"rxok,omitempty"`          // Number of radio packets received with a valid PHY CRC
	Time *time.Time `full:"Time" json:"-"`                          // UTC 'system' time of the gateway, ISO 8601 'expanded' format
	Txnb *uint32    `full:"TXEmitted" json:"txnb,omitempty"`        // Number of packets emitted (unsigned integer)
}

// Packet as seen by the gateway.
type Packet struct {
	Version    byte     // Protocol version, should always be 1 here
	Token      []byte   // Random number generated by the gateway on some request. 2-bytes long.
	Identifier byte     // Packet's command identifier
	GatewayId  []byte   // Source gateway's identifier (Only PULL_DATA and PUSH_DATA)
	Payload    *Payload // JSON payload transmitted if any, nil otherwise
}

func (p *Packet) String() string {
	if p == nil {
		return "nil"
	}
	header := fmt.Sprintf("Version:%x,Token:%v,Identifier:%x,GatewayId:%v", p.Version, p.Token, p.Identifier, p.GatewayId)
	if p.Payload == nil {
		return fmt.Sprintf("Packet{%s}", header)
	}

	var payload string

	if p.Payload.Stat != nil {
		payload = fmt.Sprintf("%s,%s", payload, pointer.DumpPStruct(*p.Payload.Stat, false))
	}

	if p.Payload.TXPK != nil {
		payload = fmt.Sprintf("%s,%s", payload, pointer.DumpPStruct(*p.Payload.TXPK, false))
	}

	var rxpk string
	for _, r := range p.Payload.RXPK {
		if rxpk == "" {
			rxpk = fmt.Sprintf("%s", pointer.DumpPStruct(r, false))
		} else {
			rxpk = fmt.Sprintf("%s,%s", rxpk, pointer.DumpPStruct(r, false))
		}
	}
	if rxpk != "" {
		payload = fmt.Sprintf("%s,RXPK:[%s]", payload, rxpk)
	}
	return fmt.Sprintf("Packet{%s,Payload:{%s}}", header, payload)
}

// Payload refers to the JSON payload sent by a gateway or a server.
type Payload struct {
	Raw  []byte `json:"-"`              // The raw unparsed response
	RXPK []RXPK `json:"rxpk,omitempty"` // A list of RXPK messages transmitted if any
	Stat *Stat  `json:"stat,omitempty"` // A Stat message transmitted if any
	TXPK *TXPK  `json:"txpk,omitempty"` // A TXPK message transmitted if any
}

// Available packet commands
const (
	PUSH_DATA byte = iota // Sent by the gateway for an uplink message with data
	PUSH_ACK              // Sent by the gateway's recipient in response to a PUSH_DATA
	PULL_DATA             // Sent periodically by the gateway to keep a connection open
	PULL_RESP             // Sent by the gateway's recipient to transmit back data to the Gateway
	PULL_ACK              // Sent by the gateway's recipient in response to PULL_DATA
)

// Protocol version in use
const VERSION = 0x01

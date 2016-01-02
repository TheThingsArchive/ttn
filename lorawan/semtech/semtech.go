// Copyright © 2015 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

// Package semtech provides useful methods and types to handle communications with a gateway.
//
// This package relies on the SemTech Protocol 1.2 accessible on github: https://github.com/TheThingsNetwork/packet_forwarder/blob/master/PROTOCOL.TXT
package semtech

import (
	"encoding/base64"
	"fmt"
	"time"
)

type DeviceAddress [4]byte

// RXPK represents an uplink json message format sent by the gateway
type RXPK struct {
	Chan    *uint          `json:"chan,omitempty"` // Concentrator "IF" channel used for RX (unsigned integer)
	Codr    *string        `json:"codr,omitempty"` // LoRa ECC coding rate identifier
	Data    *string        `json:"data,omitempty"` // Base64 encoded RF packet payload, padded
	Datr    *string        `json:"-"`              // FSK datarate (unsigned in bit per second) || LoRa datarate identifier
	Freq    *float64       `json:"freq,omitempty"` // RX Central frequency in MHx (unsigned float, Hz precision)
	Lsnr    *float64       `json:"lsnr,omitempty"` // LoRa SNR ratio in dB (signed float, 0.1 dB precision)
	Modu    *string        `json:"modu,omitempty"` // Modulation identifier "LORA" or "FSK"
	Rfch    *uint          `json:"rfch,omitempty"` // Concentrator "RF chain" used for RX (unsigned integer)
	Rssi    *int           `json:"rssi,omitempty"` // RSSI in dBm (signed integer, 1 dB precision)
	Size    *uint          `json:"size,omitempty"` // RF packet payload size in bytes (unsigned integer)
	Stat    *int           `json:"stat,omitempty"` // CRC status: 1 - OK, -1 = fail, 0 = no CRC
	Time    *time.Time     `json:"-"`              // UTC time of pkt RX, us precision, ISO 8601 'compact' format
	Tmst    *uint          `json:"tmst,omitempty"` // Internal timestamp of "RX finished" event (32b unsigned)
	devAddr *DeviceAddress // End-Device address, according to the Data. Memoized here.
}

// DevAddr returns the end-device address described in the payload
func (rxpk *RXPK) DevAddr() *DeviceAddress {
	if rxpk.devAddr != nil {
		return rxpk.devAddr
	}

	if rxpk.Data == nil {
		return nil
	}

	buf, err := base64.StdEncoding.DecodeString(*rxpk.Data)
	if err != nil || len(buf) < 5 {
		return nil
	}

	rxpk.devAddr = new(DeviceAddress)
	copy((*rxpk.devAddr)[:], buf[1:5]) // Device Address corresponds to the first 4 bytes of the Frame Header, after one byte of MAC_HEADER
	return rxpk.devAddr
}

// TXPK represents a downlink j,omitemptyson message format received by the gateway.
// Most field are optional.
type TXPK struct {
	Codr    *string        `json:"codr,omitempty"`  // LoRa ECC coding rate identifier
	Data    *string        `json:"data,omirtmepty"` // Base64 encoded RF packet payload, padding optional
	Datr    *string        `json:"-"`               // LoRa datarate identifier (eg. SF12BW500) || FSK Datarate (unsigned, in bits per second)
	Fdev    *uint          `json:"fdev,omitempty"`  // FSK frequency deviation (unsigned integer, in Hz)
	Freq    *float64       `json:"freq,omitempty"`  // TX central frequency in MHz (unsigned float, Hz precision)
	Imme    *bool          `json:"imme,omitempty"`  // Send packet immediately (will ignore tmst & time)
	Ipol    *bool          `json:"ipol,omitempty"`  // Lora modulation polarization inversion
	Modu    *string        `json:"modu,omitempty"`  // Modulation identifier "LORA" or "FSK"
	Ncrc    *bool          `json:"ncrc,omitempty"`  // If true, disable the CRC of the physical layer (optional)
	Powe    *uint          `json:"powe,omitempty"`  // TX output power in dBm (unsigned integer, dBm precision)
	Prea    *uint          `json:"prea,omitempty"`  // RF preamble size (unsigned integer)
	Rfch    *uint          `json:"rfch,omitempty"`  // Concentrator "RF chain" used for TX (unsigned integer)
	Size    *uint          `json:"size,omitempty"`  // RF packet payload size in bytes (unsigned integer)
	Time    *time.Time     `json:"-"`               // Send packet at a certain time (GPS synchronization required)
	Tmst    *uint          `json:"tmst,omitempty"`  // Send packet on a certain timestamp value (will ignore time)
	devAddr *DeviceAddress // End-Device address, according to the Data. Memoized here.
}

// DevAddr returns the end-device address described in the payload
func (txpk *TXPK) DevAddr() *DeviceAddress {
	if txpk.devAddr != nil {
		return txpk.devAddr
	}

	if txpk.Data == nil {
		return nil
	}

	buf, err := base64.StdEncoding.DecodeString(*txpk.Data)
	if err != nil || len(buf) < 5 {
		return nil
	}

	txpk.devAddr = new(DeviceAddress)
	copy((*txpk.devAddr)[:], buf[1:5]) // Device Address corresponds to the first 4 bytes of the Frame Header, after one byte of MAC_HEADER
	return txpk.devAddr
}

// Stat represents a status json message format sent by the gateway
type Stat struct {
	Ackr *float64   `json:"ackr,omitempty"` // Percentage of upstream datagrams that were acknowledged
	Alti *int       `json:"alti,omitempty"` // GPS altitude of the gateway in meter RX (integer)
	Dwnb *uint      `json:"dwnb,omitempty"` // Number of downlink datagrams received (unsigned integer)
	Lati *float64   `json:"lati,omitempty"` // GPS latitude of the gateway in degree (float, N is +)
	Long *float64   `json:"long,omitempty"` // GPS latitude of the gateway in dgree (float, E is +)
	Rxfw *uint      `json:"rxfw,omitempty"` // Number of radio packets forwarded (unsigned integer)
	Rxnb *uint      `json:"rxnb,omitempty"` // Number of radio packets received (unsigned integer)
	Rxok *uint      `json:"rxok,omitempty"` // Number of radio packets received with a valid PHY CRC
	Time *time.Time `json:"-"`              // UTC 'system' time of the gateway, ISO 8601 'expanded' format
	Txnb *uint      `json:"txnb,omitempty"` // Number of packets emitted (unsigned integer)
}

// Packet as seen by the gateway.
type Packet struct {
	Version    byte     // Protocol version, should always be 1 here
	Token      []byte   // Random number generated by the gateway on some request. 2-bytes long.
	Identifier byte     // Packet's command identifier
	GatewayId  []byte   // Source gateway's identifier (Only PULL_DATA and PUSH_DATA)
	Payload    *Payload // JSON payload transmitted if any, nil otherwise
}

// Payload refers to the JSON payload sent by a gateway or a server.
type Payload struct {
	Raw  []byte `json:"-"`              // The raw unparsed response
	RXPK []RXPK `json:"rxpk,omitempty"` // A list of RXPK messages transmitted if any
	Stat *Stat  `json:"stat,omitempty"` // A Stat message transmitted if any
	TXPK *TXPK  `json:"txpk,omitempty"` // A TXPK message transmitted if any
}

// UniformDevAddr tries to extract a device address from the different part of a payload. If the
// payload is composed of messages coming from several end-device, the method will fail.
func (p Payload) UniformDevAddr() (*DeviceAddress, error) {
	var devAddr *DeviceAddress

	// Determine the devAddress associated to that payload
	if p.RXPK == nil || len(p.RXPK) == 0 {
		if p.TXPK == nil {
			return nil, fmt.Errorf("Unable to determine device address. No RXPK neither TXPK messages")
		}
		if devAddr = p.TXPK.DevAddr(); devAddr == nil {
			return nil, fmt.Errorf("Unable to determine device address from TXPK")
		}

	} else {
		// We check them all to be sure, but all RXPK should refer to the same End-Device
		for _, rxpk := range p.RXPK {
			addr := rxpk.DevAddr()
			if addr == nil {
				return nil, fmt.Errorf("Unable to determine uniform address of given payload")
			}

			if devAddr != nil && *devAddr != *addr {
				return nil, fmt.Errorf("Payload is composed of messages from several end-devices")
			}
			devAddr = addr
		}
	}
	return devAddr, nil
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

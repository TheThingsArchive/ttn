// Copyright © 2016 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package core

import (
	"encoding"
	"fmt"

	"github.com/brocaar/lorawan"
)

// Packet represents the core base of a Packet
type Packet interface {
	DevEUI() lorawan.EUI64
	encoding.BinaryMarshaler
	fmt.Stringer
}

// RPacket represents packets manipulated by the router that hold Data
type RPacket interface {
	Packet
	GatewayID() []byte
	Metadata() Metadata
	Payload() lorawan.PHYPayload
}

// SPacket represents packets manipulated by the router that hold Stats
type SPacket interface {
	Packet
	GatewayID() []byte
	Metadata() Metadata
}

// BPacket represents packets manipulated by the broker that hold Data
type BPacket interface {
	Packet
	Commands() []lorawan.MACCommand
	ComputeFCnt(wholeCnt uint32) error
	FCnt() uint32
	Metadata() Metadata
	Payload() lorawan.PHYPayload
	ValidateMIC(key lorawan.AES128Key) (bool, error)
}

// HPacket represents packets manipulated by the handler that hold Data
type HPacket interface {
	Packet
	AppEUI() lorawan.EUI64
	FCnt() uint32
	Payload(appSKey lorawan.AES128Key) ([]byte, error) // Unencrypted FRMPayload
	Metadata() Metadata                                // TTL on down, DutyCycle + Rssi on Up
}

// APacket represents packets sent towards an application
type APacket interface {
	Packet
	AppEUI() lorawan.EUI64
	Payload() []byte
	Metadata() []Metadata
}

// JPacket represents join request packets
type JPacket interface {
	Packet
	AppEUI() lorawan.EUI64
	DevNonce() [2]byte
	Metadata() Metadata // Rssi + DutyCycle
}

// CPacket represents join accept (confirmation) packets
type CPacket interface {
	Packet
	AppEUI() lorawan.EUI64
	Payload() []byte
	NwkSKey() lorawan.AES128Key
}
